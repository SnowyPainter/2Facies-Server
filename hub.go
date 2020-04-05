package main

import (
	"log"
	"packet"
	"strconv"
)

type Hub struct {
	clients      map[*Client]bool
	register     chan *Client
	unregister   chan *Client
	createRoom   chan *Room
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
	room     string
	caster   *Client
	message  []byte
	dataType int
}

func newHub() *Hub {
	return &Hub{
		broadcast:    make(chan *RoomBroadcast),
		register:     make(chan *Client),
		unregister:   make(chan *Client),
		createRoom:   make(chan *Room),
		join:         make(chan *Client),
		leave:        make(chan *Client),
		participants: make(chan string),
		clients:      make(map[*Client]bool),
		rooms:        make(map[string]*Room),
	}
}
func newRoom(roomId string, title string, maxClients int) *Room {
	//hexTitle, _ := utility.RandomHex(5)
	return &Room{
		id:         roomId,
		title:      title,
		clients:    make(map[*Client]bool),
		maxClients: maxClients,
	}
}
func newRoomBroadcast(roomId string, msg []byte, broadcaster *Client, typeOfData int) *RoomBroadcast {
	return &RoomBroadcast{
		room:     roomId,
		message:  msg,
		caster:   broadcaster,
		dataType: typeOfData,
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
					client.send <- packet.SockError(packet.RoomFull)
				} else if !room.clients[client] {
					h.rooms[roomId].clients[client] = true
				}
			} else {
				client.send <- packet.SockError(packet.NotFound)
			}
		case room := <-h.createRoom:
			if _, ok := h.rooms[room.id]; ok { // exist room
				for c, _ := range room.clients {
					c.send <- packet.SockError(packet.ExistRoom)
				}
			} else {
				h.rooms[room.id] = room
			}
		case client := <-h.leave:
			roomId := client.room
			log.Println("Leave")
			if _, ok := h.rooms[roomId]; ok {
				if _, ok := h.rooms[roomId].clients[client]; ok {
					delete(h.rooms[roomId].clients, client)
					if len(h.rooms[roomId].clients) < 1 {
						delete(h.rooms, roomId)
					}
				}
			} else {
				client.send <- packet.SockError(packet.NotFound)
			}

		case bc := <-h.broadcast:
			if bc.dataType == packet.TypeAudioBroadcast {
				for client := range h.rooms[bc.room].clients {
					select {
					case client.send <- packet.SockPacket(packet.BroadcastAudioHeader, bc.message):
					default:
						close(client.send)
						delete(h.clients, client)
					}
				}
			} else if bc.dataType == packet.TypeTextBroadcast {
				for client := range h.rooms[bc.room].clients {
					if client == bc.caster {
						continue
					}
					select {
					case client.send <- packet.SockPacket(packet.BroadcastHeader, bc.message):
					default:
						close(client.send)
						delete(h.clients, client)
					}
				}
			} else {
				bc.caster.send <- packet.SockError(packet.IncorrectDataType)
			}

		case r := <-h.participants:
			//room still alive -> client in there
			//log.Println("Participants client [r :", r, "]")
			if room, ok := h.rooms[r]; ok {
				count := []byte(strconv.Itoa(len(room.clients)))
				for client := range room.clients {
					select {
					case client.send <- packet.SockPacket(packet.ParticipantsHeader, count):
					default:
						close(client.send)
						delete(h.clients, client)
					}
				}
			}
		}
	}
}
