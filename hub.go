package main

import (
	"log"
	"strconv"
	"utility"
)

const (
	ConnectionError = 101
	UnexceptError   = 102
	RoomJoinError   = 301
	RoomLeaveError  = 302
	RoomFull        = 303
	ChatSend        = 401
	ChatConnect     = 402
	ChatRecv        = 403
	NotFound        = 501
)

type Hub struct {
	clients      map[*Client]bool
	register     chan *Client
	unregister   chan *Client
	join         chan *Client
	leave        chan *Client
	broadcast    chan *RoomBroadcast
	participants chan string
	rooms        map[string]*Room
}

type Room struct {
	id         string
	title      string
	clients    map[*Client]bool
	maxClients int
}
type RoomBroadcast struct {
	room    string
	caster  *Client
	message []byte
}

func newHub() *Hub {
	return &Hub{
		broadcast:    make(chan *RoomBroadcast),
		register:     make(chan *Client),
		unregister:   make(chan *Client),
		join:         make(chan *Client),
		leave:        make(chan *Client),
		participants: make(chan string),
		clients:      make(map[*Client]bool),
		rooms:        make(map[string]*Room),
	}
}
func newRoom(roomId string, maxClients int) *Room {
	hexTitle, _ := utility.RandomHex(5)
	return &Room{
		id:         roomId,
		title:      hexTitle,
		clients:    make(map[*Client]bool),
		maxClients: maxClients,
	}
}
func newRoomBroadcast(roomId string, msg []byte, broadcaster *Client) *RoomBroadcast {
	return &RoomBroadcast{
		room:    roomId,
		message: msg,
		caster:  broadcaster,
	}
}

func (h *Hub) run() {
	for {
		select {
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
			if room, ok := h.rooms[roomId]; ok { // exist room
				if len(room.clients) >= room.maxClients {
					log.Println("FULL ROOm")
					client.send <- []byte("error@" + strconv.Itoa(RoomFull))
				} else if !room.clients[client] {
					log.Println("clinet joined", roomId)
					h.rooms[roomId].clients[client] = true
				}
			} else {
				log.Println("room created", roomId)
				room := newRoom(roomId, 2)
				room.clients[client] = true
				h.rooms[room.id] = room
				//client.send <- []byte("error@" + strconv.Itoa(NotFound))
			}
		case client := <-h.leave:
			roomId := client.room
			if _, ok := h.rooms[roomId]; ok {
				if _, ok := h.rooms[roomId].clients[client]; ok {
					log.Println("client leave", roomId)
					delete(h.rooms[roomId].clients, client)
					if len(h.rooms[roomId].clients) < 1 {
						log.Println(roomId, "deleted")
						delete(h.rooms, roomId)
					}
				}
			} else {
				client.send <- []byte("error@" + strconv.Itoa(NotFound))
			}

		case bc := <-h.broadcast: //broadcast to all clients in room
			//space @, middle one is room name
			//axess room h.rooms[rid].clients
			for client := range h.rooms[bc.room].clients {
				if client == bc.caster {
					continue
				}

				select {
				case client.send <- append([]byte("message@"), bc.message...):
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		case r := <-h.participants:
			log.Println(r)
			//room still alive -> client in there
			if room, ok := h.rooms[r]; ok {
				p := []byte(strconv.Itoa(len(room.clients)))
				for client := range h.rooms[r].clients {
					/*if client == c {
						continue
					}*/
					select {
					case client.send <- append([]byte("participants@"), p...):
					default:
						close(client.send)
						delete(h.clients, client)
					}
				}
			}
		}
	}
}
