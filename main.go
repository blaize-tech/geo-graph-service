package main

import (
	"fmt"
	"net/http"
	"github.com/gorilla/websocket"
)


func main() {
	s := newServer()
	go s.run()
	go s.sendTestData()

	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println(err)
			return
		}
		client := Client { Conn: conn,
			Event: make(chan []byte),
			S: s }
		client.run()
		s.register <- &client
	})

	err := http.ListenAndServe(":1234", nil)
	if err != nil {
		fmt.Println("ListenAndServe: ", err)
	}
}