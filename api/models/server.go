package models

import (
	"fmt"
)

const ServerMaxEventBufferSize int = 1

type Server struct {
	clients       map[*Client]bool
	register      chan *Client
	unregister    chan *Client
	event         chan []byte
	PendingEvents ConcurrentSlice //events waiting to be written to DB
}

func NewServer() *Server {
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
	for {
		select {
		case client := <-s.register:
			s.clients[client] = true
		case client := <-s.unregister:
			if _, ok := s.clients[client]; ok {
				delete(s.clients, client)
				for client := range s.clients {
					fmt.Println(client)
				}
			}
		case event := <-s.event:
			s.PendingEvents.append(event)
			for client := range s.clients {
				client.pushEvent(event)
			}
		}
	}
}
