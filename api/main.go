package main

import (
	"fmt"
	"net/http"
	"github.com/gorilla/websocket"
)

func main() {
	s := newServer()
	go s.run()
	//go s.sendTestData()

	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	http.HandleFunc("/api/v1/updates", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Lppppppp")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if r.Method == "GET" {
			getAllItems(w, r)
		} else {
			http.Error(w, "Invalid request method.", 405)
		}
	})

	http.HandleFunc("/api/v1/trustlines", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			postTrustlineItem(s, w, r)
		} else {
			http.Error(w, "Invalid request method.", 405)
		}

	})

	http.HandleFunc("/api/v1/payments", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			postPaymentItem(s, w, r)
		} else {
			http.Error(w, "Invalid request method.", 405)
		}

	})

	http.HandleFunc("/api/v1/clear", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			deleteItem(w, r)
		} else {
			http.Error(w, "Invalid request method.", 405)
		}

	})


	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println(err)
			return
		}
		client := Client { Conn: conn,
			S: s }
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