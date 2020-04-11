package socket

import (
	"log"
	"packet"
	"strconv"
)

type Hub struct {
	Clients      map[*Client]bool
	Register     chan *Client
	Unregister   chan *Client
	CreateRoom   chan *Room
	Join         chan *Client
	Leave        chan *Client
	Broadcast    chan *RoomBroadcast
	Participants chan string
	Rooms        map[string]*Room
}

type Room struct {
	Id         string
	Title      string
	Clients    map[*Client]bool
	MaxClients int
}
type RoomBroadcast struct {
	Room     string
	User     string
	Caster   *Client
	Message  []byte
	DataType int
}

func NewHub() *Hub {
	return &Hub{
		Broadcast:    make(chan *RoomBroadcast),
		Register:     make(chan *Client),
		Unregister:   make(chan *Client),
		CreateRoom:   make(chan *Room),
		Join:         make(chan *Client),
		Leave:        make(chan *Client),
		Participants: make(chan string),
		Clients:      make(map[*Client]bool),
		Rooms:        make(map[string]*Room),
	}
}
func NewRoom(roomId string, title string, maxClients int) *Room {
	return &Room{
		Id:         roomId,
		Title:      title,
		Clients:    make(map[*Client]bool),
		MaxClients: maxClients,
	}
}
func NewRoomBroadcast(roomId string, userId string, msg []byte, broadcaster *Client, typeOfData int) *RoomBroadcast {
	return &RoomBroadcast{
		Room:     roomId,
		Message:  msg,
		Caster:   broadcaster,
		DataType: typeOfData,
		User:     userId,
	}
}
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register: //wait for register request for clients
			h.Clients[client] = true

		case client := <-h.Unregister: //close connection
			if _, ok := h.Clients[client]; ok { //check the closing conn had connected
				if r, ok := h.Rooms[client.Room]; ok && r.Clients[client] {
					delete(h.Rooms[client.Room].Clients, client)
					if r, ok := h.Rooms[client.Room]; ok && len(r.Clients) < 1 {
						delete(h.Rooms, client.Room)
					}
				}
				delete(h.Clients, client)
				close(client.Send)
			}
		case client := <-h.Join:
			roomId := client.Room
			if room, ok := h.Rooms[roomId]; ok { // exist room

				if len(room.Clients) >= room.MaxClients {
					client.Send <- packet.SockError(packet.RoomFull)
				} else if !room.Clients[client] {
					h.Rooms[roomId].Clients[client] = true
				}
			} else {
				client.Send <- packet.SockError(packet.NotFound)
			}
		case room := <-h.CreateRoom:
			if _, ok := h.Rooms[room.Id]; ok { // exist room
				for c, _ := range room.Clients {
					c.Send <- packet.SockError(packet.ExistRoom)
				}
			} else {
				h.Rooms[room.Id] = room
			}
		case client := <-h.Leave:
			roomId := client.Room
			log.Println("Leave")
			if _, ok := h.Rooms[roomId]; ok {
				if _, ok := h.Rooms[roomId].Clients[client]; ok {
					delete(h.Rooms[roomId].Clients, client)
					if len(h.Rooms[roomId].Clients) < 1 {
						delete(h.Rooms, roomId)
					}
				}
			} else {
				client.Send <- packet.SockError(packet.NotFound)
			}

		case bc := <-h.Broadcast:
			if bc.DataType == packet.TypeAudioBroadcast {
				for client := range h.Rooms[bc.Room].Clients {
					select {
					case client.Send <- packet.SockIdentifyPacket(packet.BroadcastAudioHeader, bc.User, bc.Message):
					default:
						close(client.Send)
						delete(h.Clients, client)
					}
				}
			} else if bc.DataType == packet.TypeTextBroadcast {
				for client := range h.Rooms[bc.Room].Clients {
					if client == bc.Caster {
						continue
					} // /* */ means for testing
					select {
					case client.Send <- packet.SockIdentifyPacket(packet.BroadcastHeader, bc.User, bc.Message):
					default:
						close(client.Send)
						delete(h.Clients, client)
					}
				}
			} else {
				bc.Caster.Send <- packet.SockError(packet.IncorrectDataType)
			}

		case r := <-h.Participants:
			//room still alive -> client in there
			//log.Println("Participants client [r :", r, "]")
			if room, ok := h.Rooms[r]; ok {
				count := []byte(strconv.Itoa(len(room.Clients)))
				for client := range room.Clients {
					select {
					case client.Send <- packet.SockPacket(packet.ParticipantsHeader, count):
					default:
						close(client.Send)
						delete(h.Clients, client)
					}
				}
			}
		}
	}
}
