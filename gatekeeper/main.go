package main

import (
	"crypto/rand"
	"crypto/sha512"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

var cypdir = os.ExpandEnv("$HOME/.cypress")

func onLogin(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	//buf := new(bytes.Buffer)
	//buf.ReadFrom(r.Body)
	u := r.FormValue("username")
	p := r.FormValue("password")

	setPassword(u, p)

	log.Printf("login -- (%s,%s)", u, p)

	uuid, err := exec.Command("uuidgen").Output()
	if err != nil {
		log.Println(err)
		log.Println("uuidgen failed")
		w.WriteHeader(500)
		return
	}
	cookie := new(http.Cookie)
	cookie.Name = "cypress-session-cookie"
	cookie.Value = strings.TrimSuffix(string(uuid), "\n")
	cookie.Path = "/"
	dur, err := time.ParseDuration("1h")
	if err != nil {
		log.Println(err)
		log.Println("parse duration failed")
		w.WriteHeader(500)
		return
	}
	cookie.Expires = time.Now().Add(dur)

	http.SetCookie(w, cookie)

	w.Write([]byte("hello there " + u))
}

func setPassword(user string, password string) error {
	salt := make([]byte, 32)
	_, err := io.ReadFull(rand.Reader, salt)
	if err != nil {
		log.Println(err)
		return fmt.Errorf("failed to generate salt")
	}
	salted := string(salt) + password

	hash := sha512.Sum512([]byte(salted))
	log.Printf("%x", hash)

	return nil
}

func main() {

	router := httprouter.New()
	router.POST("/login", onLogin)

	log.Printf("listening on http://::0:8081")
	log.Fatal(
		http.ListenAndServeTLS(":8081",
			cypdir+"/keys/cert.pem",
			cypdir+"/keys/key.pem",
			router))

}
