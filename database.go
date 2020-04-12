package main

import (
	"database/sql"
	"errors"
	"log"

	"golang.org/x/crypto/bcrypt"

	_ "github.com/mattn/go-sqlite3"
)

var ErrAlreadyLogined error = errors.New("Already logined. cannot login")
var ErrPasswordNotMatch error = errors.New("Password don't match. cannot login")
var ErrIdNotFound error = errors.New("Id couldn't be found")

type User struct {
	Id       string `json:"id" form:"id" query:"id"`
	UserId   string `json:"userId" form:"userId" query:"userId"`
	Password string `json:"password" form:"password" query:"password"`
	Name     string `json:"name" form:"name" query:"name"`
	Age      string `json:"age" form:"age" query:"age"`
	Email    string `json:"email" form:"email" query:"email"`
	Login    string `json:"login" form:"login" query:"login"`
}
type PublicUser struct {
	Id   string `json:"Id" form:"id" query:"id"`
	Name string `json:"Name" form:"name" query:"name"`
	Age  string `json:"Age" form:"age" query:"age"`
}

func InitDB(file string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", file)
	if err != nil {
		return nil, err
	}

	createTableQuery := `
		create table IF NOT EXISTS useraccount ( 
		id integer PRIMARY KEY autoincrement,
		userId text,
		password text,
		name text,
		age integer,
		email text,
		login integer,
		UNIQUE (id, userId)
		)
	`
	_, e := createTable(db, createTableQuery)
	if e != nil {
		return nil, e
	}
	return db, nil
}
func createTable(db *sql.DB, query string) (sql.Result, error) {
	return db.Exec(query)
}
func AddUser(db *sql.DB, id string, password string, name string, age int, email string) error {
	tx, _ := db.Begin()
	stmt, _ := tx.Prepare("insert into useraccount (userId,password,name,age,email, login) values (?,?,?,?,?,?)")
	_, err := stmt.Exec(id, password, name, age, email, 0)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	tx.Commit()
	return nil
}

func GetUser(db *sql.DB, userId string) (User, error) {
	var user User
	rows := db.QueryRow("select * from useraccount where userId = $1", userId)
	err := rows.Scan(&user.Id, &user.UserId, &user.Password, &user.Name,
		&user.Age, &user.Email, &user.Login)
	if err != nil {
		return User{}, err
	}

	return user, nil
}
func GetUserPublic(db *sql.DB, userId string) (User, error) {
	var user User
	rows := db.QueryRow("select Id, UserId,Age, Name, Email, Login from useraccount where userId = $1", userId)
	err := rows.Scan(&user.Id, &user.UserId, &user.Age, &user.Name,
		&user.Email, &user.Login)
	if err != nil {
		return User{}, err
	}

	return user, nil
}
func Login(db *sql.DB, userId string, password string) error {
	u, err := GetUser(db, userId)

	if err != nil {
		return ErrIdNotFound
	} else if pwerr := CompareBcrypt([]byte(u.Password), []byte(password)); pwerr != nil { // not same
		return ErrPasswordNotMatch
	} else if u.Login == "1" {
		return ErrAlreadyLogined
	}
	tx, _ := db.Begin()
	stmt, _ := tx.Prepare("update useraccount set login=1 where userId=?")
	_, dberr := stmt.Exec(u.UserId)
	if dberr != nil {
		log.Println("db error, ", dberr.Error())
		return dberr
	}
	tx.Commit()
	return nil
}
func Logout(db *sql.DB, userId string) error {
	tx, _ := db.Begin()
	stmt, _ := tx.Prepare("update useraccount set login=0 where userId=?")
	_, dberr := stmt.Exec(userId)
	if dberr != nil {
		log.Println("logout error, ", dberr.Error())
		return dberr
	}
	tx.Commit()
	return nil
}

func CompareBcrypt(hash []byte, plain []byte) error {
	err := bcrypt.CompareHashAndPassword(hash, plain)
	if err != nil {
		return err
	}

	return nil
}
