package socket

import (
	"log"
	"packet"
	"strconv"
)

type Hub struct {
	Clients          map[*Client]bool
	Register         chan *Client
	Unregister       chan *Client
	CreateRoom       chan *Room
	Join             chan *Client
	Leave            chan *Client
	Broadcast        chan *RoomBroadcast
	Participants     chan string
	Rooms            map[string]*Room
	CreatedRoomCount int
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
		Broadcast:        make(chan *RoomBroadcast),
		Register:         make(chan *Client),
		Unregister:       make(chan *Client),
		CreateRoom:       make(chan *Room),
		Join:             make(chan *Client),
		Leave:            make(chan *Client),
		Participants:     make(chan string),
		Clients:          make(map[*Client]bool),
		Rooms:            make(map[string]*Room),
		CreatedRoomCount: 0,
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
					log.Println("JOIN ROOM FULL :", roomId)
					client.Send <- packet.SockError(packet.RoomFull)
				} else {
					log.Println("Join Room ", roomId)
					h.Rooms[roomId].Clients[client] = true
				}
			} else {
				log.Println("Join NOT FOUND", roomId)
				client.Send <- packet.SockError(packet.NotFound)
			}

		case client := <-h.Leave:
			roomId := client.Room
			if _, ok := h.Rooms[roomId]; ok {
				if _, ok := h.Rooms[roomId].Clients[client]; ok {
					log.Println("Room Leave")
					delete(h.Rooms[roomId].Clients, client)
					if len(h.Rooms[roomId].Clients) < 1 {
						log.Println("Room Deleted", roomId)
						delete(h.Rooms, roomId)
					}
				}
			} else {
				client.Send <- packet.SockError(packet.NotFound)
			}
		case room := <-h.CreateRoom:
			if _, ok := h.Rooms[room.Id]; ok { // exist room
				log.Println("Room Already Exist", room.Id)
				for c, _ := range room.Clients {
					c.Send <- packet.SockError(packet.ExistRoom)
				}
			} else {
				log.Println("Created Room", room.Id)
				h.Rooms[room.Id] = room
			}
		case bc := <-h.Broadcast:
			if _, ok := h.Rooms[bc.Room]; !ok {
				continue
			}
			if bc.DataType == packet.TypeAudioBroadcast {
				for client := range h.Rooms[bc.Room].Clients {
					if client == bc.Caster {
						continue
					}
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
					}
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
