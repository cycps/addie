package main

import (
	"addie"
	"addie/db"
	"addie/protocol"
	"encoding/json"
	"fmt"
	"github.com/deter-project/go-spi/spi"
	"github.com/julienschmidt/httprouter"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

var cypdir = os.ExpandEnv("/cypress")

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
	dur, err := time.ParseDuration("47m")
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

	//log into the DeterLab SPI
	cert, err := spi.Login(u, p)
	if err != nil {
		log.Println("Unable to login to Deter")
		log.Println(err)
		w.WriteHeader(401) //unauthorized
	}
	//if the cert is empty that means we are already logged in and the existing
	//cert file should be fine
	err = os.MkdirAll("/cypress/"+u, 0755)
	if err != nil {
		log.Println(err)
		log.Printf("failed to create user directory for %s", u)
	}

	if string(cert) != "" {
		err := ioutil.WriteFile("/cypress/"+u+"/spi.cert", cert, 0644)
		if err != nil {
			log.Println(err)
			log.Printf("failed to write user cert for %s", u)
		}
	}

	log.Printf("user login success: '%s'", u)

	w.WriteHeader(200)
}

func getUser(r *http.Request) (string, error) {

	cookie, err := r.Cookie("cypress-session-cookie")
	if err != nil {
		log.Println(err)
		return "", fmt.Errorf("[getUser] error getting cookie")
	}
	if cookie == nil {
		return "", fmt.Errorf("[getUser] nil cookie")
	}
	user, ok := userCookies[cookie.Value]
	if !ok {
		return "", fmt.Errorf("[getUser] unkown cookie '%s'", cookie.Value)
	}

	log.Printf("[getUser] %s", user)

	return user, nil
}

func thisUser(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	user, err := getUser(r)
	if err != nil {
		w.WriteHeader(401)
		return
	}

	w.Write([]byte(user))

}

func launchAddie(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	log.Printf("addie launch requested")

	u := r.URL.Query().Get("user")
	if u == "" {
		log.Println("Caller requested addie launch with no user")
		w.WriteHeader(401)
		return
	}
	d := r.URL.Query().Get("design")
	if d == "" {
		log.Println("Caller requested addie launch with no target design")
		w.WriteHeader(401)
		return
	}
	err := exec.Command("pgrep", "-lfa", fmt.Sprintf("addie %s %s", u, d)).Run()
	if err == nil {
		log.Printf("addie is already running for %s-%s, doing nothing", u, d)
	} else {
		log.Printf("launching an addie instance for %s", d)
		cmd := exec.Command("addie", u, d)
		cmd.Start() //lost muffins?
		time.Sleep(500 * time.Millisecond)
	}

	w.WriteHeader(200)
}

func newXP(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	user, err := getUser(r)
	if err != nil {
		w.WriteHeader(401)
		return
	}

	msg := new(protocol.NewXP)
	err = protocol.Unpack(r, &msg)
	if err != nil {
		log.Printf("error unpacking newXP message")
		log.Println(err)
		w.WriteHeader(400)
		return
	}

	log.Printf("Creating xp '%s` for user `%s`", msg.Name, user)

	err = db.CreateDesign(msg.Name, user)
	if err != nil {
		log.Println(err)
		log.Printf("[newXP] error creating design entry db")
		w.WriteHeader(500)
		return
	}

	design_key, err := db.ReadDesignKey(msg.Name, user)
	if err != nil {
		log.Println(err)
		log.Printf("[newXP] error reading design key")
		w.WriteHeader(500)
		return
	}

	//use the default sim settings to begin with
	s := addie.SimSettings{}
	s.Begin = 0
	s.End = 10
	s.MaxStep = 1e-3

	err = db.CreateSimSettings(s, design_key)
	if err != nil {
		log.Println(err)
		log.Println("[newXP] error creating default sim settings")
		w.WriteHeader(500)
		return
	}

	w.WriteHeader(200)
}

func myDesigns(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	user, err := getUser(r)
	if err != nil {
		w.WriteHeader(401)
		return
	}

	ds, err := db.ReadUserDesigns(user)
	if err != nil {
		log.Printf("[myDesigns] error reading user projects for '%s'", user)
		w.WriteHeader(500)
		return
	}

	var uds protocol.UserDesigns
	uds.Designs = ds

	js, err := json.Marshal(uds)
	if err != nil {
		log.Printf("[myDesigns] error marshalling json")
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)

}

func main() {

	router := httprouter.New()
	router.POST("/login", onLogin)
	router.GET("/thisUser", thisUser)
	router.POST("/newXP", newXP)
	router.GET("/myDesigns", myDesigns)
	router.GET("/launchAddie", launchAddie)

	log.Printf("listening on http://::0:8081")
	log.Fatal(
		http.ListenAndServeTLS(":8081",
			cypdir+"/keys/cert.pem",
			cypdir+"/keys/key.pem",
			router))

}
