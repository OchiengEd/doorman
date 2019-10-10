package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
)

// User is a type that represents users of the system
type User struct {
	ID        string `json:"id,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	DeletedAt string `json:"deleted_at,omitempty"`
	Firstname string `json:"firstname,omitempty"`
	Lastname  string `json:"lastname,omitempty"`
	Username  string `json:"username,omitempty"`
	Password  string `json:"password,omitempty"`
}

var (
	databaseUsername = os.Getenv("DATABASE_USERNAME")
	databasePasswd   = os.Getenv("DATABASE_PASSWORD")
	databaseHost     = os.Getenv("DATABASE_HOST")
	databaseName     = os.Getenv("DATABASE")
)

func currentTimeUTC() string {
	currTime := time.Now().UTC()
	return currTime.Format(time.RFC3339)
}

func newUserID() string {
	return uuid.New().String()
}

func openDatabase() *sql.DB {
	connectionString := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s",
		databaseUsername, databasePasswd, databaseHost, databaseName)
	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		panic(err)
	}
	return db
}

func createUser(user User) {
	database := openDatabase()
	defer database.Close()

	user.CreatedAt = currentTimeUTC()
	user.ID = newUserID()

	createQuery := "INSERT INTO user(id, created_at, firstname, lastname, username, password) VALUES (?, ?, ?, ?, ?, ?)"
	insert, err := database.Query(createQuery, &user.ID, &user.CreatedAt, &user.Firstname, &user.Lastname, &user.Username, &user.Password)
	if err != nil {
		panic(err.Error())
	}
	defer insert.Close()
}

func getUser(userID string) User {
	database := openDatabase()
	defer database.Close()

	userQuery := "SELECT created_at, id, firstname, lastname, username FROM user WHERE id = ? AND deleted_at is NULL"
	rows, err := database.Query(userQuery, userID)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var targetUser User
	for rows.Next() {
		if err := rows.Scan(&targetUser.CreatedAt, &targetUser.ID, &targetUser.Firstname, &targetUser.Lastname, &targetUser.Username); err != nil {
			log.Fatal(err)
		}
	} // end for

	return targetUser
}

func getUsersList() []User {
	var allStaff []User
	database := openDatabase()
	defer database.Close()

	authQuery := "SELECT created_at, id, firstname, lastname, username FROM user"
	rows, err := database.Query(authQuery)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer rows.Close()

	var sysUser User
	for rows.Next() {
		if err := rows.Scan(&sysUser.ID, &sysUser.CreatedAt, &sysUser.Firstname, &sysUser.Lastname, &sysUser.Username); err != nil {
			log.Fatal(err)
		}
		allStaff = append(allStaff, sysUser)
	} // end for

	return allStaff
}

func updateUser(user User) {
	database := openDatabase()
	defer database.Close()

	updateTime := currentTimeUTC()
	updateQuery := "UPDATE user SET updated_at = ?, firstname = ?, lastname = ?, username = ? WHERE deleted_at is NULL AND id = ?"
	response, err := database.Query(updateQuery, updateTime, &user.Firstname, &user.Lastname, &user.Username, &user.ID)
	if err != nil {
		panic(err.Error())
	}
	defer response.Close()
}

func deleteUser(user User) {
	database := openDatabase()
	defer database.Close()

	deletionTime := currentTimeUTC()
	deleteQuery := "UPDATE user SET deleted_at = ? WHERE id = ?"
	rows, err := database.Query(deleteQuery, deletionTime, &user.ID)
	if err != nil {
		panic(err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		var user User

		err = rows.Scan(&user.CreatedAt, &user.Firstname, &user.Lastname, &user.Username, &user.Password)
		if err != nil {
			panic(err.Error())
		}
		fmt.Println(user.Username)
	}
}

func authUser(user User) (User, error) {
	database := openDatabase()
	defer database.Close()

	authQuery := "SELECT created_at, id, firstname, lastname, username FROM user WHERE username = ? AND password = ? AND deleted_at is NULL"
	rows, err := database.Query(authQuery, &user.Username, &user.Password)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var authUser User
	for rows.Next() {
		if err := rows.Scan(&authUser.CreatedAt, &authUser.ID, &authUser.Firstname, &authUser.Lastname, &authUser.Username); err != nil {
			log.Fatal(err)
		}
	} // end for

	if reflect.DeepEqual(authUser, User{}) {
		return authUser, errors.New("You entered wrong username or password")
	}
	// if authUser.Username == "" {
	// 	authUser.Username = "unauthorized"
	// 	authUser.Firstname = "Anonymous"
	// 	err = errors.New("error: User not found")
	// }

	return authUser, nil
}
