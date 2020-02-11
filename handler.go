package main

import (
	"time"
	"net/http"

	"github.com/labstack/echo"
	"github.com/dgrijalva/jwt-go"

)

type handler struct {}
type LoginData struct {
	Id  string `json:"Id" form:"Id" query:"Id"`
	Password string `json:"Password" form:"Password" query:"Password"`
}
type RegisterData struct {
	Id  string `json:"Id" form:"Id" query:"Id"`
	Password string `json:"Password" form:"Password" query:"Password"`
	Name  string `json:"Name" form:"Name" query:"Name"`
	Email string `json:"Email" form:"Email" query:"Email"`
}
type ResponseData struct {
	Result  string `json:"result" form:"result" query:"result"`
	Message string `json:"message" form:"message" query:"message"`
}

func (h *handler) userLogin(c echo.Context) error {
	data := new(LoginData)
	c.Bind(data)
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["id"] = data.Id
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

	t, err := token.SignedString([]byte("secret"))
	if err != nil {
		return err
	}

	//sqlite3 database search done

	return c.JSON(http.StatusOK, map[string]string{
		"token": t, "succeed" : "true", "id": data.Id,
	})
}
func (h *handler) userRegister(c echo.Context) error {
	data := new(RegisterData)
	c.Bind(data);
	
	response := ResponseData{Result:"true", Message:"register succeed"}

	//sqlite3 database create done

	return c.JSON(http.StatusOK,response)
}

func (h *handler) privateInfo(c echo.Context) error {
	user := c.Get("user").(*jwt.Token)
    claims := user.Claims.(jwt.MapClaims)
    id := claims["id"].(string)
    return c.String(http.StatusOK, "Welcome "+id+"!")
}