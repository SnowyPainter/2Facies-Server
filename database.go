package main

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"

)

type User struct {
	Id       int    `json:"id" form:"id" query:"id"`
	UserId   string `json:"userId" form:"userId" query:"userId"`
	Password string `json:"password" form:"password" query:"password"`
	Name     string `json:"name" form:"name" query:"name"`
	Age      string `json:"age" form:"age" query:"age"`
	Email    string `json:"email" form:"email" query:"email"`
}

func InitDB(file string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", file)
	if err != nil {
		return nil, err
	}

	createTableQuery := `
		create table IF NOT EXISTS useraccount ( 
		id integer primary key autoincrement,
		userId text primary key,
		password text,
		name text,
		age integer,
		email text
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
	stmt, _ := tx.Prepare("insert into useraccount (userId,password,name,age,email) values (?,?,?,?,?)")
	_, err := stmt.Exec(id, password, name, age, email)
	if err != nil {
		return err
	}
	tx.Commit()
	return nil
}

func GetUser(db *sql.DB, userId string) (User, error) {
	var user User
	rows := db.QueryRow("select * from useraccount where userId = $1", userId)
	err := rows.Scan(&user.Id, &user.UserId, &user.Password, &user.Name,
		&user.Age, &user.Email)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

func LoginCheck(db *sql.DB, userId string, password string) error {
	return nil
}
