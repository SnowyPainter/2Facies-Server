package main

import (
	"log"
	"packet"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 35000
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
				c.send <- []byte(packet.ErrorHeader + "@" + strconv.Itoa(packet.UnexceptError))
			}
			break
		}

		header, roomId := packet.GetHeader(message)
		switch strconv.Itoa(header) {
		case packet.BroadcastHeader:
			if roomId != "" {
				pack := packet.BindPrivatePacket(message)
				roomBc := newRoomBroadcast(pack.RoomId, pack.UserId, pack.Body, c, packet.TypeTextBroadcast)
				c.hub.broadcast <- roomBc
			}
		case packet.BroadcastAudioHeader:
			if roomId != "" {
				pack := packet.BindPrivatePacket(message)
				roomBc := newRoomBroadcast(pack.RoomId, pack.UserId, pack.Body, c, packet.TypeAudioBroadcast)
				c.hub.broadcast <- roomBc
			}
		case packet.CreateHeader:
			pack := packet.BindCreateRoomPacket(message)
			if pack != nil {
				r := newRoom(strconv.Itoa(len(c.hub.rooms)), pack.Title, pack.MaxParticipants)
				r.clients[c] = true
				c.hub.createRoom <- r
				c.send <- packet.SockPacket(packet.CreateHeader, []byte(r.id))
			} else {
				log.Println("format error ", string(pack.MaxParticipants))
				c.send <- packet.SockError(packet.FormatError)
			}

		case packet.JoinHeader:
			rId := string(roomId)
			c.room = rId
			c.hub.join <- c
		case packet.LeaveHeader:
			c.room = string(roomId)
			c.hub.leave <- c
		case packet.ParticipantsHeader:
			c.hub.participants <- string(roomId)
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
