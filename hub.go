package main

import "log"

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
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

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
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

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register: //wait for register request for clients
			h.clients[client] = true

		case client := <-h.unregister: //close connection
			log.Println("unregister")
			if _, ok := h.clients[client]; ok { //check the closing conn had connected
				delete(h.rooms[client.room].clients, client)
				delete(h.clients, client)
				close(client.send)
			}
		case client := <-h.join:
			roomId := client.room
			if _, ok := h.rooms[roomId]; ok {
				log.Println(len(h.rooms[roomId].clients))
				for c, b := range h.rooms[roomId].clients {
					log.Println(b)
					select {
					case c.send <- []byte("New hello!!"):
					default:
						close(c.send)
						delete(h.rooms[roomId].clients, c)
					}
				}
			} else {
				log.Println(ok)
			}
		case client := <-h.leave:
			roomId := client.room
			if _, ok := h.rooms[roomId].clients[client]; ok {
				delete(h.rooms[roomId].clients, client)
				log.Println("Leave room left:", len(h.rooms[roomId].clients))
			}
		case message := <-h.broadcast: //add channel value to send all msgs
			//space @, middle one is room name
			//axess room h.rooms[rid].clients
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}
