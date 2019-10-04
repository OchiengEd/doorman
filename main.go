package main

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"

	jwt "github.com/dgrijalva/jwt-go"
)

var signingKey = []byte("murdershewrote")

var (
	verificationKey *rsa.PublicKey
	privateKey      *rsa.PrivateKey
)

// AuthToken returned when user is authenticated
type AuthToken struct {
	Token string `json:"token"`
}

func init() {
	certificateFile, err := ioutil.ReadFile("/etc/doorman/certs/tls.crt")
	if err != nil {
		log.Fatal(err.Error())
	}
	verificationKey, err = jwt.ParseRSAPublicKeyFromPEM(certificateFile)
	if err != nil {
		log.Fatal(err.Error())
	}

	privateKeyFile, err := ioutil.ReadFile("/etc/doorman/certs/tls.key")
	if err != nil {
		log.Fatal(err.Error())
	}

	privateKey, err = jwt.ParseRSAPrivateKeyFromPEM(privateKeyFile)
	if err != nil {
		log.Fatal(err.Error())
	}
}

func validateUser(user User) (string, error) {
	authUser, err := authUser(user)
	var authTime time.Duration

	if err != nil {
		authTime = -1
	} else {
		authTime = 360 // six hours
	}

	fmt.Println(privateKey)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"name":     strings.Join([]string{authUser.Firstname, authUser.Lastname}, " "),
		"username": authUser.Username,
		"id":       authUser.ID,
		"exp":      time.Now().Add(time.Minute * authTime).Unix(),
		"iat":      time.Now().Unix(),
	})

	tokenString, err := token.SignedString(signingKey)
	return tokenString, err
}

func authUserHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var user User
	json.NewDecoder(r.Body).Decode(&user)

	if user.Username == "" {
		log.Fatal("Your username cannot be blank")
		return
	} else if user.Password == "" {
		log.Fatal("Your password cannot be blank")
		return
	}

	jwtToken, err := validateUser(user)
	if err != nil {
		log.Fatal(err)
	}

	authToken := AuthToken{Token: jwtToken}

	json.NewEncoder(w).Encode(authToken)
}

func isAuthorized(token string) error {
	var jwtString string
	if len(strings.Split(token, " ")) == 2 {
		jwtString = strings.Split(token, " ")[1]
	}

	verifyKey, err := jwt.Parse(jwtString, func(token *jwt.Token) (interface{}, error) {
		return verificationKey, nil
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	fmt.Println(verifyKey)

	return nil
}

func createUserHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var user User
	json.NewDecoder(r.Body).Decode(&user)

	createUser(user)
}

func updateUserHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var updatedUser User
	json.NewDecoder(r.Body).Decode(&updatedUser)

}

func getUserHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	targetUser := getUser(params["id"])
	json.NewEncoder(w).Encode(targetUser)
}

func listUsersHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	// if err := isAuthorized(r.Header.Get("Authorization")); err != nil {
	// 	log.Fatal(err)
	// }
	systemUsers := getUsersList()
	json.NewEncoder(w).Encode(systemUsers)
}

func deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var targetUser User
	json.NewDecoder(r.Body).Decode(&targetUser)
	deleteUser(targetUser)
}

func main() {
	route := mux.NewRouter()
	route.HandleFunc("/user/register", createUserHandler).Methods("POST")
	route.HandleFunc("/user/{id}", getUserHandler).Methods("GET", "HEAD")
	route.HandleFunc("/users/list", listUsersHandler).Methods("GET", "HEAD")
	route.HandleFunc("/user", updateUserHandler).Methods("PUT")
	route.HandleFunc("/user", deleteUserHandler).Methods("DELETE")
	route.HandleFunc("/user/login", authUserHandler).Methods("POST")

	if err := http.ListenAndServe(":8080", route); err != nil {
		log.Fatal(err)
	}
}
