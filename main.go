package main

import (
	"database/sql"
	"log"
	"net/http"
	"socket"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

type dbHandler func(db *sql.DB, c echo.Context) error
type socketHandler func(h *socket.Hub, c echo.Context) error

func newDBHandler(h dbHandler, db *sql.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		return h(db, c)
	}
}
func newSocketHandler(h socketHandler, hub *socket.Hub) echo.HandlerFunc {
	return func(c echo.Context) error {
		return h(hub, c)
	}
}

func errorCheck(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	e := echo.New()
	hub := socket.NewHub()
	handlers := &Handler{}
	db, err := InitDB("./database/2facies.db")
	errorCheck(err)

	go hub.Run()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "")
	})
	e.GET("/api/client/version", handlers.clientVersion)

	e.GET("/user/me", newDBHandler(handlers.privateInfo, db), IsLoggedIn)
	e.GET("/user/:id", newDBHandler(handlers.publicInfo, db))

	e.POST("/user/login", newDBHandler(handlers.userLogin, db))
	e.POST("/user/register", newDBHandler(handlers.userRegister, db))
	e.GET("/user/logout", newDBHandler(handlers.userLogout, db), IsLoggedIn)

	e.GET("/list/room/:limits", newSocketHandler(handlers.roomList, hub))
	e.GET("/list/room/connectable", newSocketHandler(handlers.connectableRoom, hub))
	e.GET("/ws", newSocketHandler(handlers.ws, hub))

	e.Logger.Fatal(e.Start(":8000"))
}
