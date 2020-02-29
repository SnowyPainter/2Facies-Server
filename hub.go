package main

import (
	"log"
	"strconv"

)

const (
	ConnectionError = 101
	UnexceptError   = 102
	RoomJoinError   = 301
	RoomLeaveError  = 302
)

type Hub struct {
	clients    map[*Client]bool
	errors     chan *ErrorClient
	broadcast  chan *RoomBroadcast
	register   chan *Client
	unregister chan *Client
	join       chan *Client
	leave      chan *Client
	rooms      map[string]*Room
}

type Room struct {
	id      string
	clients map[*Client]bool
}
type RoomBroadcast struct {
	room    string
	message []byte
}
type ErrorClient struct {
	client *Client
	code   int
}

func newHub() *Hub {
	return &Hub{
		errors:     make(chan *ErrorClient),
		broadcast:  make(chan *RoomBroadcast),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		join:       make(chan *Client),
		leave:      make(chan *Client),
		clients:    make(map[*Client]bool),
		rooms:      make(map[string]*Room),
	}
}
func newRoom(roomId string) *Room {
	return &Room{
		id:      roomId,
		clients: make(map[*Client]bool),
	}
}
func newRoomBroadcast(roomId string, msg []byte) *RoomBroadcast {
	return &RoomBroadcast{
		room:    roomId,
		message: msg,
	}
}
func newError(ecode int, c *Client) *ErrorClient {
	return &ErrorClient{
		code:   ecode,
		client: c,
	}
}

func (h *Hub) run() {
	for {
		select {
		case code := <-h.errors:
			code.client.send <- []byte("error@" + strconv.Itoa(code.code))
		case client := <-h.register: //wait for register request for clients
			h.clients[client] = true

		case client := <-h.unregister: //close connection
			if _, ok := h.clients[client]; ok { //check the closing conn had connected
				if r, ok := h.rooms[client.room]; ok && r.clients[client] {
					delete(h.rooms[client.room].clients, client)
					if r, ok := h.rooms[client.room]; ok && len(r.clients) < 1 {
						delete(h.rooms, client.room)
					}
				}
				delete(h.clients, client)
				close(client.send)
			}
		case client := <-h.join:
			roomId := client.room
			if _, ok := h.rooms[roomId]; ok { // exist room
				h.rooms[roomId].clients[client] = true
			} else { //create new room
				log.Println("create new room")
				room := newRoom(roomId)
				room.clients[client] = true
				h.rooms[roomId] = room
			}
		case client := <-h.leave:
			roomId := client.room
			if _, ok := h.rooms[roomId].clients[client]; ok {
				delete(h.rooms[roomId].clients, client)
				if len(h.rooms[roomId].clients) < 1 {
					delete(h.rooms, roomId)
				}
			}
		case bc := <-h.broadcast: //broadcast to all clients in room
			//space @, middle one is room name
			//axess room h.rooms[rid].clients
			for client := range h.rooms[bc.room].clients {
				select {
				case client.send <- append([]byte("message@"), bc.message...):
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}
