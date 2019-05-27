package models

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"geo-graph-service/api/models/item"

	"github.com/gorilla/websocket"
	"gopkg.in/mgo.v2"
)

const ClientMaxEventBufferSize int = 1
const PingIntervalSec int = 10

type Client struct {
	S             *Server
	Conn          *websocket.Conn
	Database      *mgo.Database
	PendingEvents ConcurrentSlice
}

func (c *Client) pushEvent(event []byte) {
	c.PendingEvents.append(event)
	if len(c.PendingEvents.Items) >= ClientMaxEventBufferSize {
		events := c.PendingEvents.dumpSlice()
		if err := c.writeUint64(uint64(len(events))); err != nil {
			c.Conn.Close()
			return
		}
		for _, event := range events {
			if err := c.write(event); err != nil {
				c.Conn.Close()
				return
			}
		}
	}
}

func (c *Client) ReadPing() {
	go func() {
		for {
			c.Conn.SetReadDeadline(time.Now().Add(time.Duration(PingIntervalSec) * time.Second))
			_, _, err := c.Conn.ReadMessage()
			if err != nil {
				c.Conn.Close()
				c.S.unregister <- c
				return
			}
		}
	}()
}

func (c *Client) SendDB() error {
	rsNodes, rsTrustlines, _ := GetItems()
	for _, node := range rsNodes {
		nodeConv := item.Trustline{Source: node.Hash, Destination: node.Hash, Time: time.Now()}

		bsnode, err := json.Marshal(nodeConv)
		if err != nil {
			log.Println("Error:", err)
		}
		if err := c.write(bsnode); err != nil {
			log.Println("Error:", err)
			return err
		}
	}

	for _, trustline := range rsTrustlines {
		bsnode, err := json.Marshal(trustline)

		if err != nil {
			log.Println("Error:", err)
		}
		if err := c.write(bsnode); err != nil {
			log.Println("Error:", err)
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
