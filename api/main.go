package main

import (
	"fmt"
	"net/http"

	"github.com/geo-graph-service/api/models"
	"github.com/geo-graph-service/api/models/item/db"

	"github.com/gorilla/websocket"
)

func main() {
	db.InitDB()
	s := models.NewServer()
	go s.run()

	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	http.HandleFunc("/api/v1/nodes", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			models.CreateNode(s, w, r)
		case "DELETE":
			models.DeleteNode(s, w, r)
		default:
			http.Error(w, "Invalid request method.", 405)
		}
	})

	http.HandleFunc("/api/v1/trustlines", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			models.CreateTrustline(s, w, r)
		} else if r.Method == "DELETE" {
			models.DeleteTrustline(s, w, r)
		} else {
			http.Error(w, "Invalid request method.", 405)
		}
	})

	http.HandleFunc("/api/v1/payments", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			models.PostPaymentItem(s, w, r)
		} else {
			http.Error(w, "Invalid request method.", 405)
		}
	})

	http.HandleFunc("/api/v1/clear", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			models.DeleteAll(w, r)
		} else {
			http.Error(w, "Invalid request method.", 405)
		}

	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		client := models.Client{Conn: conn,
			S: s}

		client.readPing()
		err = client.sendDB()
		if err == nil {
			s.register <- &client
		}
	})

	err := http.ListenAndServe(":3030", nil)
	if err != nil {
		fmt.Println("ListenAndServe: ", err)
	}
}
