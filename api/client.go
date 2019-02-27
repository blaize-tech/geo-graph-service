package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"gopkg.in/mgo.v2"
	"log"
)

const ClientMaxEventBufferSize int = 1


type Client struct {
	S *Server
	Conn *websocket.Conn
	Database *mgo.Database
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
		if len(c.PendingEvents.Items) >= ClientMaxEventBufferSize {
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
			fmt.Println("Unregistered")
			c.Conn.Close()
		}()
		if err := c.sendDB(); err == nil {
			c.writeEvents()
		}
	}()
}

func (c *Client) sendDB() error {
	//var wg sync.WaitGroup
	//Database.wg.Wait()
	//Database. Wg2.Add(1)
	//defer Database.Wg2.Done()

	rsTrustlines, _ := geItems()

	for _, node := range rsTrustlines {
		bsnode, err := json.Marshal(node)
		if err !=nil {
			log.Println("Error:",err)
		}
		if err := c.write(bsnode); err != nil {
			log.Println("Error:",err)
			return err
		}
	}

	//nodes := [][]byte{ []byte(bsTrustlines), []byte(bsPayments) } //TODO: replace with db request
	//nodes := bsTrustlines //TODO: replace with db request
	//log.Println(nodes)
	//if err := c.writeUint64(uint64(len(nodes))); err != nil {
	//	return err
	//}
	//for _, node := range nodes {
	//	if err := c.write(node); err != nil {
	//		return err
	//	}
	//}
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