package main

import (
	"encoding/json"
	"fmt"
	"github.com/cycps/addie"
	"github.com/cycps/addie/db"
	"github.com/cycps/addie/protocol"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"os"
	"reflect"
)

var design addie.Design
var cypdir = os.ExpandEnv("$HOME/.cypress")
var user = ""

func main() {

	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "usage: go-addie <user id> <design id>\n")
		os.Exit(1)
	}

	loadDesign(os.Args[2])
	user = os.Args[1]
	listen()
}

func loadDesign(id string) {

	design = addie.EmptyDesign(id)

}

func dbCreate(e addie.Identify) {
	log.Printf("[dbCreate] %T '%s'", e, e.Identify())
	var err error = nil

	switch t := e.(type) {
	case addie.Computer:
		c := e.(addie.Computer)
		err = db.CreateComputer(c, user)
	case addie.Switch:
		s := e.(addie.Switch)
		err = db.CreateSwitch(s, user)
	case addie.Router:
		r := e.(addie.Router)
		err = db.CreateRouter(r, user)
	case addie.Link:
		l := e.(addie.Link)
		err = db.CreateLink(l, user)
	case addie.Phyo:
		m := e.(addie.Phyo)
		db.CreatePhyo(m, user)
	case addie.Sax:
		s := e.(addie.Sax)
		db.CreateSax(s, user)
	case addie.Plink:
		p := e.(addie.Plink)
		db.CreatePlink(p, user)
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
		_, err = db.UpdateComputer(oid, old.(addie.Computer), c, user)
	case addie.Switch:
		s := e.(addie.Switch)
		_, err = db.UpdateSwitch(oid, old.(addie.Switch), s, user)
	case addie.Router:
		r := e.(addie.Router)
		_, err = db.UpdateRouter(oid, old.(addie.Router), r, user)
	case addie.Link:
		l := e.(addie.Link)
		_, err = db.UpdateLink(oid, l, user)
	case addie.Phyo:
		m := e.(addie.Phyo)
		_, err = db.UpdatePhyo(oid, m, user)
	case addie.Sax:
		s := e.(addie.Sax)
		_, err = db.UpdateSax(oid, old.(addie.Sax), s, user)
	case addie.Plink:
		p := e.(addie.Plink)
		_, err = db.UpdatePlink(oid, p, user)
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
	err := protocol.Unpack(r, msg)
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
			case addie.Link, addie.Plink:
				new_links = append(new_links, e)
			default:
				new_nodes = append(new_nodes, e)
			}
		} else {
			switch e.(type) {
			case addie.Link, addie.Plink:
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
		case "Phyo":
			var m addie.Phyo
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
		case "Plink":
			var p addie.Plink
			err := json.Unmarshal(u.Element, &p)
			if err != nil {
				log.Println("unable to unmarshal plink")
			}
			place(u.OID, p)
		case "Sax":
			var s addie.Sax
			err := json.Unmarshal(u.Element, &s)
			if err != nil {
				log.Println("unable to marshal sax")
			}
			place(u.OID, s)
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

func onRead(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	//log.Println(design)
	json, err := modelJson()
	//log.Println(string(json))

	if err != nil {
		log.Println("modelJson failed")
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(json)

}

type TypeWrapper struct {
	Type   string      `json:"type"`
	Object interface{} `json:"object"`
}

func typeWrap(obj interface{}) TypeWrapper {

	return TypeWrapper{Type: reflect.TypeOf(obj).Name(), Object: obj}

}

func doRead() error {

	dsg, err := db.ReadDesign(design.Name, user)
	if err != nil {
		log.Println(err)
		return fmt.Errorf("Failed to read design")
	}

	design = *dsg

	return nil
}

type JsonModel struct {
	Name     string        `json:"name"`
	Elements []TypeWrapper `json:"elements"`
}

func modelJson() ([]byte, error) {

	var mdl JsonModel
	mdl.Name = design.Name

	mdl.Elements = make([]TypeWrapper, len(design.Elements))

	i := 0
	for _, v := range design.Elements {
		mdl.Elements[i] = typeWrap(v)
		i++
	}

	_json, err := json.Marshal(mdl)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("Failed to marshal element array to JSON")
	}

	return _json, nil

}

func listen() {

	router := httprouter.New()
	//router.POST("/:xpid/design/update", onUpdate)
	router.POST("/"+design.Name+"/design/update", onUpdate)
	router.GET("/"+design.Name+"/design/read", onRead)
	//router.POST("/:xpid/design/delete", onDelete)

	err := doRead()
	if err != nil {
		log.Println(err)
		log.Fatal("Could not read design from db")
	}

	log.Printf("listening on https://::0:8080/%s/design/", design.Name)
	log.Fatal(
		http.ListenAndServeTLS(":8080",
			cypdir+"/keys/cert.pem",
			cypdir+"/keys/key.pem",
			router))

}
