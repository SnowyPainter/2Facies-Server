package main

import (
	"bytes"
	"log"
	"time"
	"utility"

	"github.com/gorilla/websocket"

)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

type Client struct {
	hub  *Hub
	room string
	conn *websocket.Conn
	send chan []byte
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		header := string(bytes.Split(message, space)[0])
		switch header {
		case "broadcast":
			c.hub.broadcast <- utility.DeleteFirst(message, space) //resend all message
		case "join":
			rId := string(utility.DeleteFirst(message, space))
			c.room = rId
			if _, ok := c.hub.rooms[rId]; ok { // exist room
				log.Println("exist room", rId)
				c.hub.rooms[rId].clients[c] = true
			} else { //create new room
				log.Println("create new", rId)
				room := newRoom(rId)
				room.clients[c] = true
				c.hub.rooms[rId] = room
			}
			c.hub.join <- c
		case "leave":
			c.room = string(utility.DeleteFirst(message, space))
			c.hub.leave <- c
		}
	}
}

func (c *Client) writePump() { //loop for hub.broadcast's send channel input
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send: //if c.send contains nothing, then wait (inf loop principal)
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			log.Println(string(message))
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
