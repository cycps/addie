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

	old, ok := design.Elements[oid]
	if !ok {
		log.Printf("[Update] bad oid %v\n", oid)
		return
	}

	//todo perform check old.(type) conversion
	switch t := e.(type) {
	case addie.Computer:
		c := e.(addie.Computer)
		_, err = db.UpdateComputer(oid, old.(addie.Computer), c)
	case addie.Switch:
		s := e.(addie.Switch)
		_, err = db.UpdateSwitch(oid, old.(addie.Switch), s)
	case addie.Router:
		r := e.(addie.Router)
		_, err = db.UpdateRouter(oid, old.(addie.Router), r)
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

func onUpdate(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	//unpack the message
	msg := new(protocol.Update)
	err := unpack(r, msg)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var new_nodes, new_links []addie.Identify
	var changed_nodes, changed_links []addie.Identify
	var changed_node_oids, changed_link_oids []addie.Id
	var killList []addie.Id

	var place = func(oid addie.Id, e addie.Identify) {

		if e.Identify() != oid {
			killList = append(killList, oid)
		}
		_, ok := design.Elements[oid]
		if !ok {
			switch e.(type) {
			case addie.Link:
				new_links = append(new_links, e)
			default:
				new_nodes = append(new_nodes, e)
			}
		} else {
			switch e.(type) {
			case addie.Link:
				changed_links = append(changed_links, e)
				changed_link_oids = append(changed_link_oids, oid)
			default:
				changed_nodes = append(changed_nodes, e)
				changed_node_oids = append(changed_node_oids, oid)
			}
		}

	}

	for _, u := range msg.Elements {

		switch u.Type {
		case "Computer":
			var c addie.Computer
			err := json.Unmarshal(u.Element, &c)
			if err != nil {
				log.Println("unable to unmarshal computer")
			}
			place(u.OID, c)
		case "Switch":
			var s addie.Switch
			err := json.Unmarshal(u.Element, &s)
			if err != nil {
				log.Println("unable to unmarshal switch")
			}
			place(u.OID, s)
		case "Router":
			var r addie.Router
			err := json.Unmarshal(u.Element, &r)
			if err != nil {
				log.Println("unable to unmarshal router")
			}
			place(u.OID, r)
		case "Model":
			var m addie.Model
			err := json.Unmarshal(u.Element, &m)
			if err != nil {
				log.Println("unable to unmarshal model")
			}
			place(u.OID, m)
		case "Sensor":
			var s addie.Sensor
			err := json.Unmarshal(u.Element, &s)
			if err != nil {
				log.Println("unable to unmarshal sensor")
			}
			place(u.OID, s)
		case "Actuator":
			var a addie.Actuator
			err := json.Unmarshal(u.Element, &a)
			if err != nil {
				log.Println("unable to unmarshal sensor")
			}
			place(u.OID, a)
		case "Link":
			var l addie.Link
			err := json.Unmarshal(u.Element, &l)
			if err != nil {
				log.Println("unable to unmarshal link")
			}
			place(u.OID, l)
		default:
			log.Println("unkown element type: ", u.Type)
		}

	}

	for i, u := range changed_nodes {
		dbUpdate(changed_node_oids[i], u)
		design.Elements[u.Identify()] = u
	}
	for _, c := range new_nodes {
		dbCreate(c)
		design.Elements[c.Identify()] = c
	}

	for i, u := range changed_links {
		dbUpdate(changed_link_oids[i], u)
		design.Elements[u.Identify()] = u
	}

	for _, c := range new_links {
		dbCreate(c)
		design.Elements[c.Identify()] = c
	}

	for _, k := range killList {
		delete(design.Elements, k)
	}

	//log.Println("\n", design.String())

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
