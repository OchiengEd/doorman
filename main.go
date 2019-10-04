package main

import (
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"

	jwt "github.com/dgrijalva/jwt-go"
)

var (
	verificationKey *rsa.PublicKey
	signingKey      *rsa.PrivateKey
)

// AuthToken returned when user is authenticated
type AuthToken struct {
	Token string `json:"token"`
}

// MyClaims structure for the data in the JWT token
type MyClaims struct {
	Name string `json:"name"`
	jwt.StandardClaims
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

	signingKey, err = jwt.ParseRSAPrivateKeyFromPEM(privateKeyFile)
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

	token := jwt.NewWithClaims(jwt.GetSigningMethod("RS256"), jwt.MapClaims{
		"name": strings.Join([]string{authUser.Firstname, authUser.Lastname}, " "),
		"sub":  authUser.Username,
		"id":   authUser.ID,
		"exp":  time.Now().Add(time.Minute * authTime).Unix(),
		"iat":  time.Now().Unix(),
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

	verifyKey, err := jwt.ParseWithClaims(jwtString, &MyClaims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Method.Alg() != "RS256" {
			return nil, errors.New("Unsupported signing key")
		}

		return verificationKey, nil
	})
	if err != nil {
		log.Fatal(err.Error())
	}

	if claims, ok := verifyKey.Claims.(*MyClaims); ok && verifyKey.Valid {
		expTime, _ := time.Parse(time.RFC3339, strconv.FormatInt(claims.StandardClaims.ExpiresAt, 10))
		fmt.Printf("Token issued for %v %v", claims.Name, expTime)
	} else {
		return err
	}

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
	if err := isAuthorized(r.Header.Get("Authorization")); err != nil {
		log.Fatal(err)
	}
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

	if err := http.ListenAndServe(":5000", route); err != nil {
		log.Fatal(err)
	}
}
