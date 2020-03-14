package main

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo"
	"golang.org/x/crypto/bcrypt"
)

type handler struct{}
type LoginData struct {
	Id       string `json:"id" form:"id" query:"id"`
	Password string `json:"password" form:"password" query:"password"`
}
type RegisterData struct {
	Id       string `json:"id" form:"id" query:"id"`
	Password string `json:"password" form:"password" query:"password"`
	Name     string `json:"name" form:"name" query:"name"`
	Email    string `json:"email" form:"email" query:"email"`
	Age      int    `json:"age" form:"age" query:"age"`
}
type ResponseData struct {
	Result  string `json:"result" form:"result" query:"result"`
	Message string `json:"message" form:"message" query:"message"`
}
type RoomData struct {
	Id           string `json:"Id" form:"Id" query:"Id"`
	Title        string `json:"Title" form:"title" query:"title"`
	Participants int    `json:"Participants" form:"participants" query:"participants"`
	Max          int    `json:"Max" form:"max" query:"max"`
}

var (
	upgrader = websocket.Upgrader{}
)

func (room *Room) ToRoomData() RoomData {
	data := RoomData{
		Id:           room.id,
		Max:          room.maxClients,
		Title:        room.title,
		Participants: len(room.clients),
	}

	return data
}

func (h *handler) clientVersion(c echo.Context) error {
	return c.String(http.StatusOK, "v0.0.1")
}

//user interact
func (h *handler) userLogin(db *sql.DB, c echo.Context) error {
	data := new(LoginData)
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	c.Bind(data)
	//data.password to bcrypt byted password

	if err := Login(db, data.Id, data.Password); err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return c.JSON(http.StatusOK, map[string]string{
				"token": "", "succeed": "false", "message": "password is not correct",
			})
		} else if err == ErrAlreadyLogined {
			return c.JSON(http.StatusOK, map[string]string{
				"token": "", "succeed": "false", "message": "already logined",
			})
		}
		return c.JSON(http.StatusOK, map[string]string{
			"token": "", "succeed": "false", "message": "id is not exist",
		})
	}
	claims["id"] = data.Id
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

	t, err := token.SignedString([]byte("secret"))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]string{
		"token": t, "succeed": "true",
	})
}
func (h *handler) userRegister(db *sql.DB, c echo.Context) error {
	data := new(RegisterData)
	c.Bind(data)

	response := ResponseData{Result: "true", Message: "register succeed"}
	hash, err := bcrypt.GenerateFromPassword([]byte(data.Password), bcrypt.DefaultCost)
	if err != nil {
		response.Result = "false"
		response.Message = "bcrypt error"
	}
	err = AddUser(db, data.Id, string(hash), data.Name, data.Age, data.Email)
	if err != nil {
		response.Result = "false"
		response.Message = "database error"
	}

	return c.JSON(http.StatusOK, response)
}
func (h *handler) userLogout(db *sql.DB, c echo.Context) error {
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	id := claims["id"].(string)

	if err := Logout(db, id); err != nil {
		return err
	}

	return nil
}

//data get,set
func (h *handler) privateInfo(db *sql.DB, c echo.Context) error {
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	id := claims["id"].(string)

	data, err := GetUser(db, id)

	if err != nil {
		return c.JSON(http.StatusOK, ResponseData{Result: "false", Message: "No results"})
	}
	return c.JSON(http.StatusOK, data)
}
func (h *handler) publicInfo(db *sql.DB, c echo.Context) error {
	id := c.Param("id")
	data, err := GetUserPublic(db, id)
	if err != nil {
		return c.JSON(http.StatusOK, ResponseData{Result: "false", Message: "No results"})
	}
	return c.JSON(http.StatusOK, data)
}
func (h *handler) roomList(hub *Hub, c echo.Context) error {
	roomCount := len(hub.rooms)
	if roomCount < 1 {
		return c.JSON(http.StatusOK, make([]RoomData, 0))
	}
	limit, err := strconv.Atoi(c.Param("limits"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Limits must be Intager")
	}
	list := make([]RoomData, roomCount-1)
	i := 0
	for k, v := range hub.rooms {
		if i >= limit {
			break
		}
		r := RoomData{Id: k, Title: v.title, Participants: len(v.clients), Max: v.maxClients}
		list = append(list, r)
		i++
	}

	return c.JSON(http.StatusOK, list)
}
func (h *handler) connectableRoom(hub *Hub, c echo.Context) error {
	const searchLimit = 3
	roomPackets := make([]RoomData, searchLimit)
	searchCount := 0
	for _, room := range hub.rooms {
		log.Println("count", searchCount)
		if searchCount >= searchLimit {
			break
		}
		if len(room.clients) < room.maxClients {
			log.Println("toroomdata", room.ToRoomData().Id)
			roomPackets[searchCount] = room.ToRoomData()
		}

		searchCount++
	}
	roomPackets = roomPackets[:searchCount]
	return c.JSON(http.StatusOK, roomPackets)
}

//'Client' would send data then processed with readPump
//after readPump get the data, readPump handle message to use hub. events e.g. broadcast
//readPump -> PROCESS -> hub.broadcast or some events ...
func (h *handler) ws(hub *Hub, c echo.Context) error {
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}

	client := &Client{hub: hub, conn: ws, send: make(chan []byte, 256)}
	client.hub.register <- client //request register own pointer(static) hub
	go client.writePump()
	go client.readPump()
	return nil
}
