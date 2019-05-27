package main

import (
	"fmt"
	"net/http"

	"github.com/GeoServer/project/api/models"
	_ "github.com/GeoServer/project/api/models/item/db"

	"github.com/gorilla/websocket"
)

func main() {
	s := models.NewServer()
	go s.Run()

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

	http.HandleFunc("/api/v1/topology/range", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			models.TopologyRange(w, r)
		default:
			http.Error(w, "Invalid request method.", 405)
		}
	})

	http.HandleFunc("/api/v1/topology", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			models.Topology(w, r)
		// case "POST":
		// 	models.TopologyRange(w, r)
		default:
			http.Error(w, "Invalid request method.", 405)
		}
	})

	http.HandleFunc("/api/v1/trustlines", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			models.PostTrustline(s, w, r)
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
			models.DeleteAll(s, w, r)
		} else {
			http.Error(w, "Invalid request method.", 405)
		}

	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		client := models.Client{Conn: conn, S: s}

		client.ReadPing()
		err = client.SendDB()
		if err == nil {
			s.Register <- &client
		}
	})

	err := http.ListenAndServe(":3030", nil)
	if err != nil {
		fmt.Println("ListenAndServe: ", err)
	}
}
