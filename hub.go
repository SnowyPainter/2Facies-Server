package main

import "log"

type Hub struct {
	clients    map[*Client]bool
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

func newHub() *Hub {
	return &Hub{
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

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register: //wait for register request for clients
			h.clients[client] = true

		case client := <-h.unregister: //close connection
			log.Println("unregister")
			if _, ok := h.clients[client]; ok { //check the closing conn had connected
				if r, ok := h.rooms[client.room]; ok && r.clients[client] {
					log.Println("unregister, del client in room")
					delete(h.rooms[client.room].clients, client)
					if r, ok := h.rooms[client.room]; ok && len(r.clients) < 1 {
						log.Println("unregister, del room")
						delete(h.rooms, client.room)
					}
				}
				delete(h.clients, client)
				close(client.send)
			}
		case client := <-h.join:
			roomId := client.room
			if _, ok := h.rooms[roomId]; !ok {
				log.Println("join the room is not exist [hub]")
			}
		case client := <-h.leave:
			roomId := client.room
			if _, ok := h.rooms[roomId].clients[client]; ok {
				delete(h.rooms[roomId].clients, client)
				if len(h.rooms[roomId].clients) < 1 {
					log.Println("delete room ")
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
