package main

import (
	"bytes"
	//"encoding/json"
	"fmt"
	"github.com/cycps/addie"
	"github.com/cycps/addie/protocol"
	"github.com/julienschmidt/httprouter"
	"github.com/pquerna/ffjson/ffjson"
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

func unpack(r *http.Request) interface{} {
	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)
	var x interface{}
	err := ffjson.Unmarshal(buf.Bytes(), &x)
	if err != nil {
		log.Println("[unpack] bad message")
		log.Println(err)
		log.Println(buf.String())
		return nil
	}
	return x
}

func onUpdate(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	//unpack the message

	msg := new(protocol.Update)

	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)
	err := ffjson.Unmarshal(buf.Bytes(), &msg)

	//upd := asUpdate(unpack(r))
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//perform the updates
	//for _, c := range msg.Computers {
	//		design.Computers[c.Id] = c
	//}

	//send response
	w.WriteHeader(http.StatusOK)

}

/*
func onDelete(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	//unpack the message
	var msg protocol.Delete
	if unpack(r, msg) != nil {
		return
	}

	//perform the deletes
	for _, id := range msg.Computers {
		delete(design.Computers, id)
	}

	//send response
	w.WriteHeader(http.StatusOK)
}
*/

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
