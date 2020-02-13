package main

import (
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"

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

func (h *handler) userLogin(c echo.Context) error {
	data := new(LoginData)
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	c.Bind(data)

	u, err := GetUser(Database, data.Id)
	if err != nil {
		return c.JSON(http.StatusOK, map[string]string{
			"token": "", "succeed": "false", "message": "id is not exist",
		})
	} else if u.Password != data.Password {
		return c.JSON(http.StatusOK, map[string]string{
			"token": "", "succeed": "false", "message": "password is not correct",
		})
	}
	claims["id"] = data.Id
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

	t, err := token.SignedString([]byte("secret"))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]string{
		"token": t, "succeed": "true",
	})
}
func (h *handler) userRegister(c echo.Context) error {
	data := new(RegisterData)
	c.Bind(data)

	response := ResponseData{Result: "true", Message: "register succeed"}

	err := AddUser(Database, data.Id, data.Password, data.Name, data.Age, data.Email)
	errorCheck(err)

	return c.JSON(http.StatusOK, response)
}

func (h *handler) privateInfo(c echo.Context) error {
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	id := claims["id"].(string)

	data, err := GetUser(Database, id)

	if err != nil {
		return c.JSON(http.StatusOK, ResponseData{Result: "false", Message: "No results"})
	}
	return c.JSON(http.StatusOK, data)
}
