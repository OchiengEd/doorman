package main

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

// Credentials for database
type Credentials struct {
	Username string
	Password string
	Database string
	Hostname string
}

// Table info
type Table struct {
	Name string
}

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

func scramblePassword(password string) string {
	hashedPasswd, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	return string(hashedPasswd)
}

func newUserID() string {
	return uuid.New().String()
}

func currentTimeUTC() string {
	currTime := time.Now().UTC()
	return currTime.Format(time.RFC3339)
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

// IfTableExists checks if table exists
func IfTableExists(table string) bool {
	database := openDatabase()
	defer database.Close()

	err := database.Ping()
	for err != nil {
		log.Error("Database not ready. %v", err)
		time.Sleep(time.Second * 5)
		err = database.Ping()
	}

	query := fmt.Sprintf("SHOW TABLES LIKE \"%s\"", table)
	rows, err := database.Query(query)
	if err != nil {
		log.Error(" %v", err)
	}

	tbl := &Table{}
	for rows.Next() {
		if err := rows.Scan(&tbl.Name); err != nil {
			log.Error("Error retrieving rows. %v", err)
		}
	}
	defer rows.Close()

	if tbl.Name == "" {
		return false
	}

	return true
}

func createTable() {
	database := openDatabase()
	defer database.Close()

	query := "CREATE TABLE `user` (`created_at` varchar(20) DEFAULT NULL, `updated_at` varchar(20) DEFAULT NULL, `deleted_at` varchar(20) DEFAULT NULL, `id` varchar(48) NOT NULL, `firstname` varchar(48) DEFAULT NULL, `lastname` varchar(48) DEFAULT NULL, `username` varchar(48) NOT NULL, `password` varchar(128) DEFAULT NULL, UNIQUE KEY `username` (`username`), UNIQUE KEY `id` (`id`)) ENGINE=InnoDB"
	_, err := database.Query(query)
	if err != nil {
		log.Error("Error creating table. %v", err)
	}

	admin := User{
		Firstname: "Administrator",
		Username:  "admin",
		Password:  "doorman",
	}
	createUser(admin)
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

func main() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02T15:04:05.000",
	})

	if !IfTableExists("user") {
		createTable()
	}

	fmt.Println("Table is accessible")
}
