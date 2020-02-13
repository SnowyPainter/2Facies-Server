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
	handlers := &handler{}
	db, err := InitDB("./database/2facies.db")
	errorCheck(err)
	Database = db

	AddUser(db, "a", "password", "name", 13, "email address")

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})
	e.GET("/user/me", handlers.privateInfo, IsLoggedIn)
	e.POST("/user/login", handlers.userLogin)
	e.POST("/user/register", handlers.userRegister)

	e.Logger.Fatal(e.Start(":8000"))
}
