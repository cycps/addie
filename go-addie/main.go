package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/cycps/addie"
	"github.com/cycps/addie/db"
	"github.com/cycps/addie/protocol"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"os"
)

var design addie.Design
var cypdir = os.ExpandEnv("$HOME/.cypress")

func main() {

	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: go-addie <design id>\n")
		os.Exit(1)
	}

	loadDesign(os.Args[1])
	listen()
}

func loadDesign(id string) {

	design = addie.EmptyDesign(id)

}

func unpack(r *http.Request, x interface{}) error {
	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)
	err := json.Unmarshal(buf.Bytes(), &x)
	if err != nil {
		log.Println("[unpack] bad message")
		log.Println(err)
		log.Println(buf.String())
		return nil
	}
	return nil
}

func dbInsert(e addie.Identify) {
	log.Printf("[dbInsert] %T '%s'", e, e.Identify())
	switch t := e.(type) {
	case addie.Computer:
		c := e.(addie.Computer)
		err := db.InsertComputer(c)
		if err != nil {
			log.Println(err)
		}
	case addie.Switch:
		s := e.(addie.Switch)
		err := db.InsertSwitch(s)
		if err != nil {
			log.Println(err)
		}
	case addie.Router:
		r := e.(addie.Router)
		err := db.InsertRouter(r)
		if err != nil {
			log.Println(err)
		}
	case addie.Link:
		l := e.(addie.Link)
		err := db.InsertLink(l)
		if err != nil {
			log.Println(err)
		}
		/*
			case addie.Model:
				m := e.(addie.Model)
			case addie.Sensor:
				s := e.(addie.Sensor)
			case addie.Actuator:
				a := e.(addie.Actuator)
		*/
	default:
		log.Printf("[dbInsert] unkown or unsupported element type: %T \n", t)
	}

}

func dbUpdate(e addie.Identify) {
	log.Printf("[dbUpdate] %T '%s'", e, e.Identify())
	switch t := e.(type) {
	case addie.Computer:
	case addie.Switch:
	case addie.Router:
	case addie.Model:
	case addie.Sensor:
	case addie.Actuator:
	case addie.Link:
	default:
		log.Printf("[dbUpdate] unkown or unsupported element type: %T \n", t)
	}

}

//func updateElement(id addie.Id, e addie.Identify) {
func dedup(id addie.Id, e addie.Identify) {

	//_, ok := design.Elements[e.Identify()]

	//design.Elements[e.Identify()] = e
	if e.Identify() != id {
		log.Println("deleting ", id)
		delete(design.Elements, id)
	}

	//collect updates and inserts, then do updates first as inserts
	//may depend on updated data
	//var updates, inserts []addie.Identify

	/*
		if !ok {
			inserts = append(inserts, e)
		} else {
			updates = append(updates, e)
		}

		for _, u := range updates {
			dbUpdate(u)
		}
		for _, i := range inserts {
			dbInsert(i)
		}
	*/

}

func onUpdate(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	//unpack the message
	msg := new(protocol.Update)
	err := unpack(r, msg)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var updates, inserts []addie.Identify

	var place = func(e addie.Identify) {
		_, ok := design.Elements[e.Identify()]
		if !ok {
			inserts = append(inserts, e)
		} else {
			updates = append(updates, e)
		}
		design.Elements[e.Identify()] = e
	}

	for _, u := range msg.Elements {

		switch u.Type {
		case "Computer":
			var c addie.Computer
			err := json.Unmarshal(u.Element, &c)
			if err != nil {
				log.Println("unable to unmarshal computer")
			}
			dedup(u.OID, c)
			place(c)
		case "Switch":
			var s addie.Switch
			err := json.Unmarshal(u.Element, &s)
			if err != nil {
				log.Println("unable to unmarshal switch")
			}
			dedup(u.OID, s)
			place(s)
		case "Router":
			var r addie.Router
			err := json.Unmarshal(u.Element, &r)
			if err != nil {
				log.Println("unable to unmarshal router")
			}
			dedup(u.OID, r)
			place(r)
		case "Model":
			var m addie.Model
			err := json.Unmarshal(u.Element, &m)
			if err != nil {
				log.Println("unable to unmarshal model")
			}
			dedup(u.OID, m)
			place(m)
		case "Sensor":
			var s addie.Sensor
			err := json.Unmarshal(u.Element, &s)
			if err != nil {
				log.Println("unable to unmarshal sensor")
			}
			dedup(u.OID, s)
			place(s)
		case "Actuator":
			var a addie.Actuator
			err := json.Unmarshal(u.Element, &a)
			if err != nil {
				log.Println("unable to unmarshal sensor")
			}
			dedup(u.OID, a)
			place(a)
		case "Link":
			var l addie.Link
			err := json.Unmarshal(u.Element, &l)
			if err != nil {
				log.Println("unable to unmarshal link")
			}
			dedup(u.OID, l)
			place(l)
		default:
			log.Println("unkown element type: ", u.Type)
		}

	}

	for _, u := range updates {
		dbUpdate(u)
	}
	for _, i := range inserts {
		dbInsert(i)
	}

	log.Println("\n", design.String())

	//send response
	w.WriteHeader(http.StatusOK)

}

func onDelete(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	//TODO thundermuffin

}

func listen() {

	router := httprouter.New()
	router.POST("/:xpid/design/update", onUpdate)
	//router.POST("/:xpid/design/delete", onDelete)

	log.Printf("listening on https://::0:8080/%s/", design.Name)
	log.Fatal(
		http.ListenAndServeTLS(":8080",
			cypdir+"/keys/cert.pem",
			cypdir+"/keys/key.pem",
			router))

}
