package main

import (
	"bufio"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	simplejson "github.com/bitly/go-simplejson"
	"github.com/gorilla/mux"
	"github.com/tredoe/osutil/user/crypt/sha512_crypt"
)

func getHash(salt string, secret string) string {
	c := sha512_crypt.New()
	hash, err := c.Generate([]byte(secret), []byte(salt))
	if err != nil {
		panic(err)
	}
	return hash
}

type User struct {
	mail string
	hash string
	salt string
}

func findUser(mail string) User {
	file, err := os.Open("./config/postfix-accounts.cf")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	var u User
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		str := scanner.Text()

		i := strings.Index(str, "|")
		email := str[:i]

		startSalt := strings.Index(str, "$")
		endSalt := strings.LastIndex(str, "$")
		salt := str[startSalt : endSalt+1]
		hash := str[startSalt:]
		if mail == email {
			u.hash = hash
			u.mail = email
			u.salt = salt
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return u
}

func writeToFile(mail string, newHash string) {
	input, err := ioutil.ReadFile("./config/postfix-accounts.cf")
	if err != nil {
		log.Fatalln(err)
	}

	lines := strings.Split(string(input), "\n")

	for i, line := range lines {
		if strings.HasPrefix(line, mail) {
			lines[i] = mail + "|{SHA512-CRYPT}" + newHash
		}
	}
	output := strings.Join(lines, "\n")
	err = ioutil.WriteFile("./config/postfix-accounts.cf", []byte(output), 0644)
	if err != nil {
		log.Fatalln(err)
	}
}

type UserInput struct {
	Mail          string
	CurrentSecret string
	NewSecret     string
}

func changePassword(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var u UserInput
	err := decoder.Decode(&u)
	if err != nil {
		panic(err)
	}
	json := simplejson.New()
	user := findUser(u.Mail)
	if (User{}) == user {
		json.Set("status", "user not found")
		w.WriteHeader(http.StatusNotFound)
	} else {
		currentHash := getHash(string(user.salt), u.CurrentSecret)
		if currentHash == user.hash {
			newHash := getHash(string(user.salt), u.NewSecret)
			writeToFile(user.mail, newHash)
			json.Set("status", "changed password")
		} else {
			json.Set("status", "wrong password")
			w.WriteHeader(http.StatusForbidden)
		}
	}
	payload, err := json.MarshalJSON()
	w.Header().Set("Content-Type", "application/json")
	w.Write(payload)
}

func serveForm(w http.ResponseWriter, r *http.Request) {
	http.FileServer(http.Dir("./static/"))
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/change", changePassword).Methods("POST")
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))
	log.Fatal(http.ListenAndServe(":8000", router))
}
