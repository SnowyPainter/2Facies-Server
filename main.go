package main

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	_ "github.com/mattn/go-sqlite3"

)

func newCookie(name string, value string, expires time.Time) *http.Cookie {
	c := new(http.Cookie)
	c.Name = name
	c.Value = value
	c.Expires = expires
	return c
}

func errorCheck(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

var Database *sql.DB

func main() {
	e := echo.New()
	hub := newHub()
	handlers := &handler{}
	db, err := InitDB("./database/2facies.db")
	errorCheck(err)
	Database = db

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "")
	})
	e.GET("/api/client/version", handlers.clientVersion)

	e.GET("/user/me", handlers.privateInfo, IsLoggedIn)
	e.GET("/user/:id", handlers.publicInfo)

	e.POST("/user/login", handlers.userLogin)
	e.POST("/user/register", handlers.userRegister)

	go hub.run()
	e.GET("/ws", func(c echo.Context) error {
		return handlers.ws(hub, c)
	})

	e.Logger.Fatal(e.Start(":8000"))
}
