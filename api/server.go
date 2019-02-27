package main

import (
	"strconv"
	"fmt"
)

const ServerMaxEventBufferSize int = 1


type Server struct {
	clients map[*Client]bool
	register chan *Client
	unregister chan *Client
	event chan []byte
	PendingEvents ConcurrentSlice //events waiting to be written to DB
}

func newServer() *Server {
	return &Server{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		event:      make(chan []byte),
	}
}

func (s *Server) pushEvent(event []byte) {
	s.event <- event
}

func (s *Server) run() {
	go s.writeEvents()
	for {
		select {
		case client := <-s.register:
			s.clients[client] = true
			fmt.Println("Connected")
		case client := <-s.unregister:
			if _, ok := s.clients[client]; ok {
				delete(s.clients, client)
				fmt.Println("Disconnected")
			}
		case event := <-s.event:
			s.PendingEvents.append(event)
			for client := range s.clients {
				client.pushEvent(event)
			}
		}
	}
}

//need to do this job in another thread
//in order to have possibility to write events to db after some time passed
func (s *Server) writeEvents() {
	for {
		if len(s.PendingEvents.Items) >= ServerMaxEventBufferSize {
			events := s.PendingEvents.dumpSlice()
			_ = events
			//lock & update DB
		}
	}
}

//for test
func (s *Server) sendTestData() {
	i := 0
	for {
		s.pushEvent([]byte(strconv.FormatUint(uint64(i), 10)))
		i += 1
		//time.Sleep(2 * time.Second)
		//fmt.Println(".")
	}
}