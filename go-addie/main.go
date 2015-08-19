package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/cycps/addie"
	"github.com/cycps/addie/db"
	"github.com/cycps/addie/deter"
	"github.com/cycps/addie/protocol"
	"github.com/cycps/addie/sim"
	"github.com/julienschmidt/httprouter"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"reflect"
	"strconv"
)

var design addie.Design
var userModels = make(map[string]addie.Model)
var simSettings addie.SimSettings
var cypdir = os.ExpandEnv("$HOME/.cypress")
var user = ""

func main() {

	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "usage: go-addie <user id> <design id>\n")
		os.Exit(1)
	}

	loadDesign(os.Args[2])
	user = os.Args[1]
	checkUserDir()
	loadUserModels()
	listen()
}

func checkUserDir() {
	os.MkdirAll("/cypress/"+user, 0755)
}

func loadDesign(id string) {

	design = addie.EmptyDesign(id)

}

func loadUserModels() {

	userModels = make(map[string]addie.Model)

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
		p := e.(addie.Phyo)
		_, err = db.CreatePhyo(p, user)
	case addie.Model:
		m := e.(addie.Model)
		err = db.CreateModel(m, user)
	case addie.Sax:
		s := e.(addie.Sax)
		_, err = db.CreateSax(s, user)
	case addie.Plink:
		p := e.(addie.Plink)
		err = db.CreatePlink(p, user)
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

	var old addie.Identify
	var ok bool

	if reflect.TypeOf(e).Name() == "Model" {
		old, ok = userModels[oid.Name]
		if !ok {
			log.Printf("[Update] bad oid %v\n", oid)
			return
		}
	} else {
		old, ok = design.Elements[oid]
		if !ok {
			log.Printf("[Update] bad oid %v\n", oid)
			return
		}
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
		p := e.(addie.Phyo)
		_, err = db.UpdatePhyo(oid, p, user)
	case addie.Model:
		m := e.(addie.Model)
		err = db.UpdateModel(oid.Name, m, user)
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

func updateSimSettings(s addie.SimSettings) {

	design_key, err := db.ReadDesignKey(design.Name, user)
	if err != nil {
		log.Println(err)
		log.Printf("[updateSimSettings] error reading design key")
		return
	}

	err = db.UpdateSimSettings(s, design_key)
	if err != nil {
		log.Println("[updateSimSettings] error updating sim settings")
		log.Println(err)
		return
	}

	simSettings = s
}

func modelId(name string) addie.Id {
	return addie.Id{Name: name, Sys: "", Design: ""}
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

	var new_models []addie.Model
	var changed_models []addie.Model
	var changed_model_oids []addie.Id

	var placeModel = func(oid string, m addie.Model) {

		_, ok := userModels[oid]
		if !ok {
			new_models = append(new_models, m)
		} else {
			changed_models = append(changed_models, m)
			changed_model_oids = append(changed_model_oids, modelId(oid))
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
			var p addie.Phyo
			err := json.Unmarshal(u.Element, &p)
			if err != nil {
				log.Println("unable to unmarshal phyo")
			}
			place(u.OID, p)
		case "Model":
			var m addie.Model
			err := json.Unmarshal(u.Element, &m)
			if err != nil {
				log.Println("unable to unmarshal model")
			}
			placeModel(u.OID.Name, m)
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
		case "SimSettings":
			var s addie.SimSettings
			err := json.Unmarshal(u.Element, &s)
			if err != nil {
				log.Println(err)
				log.Println("unable to unmarshal SimSettings")
				return
			}
			updateSimSettings(s)
		default:
			log.Println("unkown element type: ", u.Type)
		}

	}

	for i, u := range changed_models {
		dbUpdate(changed_model_oids[i], u)
		userModels[u.Name] = u
	}
	for _, c := range new_models {
		dbCreate(c)
		userModels[c.Name] = c
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

	mls, err := db.ReadUserModels(user)
	if err != nil {
		log.Println(err)
		return fmt.Errorf("failed to read user models")
	}

	userModels = make(map[string]addie.Model)

	for _, v := range mls {
		userModels[v.Name] = v
	}

	design_key, err := db.ReadDesignKey(dsg.Name, user)
	if err != nil {
		log.Println(err)
		return fmt.Errorf("error reading design key")
	}

	ss, err := db.ReadSimSettingsByDesignId(design_key)
	if err != nil {
		log.Println(err)
		return fmt.Errorf("could not read simulation settings")
	}

	simSettings = *ss

	return nil
}

type JsonModel struct {
	Name        string            `json:"name"`
	Elements    []TypeWrapper     `json:"elements"`
	Models      []addie.Model     `json:"models"`
	SimSettings addie.SimSettings `json:"simSettings"`
}

func modelJson() ([]byte, error) {

	var mdl JsonModel
	mdl.Name = design.Name

	mdl.Elements = make([]TypeWrapper, len(design.Elements))
	mdl.Models = make([]addie.Model, len(userModels))

	i := 0
	for _, v := range design.Elements {
		mdl.Elements[i] = typeWrap(v)
		i++
	}

	i = 0
	for _, v := range userModels {
		mdl.Models[i] = v
		i++
	}

	mdl.SimSettings = simSettings

	_json, err := json.Marshal(mdl)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("Failed to marshal element array to JSON")
	}

	return _json, nil

}

func userDir() string {
	return "/cypress/" + user
}

func simFileName() string {
	return userDir() + "/" + design.Name + ".cys"
}

func topdlFileName() string {
	return userDir() + "/" + design.Name + ".topdl"
}

func compileSim() {

	models := make([]addie.Model, len(userModels))
	i := 0
	for _, v := range userModels {
		models[i] = v
		i++
	}
	src := sim.GenerateSource(&design, models)
	ioutil.WriteFile(simFileName(), []byte(src), 0644)

	cmd := exec.Command("cyc", simFileName())
	cmd.Dir = userDir()
	outp, err := cmd.Output()
	if err != nil {
		log.Println("could not execute cyc")
		log.Println(err)
	}

	log.Println("cyc returned:")
	log.Println(string(outp))

	cmd = exec.Command("./build_rcomp.sh")
	cmd.Dir = userDir() + "/" + design.Name + ".cypk"
	outp, err = cmd.Output()
	if err != nil {
		log.Println("could not build simulation")
		log.Println(err)
	}
}

func compileTopDL() {

	xp := deter.DesignTopDL(&design)
	topdl, err := xml.MarshalIndent(xp, "  ", "  ")
	if err != nil {
		log.Println(err)
		return
	}

	ioutil.WriteFile(topdlFileName(), topdl, 0644)

}

func onCompile(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	log.Println("addie compiling design")
	w.Write([]byte("ok"))
	compileSim()
	compileTopDL()
}

func onRun(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	log.Println("addie running experiment")

	cmd := exec.Command("./rcomp0",
		strconv.FormatFloat(simSettings.Begin, 'e', -1, 64),
		strconv.FormatFloat(simSettings.End, 'e', -1, 64),
		strconv.FormatFloat(simSettings.MaxStep, 'e', -1, 64))
	cmd.Dir = userDir() + "/" + design.Name + ".cypk"
	_, err := cmd.Output()
	if err != nil {
		log.Println("could not run simulation")
		log.Println(err)
	}

	w.Write([]byte("ok"))
}

func onRawData(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	log.Println("getting raw data")

	data, err := ioutil.ReadFile(userDir() + "/" + design.Name + ".cypk/cnode0.results")
	if err != nil {
		log.Println("could not read results")
		log.Println(err)
		w.WriteHeader(500)
		return
	}

	w.Write([]byte(data))
}

func listen() {

	router := httprouter.New()
	router.POST("/"+design.Name+"/design/update", onUpdate)
	router.GET("/"+design.Name+"/design/read", onRead)
	router.GET("/"+design.Name+"/design/compile", onCompile)
	router.GET("/"+design.Name+"/design/run", onRun)
	router.GET("/"+design.Name+"/analyze/rawData", onRawData)

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
