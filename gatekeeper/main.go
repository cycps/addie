package main

import (
	"fmt"
	"github.com/cycps/addie/db"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

var cypdir = os.ExpandEnv("$HOME/.cypress")

func authUser(name, password string) (bool, error) {

	q := fmt.Sprintf("SELECT "+
		"(pwh = crypt('%s', pwh)) as pwmatch FROM users "+
		"WHERE name = '%s'", password, name)

	rows, err := db.RunQ(q)
	if err != nil {
		log.Println(err)
		return false, fmt.Errorf("unable to query db")
	}
	if !rows.Next() {
		log.Println("user does not exist")
		return false, nil
	}
	var pwmatch bool
	err = rows.Scan(&pwmatch)
	if err != nil {
		log.Println("error reading db result")
	}
	if !pwmatch {
		log.Printf("bad password for user '%s", name)
	}

	return pwmatch, nil
}

func newSessionCookie() (*http.Cookie, error) {

	uuid, err := exec.Command("uuidgen").Output()
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("uuidgen failed")
	}
	cookie := new(http.Cookie)
	cookie.Name = "cypress-session-cookie"
	cookie.Value = strings.TrimSuffix(string(uuid), "\n")
	cookie.Path = "/"
	dur, err := time.ParseDuration("1h")
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("parse duration failed")
	}
	cookie.Expires = time.Now().Add(dur)

	return cookie, nil

}

var userCookies = make(map[string]string)

func onLogin(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	u := r.FormValue("username")
	p := r.FormValue("password")

	log.Printf("user login: '%s'", u)

	isValidUser, err := authUser(u, p)
	if err != nil {
		log.Println(err)
		log.Println("user auth failed")
		w.WriteHeader(500)
		return
	}
	if !isValidUser {
		log.Printf("scalawagerry detected from '%s'", u)
		w.WriteHeader(401) //unauthorized
		return
	}

	//if we are here the user is valid
	cookie, err := newSessionCookie()
	if err != nil {
		log.Println("error creating session cookie")
		log.Println(err)
		w.WriteHeader(500)
		return
	}
	userCookies[cookie.Value] = u
	http.SetCookie(w, cookie)

	log.Printf("user login success: '%s'", u)

	w.WriteHeader(200)
}

func thisUser(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	cookie, err := r.Cookie("cypress-session-cookie")
	if err != nil {
		log.Println(err)
		log.Println("[thisUser] error getting cookie")
		w.WriteHeader(401)
		return
	}
	if cookie == nil {
		log.Println("[thisUser] nil cookie")
		w.WriteHeader(401)
		return
	}
	user, ok := userCookies[cookie.Value]
	if !ok {
		log.Printf("[thisUser] unkown cookie '%s'", cookie.Value)
		w.WriteHeader(401)
		return
	}

	log.Printf("[thisUser] %s", user)

	w.Write([]byte(user))

}

func main() {

	router := httprouter.New()
	router.POST("/login", onLogin)
	router.GET("/thisUser", thisUser)

	log.Printf("listening on http://::0:8081")
	log.Fatal(
		http.ListenAndServeTLS(":8081",
			cypdir+"/keys/cert.pem",
			cypdir+"/keys/key.pem",
			router))

}
