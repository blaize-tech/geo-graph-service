package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"gopkg.in/mgo.v2"
	"log"
	"time"
)

const ClientMaxEventBufferSize int = 1
const PingIntervalSec int = 10


type Client struct {
	S *Server
	Conn *websocket.Conn
	Database *mgo.Database
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

func (c * Client) readPing() {
	go func() {
		for {
			c.Conn.SetReadDeadline(time.Now().Add(time.Duration( PingIntervalSec * time.Now().Second())))
			_, _, err := c.Conn.ReadMessage()
			if err != nil {
				c.Conn.Close()
				c.S.unregister <- c
				return
			}
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