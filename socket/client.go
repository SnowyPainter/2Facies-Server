package socket

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
	Hub  *Hub
	Room string
	Conn *websocket.Conn
	Send chan []byte
}

func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error { c.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
				c.Send <- packet.SockError(packet.UnexceptError)
			}
			break
		}

		header, roomId := packet.GetHeader(message)
		switch strconv.Itoa(header) {
		case packet.BroadcastHeader:
			if roomId != "" {
				pack := packet.BindPrivatePacket(message)
				roomBc := NewRoomBroadcast(pack.RoomId, pack.UserId, pack.Body, c, packet.TypeTextBroadcast)
				c.Hub.Broadcast <- roomBc
			}
		case packet.BroadcastAudioHeader:
			if roomId != "" {
				pack := packet.BindPrivatePacket(message)
				roomBc := NewRoomBroadcast(pack.RoomId, pack.UserId, pack.Body, c, packet.TypeAudioBroadcast)
				c.Hub.Broadcast <- roomBc
			}
		case packet.CreateHeader:
			pack := packet.BindCreateRoomPacket(message)
			if pack != nil {
				r := NewRoom(strconv.Itoa(len(c.Hub.Rooms)), pack.Title, pack.MaxParticipants)
				r.Clients[c] = true
				c.Hub.CreateRoom <- r
				c.Send <- packet.SockPacket(packet.CreateHeader, []byte(r.Id))
			} else {
				log.Println("format error ", string(pack.MaxParticipants))
				c.Send <- packet.SockError(packet.FormatError)
			}

		case packet.JoinHeader:
			rId := string(roomId)
			c.Room = rId
			c.Hub.Join <- c
		case packet.LeaveHeader:
			c.Room = string(roomId)
			c.Hub.Leave <- c
		case packet.ParticipantsHeader:
			c.Hub.Participants <- string(roomId)
		}
	}
}

func (c *Client) WritePump() { //loop for hub.broadcast's send channel inputww
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.Send: //if c.send contains nothing, then wait (inf loop principal)
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
