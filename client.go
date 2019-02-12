package main

import (
	"fmt"
	"encoding/binary"
	"github.com/gorilla/websocket"
)

const ClientMaxEventBufferSize int = 3


type Client struct {
	S *Server
	Conn *websocket.Conn
	//Database *DB

	Event chan []byte
	PendingEvents ConcurrentSlice
}


func (c *Client) listenEvents() {
	for {
		event := <- c.Event
		c.PendingEvents.append(event)
	}
}

func (c *Client) writeEvents() error {
	for {
		if len(c.PendingEvents.Items) > ClientMaxEventBufferSize {
			events := c.PendingEvents.dumpSlice()
			if err := c.writeUint64(uint64(len(events))); err != nil {
				return err
			}
			for _, event := range events {
				if err := c.write(event); err != nil {
					return err
				}
			}
		}
	}
}

func (c *Client) pushEvent(event []byte) {
	c.Event <- event
}

func (c *Client) run() {
	go c.listenEvents()
	go func() {
		defer func() {
			c.S.unregister <- c
			c.Conn.Close()
		}()
		if err := c.sendDB(); err == nil {
			c.writeEvents()
		}
	}()
}

func (c *Client) sendDB() error {
	//Database.Wg1.Wait()
	//Database.Wg2.Add(1)
	//defer Database.Wg2.Done()
	nodes := [][]byte{ []byte(`123`), []byte(`456`) } //TODO: replace with db request
	if err := c.writeUint64(uint64(len(nodes))); err != nil {
		return err
	}
	for _, node := range nodes {
		if err := c.write(node); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) write(bytes []byte) error {
	if err := c.Conn.WriteMessage(websocket.BinaryMessage, bytes); err != nil {
        fmt.Println(err)
        return err
	}
	return nil
}

func (c *Client) writeUint64(n uint64) error {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], n)
	return c.write(buf[:])
}