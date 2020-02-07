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
	"golang.org/x/crypto/bcrypt"
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

// Credentials for database
type Credentials struct {
	Username string
	Password string
	Database string
	Hostname string
}

func currentTimeUTC() string {
	currTime := time.Now().UTC()
	return currTime.Format(time.RFC3339)
}

func newUserID() string {
	return uuid.New().String()
}

func openDatabase() *sql.DB {
	cred := Credentials{
		Username: os.Getenv("DATABASE_USERNAME"),
		Password: os.Getenv("DATABASE_PASSWORD"),
		Database: os.Getenv("DATABASE"),
		Hostname: os.Getenv("DATABASE_HOST"),
	}

	connectionString := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s",
		cred.Username, cred.Password, cred.Hostname, cred.Database)
	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		panic(err)
	}
	return db
}

func scramblePassword(password string) string {
	hashedPasswd, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	return string(hashedPasswd)
}

func createUser(user User) {
	database := openDatabase()
	defer database.Close()

	user.CreatedAt = currentTimeUTC()
	user.ID = newUserID()
	user.Password = scramblePassword(user.Password)

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

	authQuery := "SELECT created_at, id, firstname, lastname, username, password FROM user WHERE username = ? AND deleted_at is NULL"
	rows, err := database.Query(authQuery, &user.Username)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var authUser User
	for rows.Next() {
		if err := rows.Scan(&authUser.CreatedAt, &authUser.ID, &authUser.Firstname, &authUser.Lastname, &authUser.Username, &authUser.Password); err != nil {
			log.Fatal(err)
		}
	} // end for

	if reflect.DeepEqual(authUser, User{}) {
		return authUser, errors.New("Verify credentials and try again")
	}

	// Check is password is correct
	// Send empty user struct if password is wrong
	if err := bcrypt.CompareHashAndPassword([]byte(authUser.Password), []byte(user.Password)); err != nil {
		return User{}, errors.New("You entered a wrong username or password")
	}

	// Remove password from struct so that it is not included in Claims
	authUser.Password = ""

	return authUser, nil
}
