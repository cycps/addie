package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/cycps/addie"
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

func updateElement(id addie.Id, e addie.Identify) {

	design.Elements[e.Identify()] = e
	if e.Identify() != id {
		log.Println("deleting ", id)
		delete(design.Elements, id)
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

	for _, u := range msg.Elements {

		switch u.Type {
		case "Computer":
			var c addie.Computer
			err := json.Unmarshal(u.Element, &c)
			if err != nil {
				log.Println("unable to unmarshal computer")
			}
			updateElement(u.OID, c)
		}
		//TODO other unmarshallers go here

	}

	log.Println("\n", design.String())

	//send response
	w.WriteHeader(http.StatusOK)

}

func onDelete(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	/*
		//unpack the message
		msg := new(protocol.Delete)
		err := unpack(r, msg)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		//perform the deletes
		for _, id := range msg.Computers {
			delete(design.Computers, id)
		}
		for _, id := range msg.Switches {
			delete(design.Switches, id)
		}
		for _, id := range msg.Routers {
			delete(design.Routers, id)
		}
		for _, id := range msg.Models {
			delete(design.Models, id)
		}
		for _, id := range msg.Sensors {
			delete(design.Sensors, id)
		}
		for _, id := range msg.Actuators {
			delete(design.Actuators, id)
		}

		//send response
		w.WriteHeader(http.StatusOK)
	*/
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
