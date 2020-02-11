package main

import (
	"net/http"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

)
func newCookie(name string, value string, expires time.Time) *http.Cookie {
	c:= new(http.Cookie)
	c.Name = name 
	c.Value = value
	c.Expires = expires
	return c
}

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	h := &handler{}

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})
	e.GET("/user/info", h.privateInfo, IsLoggedIn)
	e.POST("/user/login", h.userLogin)
	e.POST("/user/register", h.userRegister)

	e.Logger.Fatal(e.Start(":8000"))
}
