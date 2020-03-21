package main

import (
	"log"
	"strconv"
	"strings"
	"time"
	"utility"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 200000
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
				c.send <- []byte("error@" + strconv.Itoa(UnexceptError))
			}
			break
		}
		header, body := utility.SplitHeaderBody(message)
		switch string(header[0]) {
		case "broadcast":
			roomBc := newRoomBroadcast(string(header[1]), body, c, TypeTextBroadcast)
			c.hub.broadcast <- roomBc //resend all message
		case "broadcast-audio":
			roomBc := newRoomBroadcast(string(header[1]), body, c, TypeAudioBroadcast)
			c.hub.broadcast <- roomBc //resend all message
		case "create":
			data := strings.Split(string(body), " ")
			if val, err := strconv.Atoi(data[1]); err == nil {

				r := newRoom(strconv.Itoa(len(c.hub.rooms)), data[0], val)
				r.clients[c] = true
				c.hub.createRoom <- r
				c.send <- []byte("created@" + r.id)
			} else {
				log.Println("format error ", string(data[1]))
				c.send <- []byte("error@" + strconv.Itoa(FormatError))
			}

		case "join":
			rId := string(header[1])
			c.room = rId
			c.hub.join <- c
		case "leave":
			c.room = string(header[1])
			c.hub.leave <- c
		case "participants":
			c.hub.participants <- string(header[1])
		}
	}
}

func (c *Client) writePump() { //loop for hub.broadcast's send channel inputww
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
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
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
