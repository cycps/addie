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

func dbCreate(e addie.Identify) {
	log.Printf("[dbCreate] %T '%s'", e, e.Identify())
	var err error = nil

	switch t := e.(type) {
	case addie.Computer:
		c := e.(addie.Computer)
		err = db.CreateComputer(c)
	case addie.Switch:
		s := e.(addie.Switch)
		err = db.CreateSwitch(s)
	case addie.Router:
		r := e.(addie.Router)
		err = db.CreateRouter(r)
	case addie.Link:
		l := e.(addie.Link)
		err = db.CreateLink(l)
		/*
			case addie.Model:
				m := e.(addie.Model)
			case addie.Sensor:
				s := e.(addie.Sensor)
			case addie.Actuator:
				a := e.(addie.Actuator)
		*/
	default:
		log.Printf("[dbCreate] unkown or unsupported element type: %T \n", t)
	}

	if err != nil {
		log.Println(err)
	}

}

func dbUpdate(oid addie.Id, e addie.Identify) {
	log.Printf("[dbUpdate] %T '%s'", e, e.Identify())
	var err error = nil

	switch t := e.(type) {
	case addie.Computer:
		c := e.(addie.Computer)
		_, err = db.UpdateComputer(oid, c)
	case addie.Switch:
		s := e.(addie.Switch)
		_, err = db.UpdateSwitch(oid, s)
	case addie.Router:
		r := e.(addie.Router)
		_, err = db.UpdateRouter(oid, r)
	case addie.Link:
		l := e.(addie.Link)
		_, err = db.UpdateLink(oid, l)
		/*
			case addie.Model:
				m := e.(addie.Model)
			case addie.Sensor:
				s := e.(addie.Sensor)
			case addie.Actuator:
				a := e.(addie.Actuator)
		*/
	default:
		log.Printf("[dbUpdate] unkown or unsupported element type: %T \n", t)
	}

	if err != nil {
		log.Println(err)
	}

}

//func updateElement(id addie.Id, e addie.Identify) {
func dedup(id addie.Id, e addie.Identify) {

	//_, ok := design.Elements[e.Identify()]

	//design.Elements[e.Identify()] = e
	if e.Identify() != id {
		log.Println("deleting ", id)
		log.Println("new-id ", e.Identify())
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

	var updates, creates []addie.Identify
	var update_oids []addie.Id

	var place = func(oid addie.Id, e addie.Identify) {
		_, ok := design.Elements[e.Identify()]
		if !ok {
			creates = append(creates, e)
		} else {
			updates = append(updates, e)
			update_oids = append(update_oids, oid)
		}
		design.Elements[e.Identify()] = e
	}

	for _, u := range msg.Elements {

		switch u.Type {
		case "Computer":
			var c addie.Computer
			err := json.Unmarshal(u.Element, &c)
			log.Println(c)
			if err != nil {
				log.Println("unable to unmarshal computer")
			}
			dedup(u.OID, c)
			place(u.OID, c)
		case "Switch":
			var s addie.Switch
			err := json.Unmarshal(u.Element, &s)
			if err != nil {
				log.Println("unable to unmarshal switch")
			}
			dedup(u.OID, s)
			place(u.OID, s)
		case "Router":
			var r addie.Router
			err := json.Unmarshal(u.Element, &r)
			if err != nil {
				log.Println("unable to unmarshal router")
			}
			dedup(u.OID, r)
			place(u.OID, r)
		case "Model":
			var m addie.Model
			err := json.Unmarshal(u.Element, &m)
			if err != nil {
				log.Println("unable to unmarshal model")
			}
			dedup(u.OID, m)
			place(u.OID, m)
		case "Sensor":
			var s addie.Sensor
			err := json.Unmarshal(u.Element, &s)
			if err != nil {
				log.Println("unable to unmarshal sensor")
			}
			dedup(u.OID, s)
			place(u.OID, s)
		case "Actuator":
			var a addie.Actuator
			err := json.Unmarshal(u.Element, &a)
			if err != nil {
				log.Println("unable to unmarshal sensor")
			}
			dedup(u.OID, a)
			place(u.OID, a)
		case "Link":
			var l addie.Link
			err := json.Unmarshal(u.Element, &l)
			if err != nil {
				log.Println("unable to unmarshal link")
			}
			dedup(u.OID, l)
			place(u.OID, l)
		default:
			log.Println("unkown element type: ", u.Type)
		}

	}

	for i, u := range updates {
		dbUpdate(update_oids[i], u)
	}
	for _, c := range creates {
		dbCreate(c)
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
