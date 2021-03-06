package main

import (
	"addie"
	"addie/db"
	"addie/deter"
	"addie/protocol"
	"addie/sema"
	"addie/sim"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/cycps/xptools/dnsc"
	"github.com/cycps/xptools/routec"
	"github.com/deter-project/go-spi/spi"
	"github.com/julienschmidt/httprouter"
	"github.com/satori/go.uuid"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"reflect"
	"strconv"
	"strings"
	"text/template"
	"time"
)

var design addie.Design
var userModels = make(map[string]addie.Model)
var simSettings addie.SimSettings
var cypdir = os.ExpandEnv("/cypress")
var user = ""
var kryClusterSize = 1
var gopath = os.Getenv("GOPATH")

func main() {

	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "usage: addie <user id> <design id>\n")
		os.Exit(1)
	}

	err := os.MkdirAll("/var/log/addie", 0755)
	if err != nil {
		log.Println("could not create log folder")
		os.Exit(1)
	}
	f, err := os.OpenFile("/var/log/addie/"+os.Args[2]+".log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Println("could not open log file")
		os.Exit(1)
	}
	defer f.Close()
	log.SetOutput(f)

	user = os.Args[1]
	err = loadSpiCert()
	if err != nil {
		log.Println("could not load spi cert!")
		os.Exit(1)
	}

	loadDesign(os.Args[2])
	checkUserDir()
	loadUserModels()
	listen()
}

func loadSpiCert() error {

	cert, err := ioutil.ReadFile("/cypress/" + user + "/spi.cert")
	if err != nil {
		log.Println(err)
		return fmt.Errorf("spi cert read failure")
	}

	err = spi.SetCertificate(cert)
	if err != nil {
		log.Println(err)
		return fmt.Errorf("failed to set go-spi session cert")
	}

	return nil

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
	var modelKillList []string

	var placeModel = func(oid string, m addie.Model) {

		if m.Name != oid {
			modelKillList = append(modelKillList, oid)
		}
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

	for _, k := range modelKillList {
		delete(userModels, k)
	}

	//log.Println("\n", design.String())

	//send response
	w.WriteHeader(http.StatusOK)

}

//TODO: as it sits, this function relies on the client to provide a consistent
//set of elements to delete, so we are garbage in garbage out essentially. This
//is probably not a good policy. For example if a client deletes a bunch of nodes
//but not the links that they are connected to, meyham will follow.
func onDelete(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	log.Println("addie got delete request")

	msg := new(protocol.Delete)
	err := protocol.Unpack(r, msg)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	nodes := make(map[addie.Id]addie.Identify)
	links := make(map[addie.Id]addie.Link)
	plinks := make(map[addie.Id]addie.Plink)

	for _, d := range msg.Elements {

		switch d.Type {
		case "Computer":
			var c addie.Computer
			err := json.Unmarshal(d.Element, &c)
			if err != nil {
				log.Println("unable to unmarshal " + d.Type)
			}
			log.Printf("deleting %s %v", d.Type, c.Id)
			nodes[c.Id] = c
		case "Router":
			var r addie.Router
			err := json.Unmarshal(d.Element, &r)
			if err != nil {
				log.Println("unable to unmarshal " + d.Type)
			}
			log.Printf("deleting %s %v", d.Type, r.Id)
			nodes[r.Id] = r
		case "Switch":
			var s addie.Switch
			err := json.Unmarshal(d.Element, &s)
			if err != nil {
				log.Println("unable to unmarshal " + d.Type)
			}
			log.Printf("deleting %s %v", d.Type, s.Id)
			nodes[s.Id] = s
		case "Sax":
			var s addie.Sax
			err := json.Unmarshal(d.Element, &s)
			if err != nil {
				log.Println("unable to unmarshal " + d.Type)
			}
			log.Printf("deleting %s %v", d.Type, s.Id)
			nodes[s.Id] = s
		case "Phyo":
			var p addie.Phyo
			err := json.Unmarshal(d.Element, &p)
			if err != nil {
				log.Println("unable to unmarshal " + d.Type)
			}
			log.Printf("deleting %s %v", d.Type, p.Id)
			nodes[p.Id] = p
		case "Link":
			var l addie.Link
			err := json.Unmarshal(d.Element, &l)
			if err != nil {
				log.Println("unable to unmarshal " + d.Type)
			}
			log.Printf("deleting %s %v", d.Type, l.Id)
			links[l.Id] = l
		case "Plink":
			var p addie.Plink
			err := json.Unmarshal(d.Element, &p)
			if err != nil {
				log.Println("unable to unmarshal " + d.Type)
			}
			log.Printf("deleting %s %v", d.Type, p.Id)
			plinks[p.Id] = p
		}

	}

	for _, n := range nodes {

		db.DeleteId(n.Identify(), user)
		delete(design.Elements, n.Identify())

	}

	for _, l := range links {

		db.DeleteId(l.Identify(), user)
		db.DeleteInterface(l.Endpoints[0], user)
		db.DeleteInterface(l.Endpoints[1], user)
		//todo kill the interface on the in-memory model
		delete(design.Elements, l.Identify())

	}

	for _, p := range plinks {

		db.DeleteId(p.Id, user)
		delete(design.Elements, p.Id)

	}

}

func onRead(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	json, err := modelJson()

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

func generateDnsServerConfig() error {

	var spec dnsc.ServerSpec

	spec.Xpname = design.Name
	spec.FQDN = "dns." + spec.Xpname + ".cypress.net"
	spec.Addr = "2001:cc:0:0::d47"

	for _, e := range design.Elements {
		switch e.(type) {
		case addie.Computer:
			c := e.(addie.Computer)
			cs := dnsc.ClientSpec{}
			cs.FQDN = fmt.Sprintf("%s.%s.cypress.net", c.Name, design.Name)
			key, err := dnsc.Keygen(cs.FQDN)
			if err != nil {
				return err
			}
			cs.Key = key
			spec.Clients = append(spec.Clients, cs)
		case addie.Sax:
			s := e.(addie.Sax)
			cs := dnsc.ClientSpec{}
			cs.FQDN = fmt.Sprintf("%s.%s.cypress.net", s.Name, design.Name)
			key, err := dnsc.Keygen(cs.FQDN)
			if err != nil {
				return err
			}
			cs.Key = key
			spec.Clients = append(spec.Clients, cs)
		}
	}

	js, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		log.Println(err)
		return fmt.Errorf("Failed to marshal dns server spec")
	}

	dir := userDir() + "/" + design.Name + ".dns"
	os.MkdirAll(dir, 0755)
	err = ioutil.WriteFile(dir+"/server_spec.json", js, 0644)
	if err != nil {
		log.Println(err)
		return fmt.Errorf("Failed to write dns spec file")
	}

	for _, cs := range spec.Clients {
		js, err = json.MarshalIndent(cs, "", "  ")
		if err != nil {
			log.Println(err)
			return fmt.Errorf("Failed to marshal dns client spec")
		}
		shortname := strings.Split(cs.FQDN, ".")[0]
		err = ioutil.WriteFile(dir+"/"+shortname+".json", js, 0644)
		if err != nil {
			log.Println(err)
			return fmt.Errorf("Failed to write dns client file")
		}
	}

	return nil
}

func switchConnections(sw addie.Switch) []addie.Id {

	var xs []addie.Id

	for _, e := range design.Elements {
		switch e.(type) {
		case addie.Link:
			l := e.(addie.Link)
			if l.Endpoints[0].Id == sw.Id {
				xs = append(xs, l.Endpoints[1].Id)
			}
			if l.Endpoints[1].Id == sw.Id {
				xs = append(xs, l.Endpoints[0].Id)
			}
		}
	}

	return xs
}

func updateCharts(charts map[string]routec.RouterChart,
	r addie.Router, x addie.Identify) {

	ch, ok := charts[r.Id.String()]
	if !ok {
		ch.Id = uint32(len(charts) + 1)
	}

	switch x.(type) {
	case addie.Router:
		r := x.(addie.Router)
		ch.PeerHosts = append(ch.PeerHosts, r.Name)
	case addie.Computer:
		c := x.(addie.Computer)
		ch.DownstreamHosts = append(ch.DownstreamHosts, c.Name)
	case addie.Sax:
		s := x.(addie.Sax)
		ch.DownstreamHosts = append(ch.DownstreamHosts, s.Name)
	case addie.Switch:
		s := x.(addie.Switch)
		sx := switchConnections(s)
		for _, i := range sx {
			if i == r.Id {
				continue
			}
			ch.DownstreamHosts = append(ch.DownstreamHosts, i.Name)
			break
		}

	}

	charts[r.Id.String()] = ch
}

func generateRouterConfigs() error {

	charts := make(map[string]routec.RouterChart)

	for _, e := range design.Elements {
		switch e.(type) {
		case addie.Link:
			l := e.(addie.Link)
			a := design.Elements[l.Endpoints[0].Id]
			b := design.Elements[l.Endpoints[1].Id]
			switch a.(type) {
			case addie.Router:
				r := a.(addie.Router)
				updateCharts(charts, r, b)
			}
			switch b.(type) {
			case addie.Router:
				r := b.(addie.Router)
				updateCharts(charts, r, a)
			}
		}
	}

	dir := userDir() + "/" + design.Name + ".route"
	os.MkdirAll(dir, 0755)
	for n, c := range charts {
		js, err := json.MarshalIndent(c, "", "  ")
		if err != nil {
			log.Println(err)
			return fmt.Errorf("Failed to marshal router chart")
		}
		err = ioutil.WriteFile(dir+"/"+n+".rc.json", js, 0644)
		if err != nil {
			log.Println(err)
			fmt.Errorf("Failed to write router chart file")
		}
	}

	return nil
}

func onCompile(w http.ResponseWriter, r *http.Request,
	ps httprouter.Params) {
	log.Println("addie compiling design")

	log.Println("checking design ...")
	diagnostics := sema.Check(&design)
	log.Println("OK")

	if !diagnostics.Fatal() {
		log.Println("compiling PnetDL ...")
		compileSim()
		log.Println("OK")

		log.Println("compiling TopDL ...")
		compileTopDL()
		log.Println("OK")

		log.Println("building dns configs ...")
		err := generateDnsServerConfig()
		if err != nil {
			log.Printf("Fail: %v\n", err)
		} else {
			log.Println("OK")
		}

		log.Println("building quagga configs ...")
		err = generateRouterConfigs()
		if err != nil {
			log.Printf("fail: %v\n", err)
		} else {
			log.Println("OK")
		}
	}

	json, err := json.Marshal(diagnostics)
	if err != nil {
		log.Println(err)
		w.WriteHeader(500)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}

func runSim() {

	log.Println("addie running simulation")

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

}

func runComputerCode(c addie.Computer) {

}

type ConfigData struct {
	User, XP, Id string
}

func syncFileTBNode(file, target string) error {
	return _syncFile(user, file, target, "/cypress/scripts/rsync_it_testbed.sh")
}

func syncFile(file, target string) error {
	return _syncFile("root", file, target, "/cypress/scripts/rsync_it.sh")
}

func _syncFile(u, file, target, script string) error {

	cmdt := []string{script, u, file, target}

	cmd := exec.Command(script, cmdt[1:]...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Println(err)
		log.Println(string(out))
		return fmt.Errorf("erorr syncing %s", file)
	}

	return nil
}

func cypressResource(rel_path string) ([]byte, error) {

	abs_path := gopath + "/src/github.com/cycps/" + rel_path
	bs, err := ioutil.ReadFile(abs_path)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("Could not read cypress resource '%s'", abs_path)
	}
	return bs, nil
}

func runCypressTemplate(rel_path, outfile string, data interface{}) error {

	tps, err := cypressResource(rel_path)
	if err != nil {
		return err
	}

	tp, err := template.New("tmpl").Parse(string(tps))
	if err != nil {
		log.Println(err)
		return fmt.Errorf("Could not parse cypress template '%s'", rel_path)
	}
	of, err := os.Create(outfile)
	if err != nil {
		log.Println(err)
		return fmt.Errorf("Could not create template output file '%s'", outfile)
	}
	err = tp.Execute(of, data)
	if err != nil {
		log.Println(err)
		return fmt.Errorf("Error executing cypress template '%s' with data %v",
			rel_path, data)
	}
	of.Close()

	return nil

}

func syncTemplateTBNode(template_fn, target string, data interface{}) error {
	out := "/tmp/" + uuid.NewV4().String() + "_" + path.Base(template_fn)
	err := runCypressTemplate(template_fn, out, data)
	err = syncFileTBNode(out, target)
	if err != nil {
		log.Println(err)
		return fmt.Errorf("sync for do-quagga-config.sh failed")
	}
	os.Remove(out)
	return nil
}

func syncRouterFiles(r *addie.Router) error {
	log.Printf("syncing router configuation files for %s...", r.Id)

	cdata := ConfigData{
		User: user,
		XP:   design.Name,
		Id:   r.Id.String(),
	}

	//sync quagga config script
	err := syncTemplateTBNode(
		"xptools/routec/quagga_config/do-quagga-config.sh",
		deterName(r.Id)+":/tmp/cypress/",
		cdata,
	)
	if err != nil {
		return err
	}

	//sync router chart
	err = syncFileTBNode(
		"/cypress/"+user+"/"+design.Name+".route/"+r.Id.String()+".rc.json",
		deterName(r.Id)+":/tmp/cypress/",
	)
	if err != nil {
		return err
	}

	//sync router node bootstrap script
	err = syncFileTBNode(
		gopath+"/src/github.com/cycps/xptools/routec/quagga_config/bootstrap.sh",
		deterName(r.Id)+":/tmp/cypress/",
	)
	if err != nil {
		return err
	}
	log.Println("configuration files synced")

	return nil
}

func deterName(id addie.Id) string {

	return fmt.Sprintf("%s.%s-%s.cypress", id.Name, user, design.Name)

}

func configRouter(r *addie.Router) error {

	err := syncRouterFiles(r)
	if err != nil {
		return err
	}

	log.Printf("bootstrapping %s...", r.Id)
	err = sshCmdTB(deterName(r.Id), "/tmp/cypress/bootstrap.sh")
	if err != nil {
		return err
	}
	log.Println("ok")

	return nil

}

func configRouters() error {
	log.Println("configuring routers")

	rs := design.Routers()
	ch := make(chan error, len(rs))

	for _, r := range rs {
		go func(r *addie.Router) {
			err := configRouter(r)
			if err != nil {
				log.Println(err)
				err = fmt.Errorf("Error configuring router %s", r.Id)
			}
			ch <- err
		}(r)
	}

	failed := false
	for err := range ch {
		if err != nil {
			log.Println(err)
			failed = true
		}
	}

	if failed {
		return fmt.Errorf("Failed to configure routers")
	}

	/*
		for _, e := range design.Elements {
			switch e.(type) {
			case addie.Router:
				r := e.(addie.Router)
				go func() {
					err := configRouter(&r)
					if err != nil {
						log.Println(err)
						log.Printf("Error configuring router %s", r.Id)
					}
				}()
			}
		}
	*/

	log.Println("routers configured")
	return nil

}

func sshCmdTB(node, rcmd string) error {

	args := []string{"ssh", "-A", "-t", user + "@users.isi.deterlab.net",
		"ssh", "-A", node, rcmd}

	cmd := exec.Command("ssh", args[1:]...)
	err := cmd.Run()
	if err != nil {
		log.Println(err)
		return fmt.Errorf("Could not execute ssh command on deter node")
	}

	return nil
}

func sshCmd(node, rcmd string) error {
	args := []string{"ssh", "-o", "StrictHostKeyChecking=no", node, rcmd}

	cmd := exec.Command("ssh", args[1:]...)
	err := cmd.Run()
	if err != nil {
		log.Println(err)
		return fmt.Errorf("Could not execute ssh command `%s` on %s", rcmd, node)
	}

	return nil
}

func onRun(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	log.Println("addie running experiment")

	err := configRouters()
	if err != nil {
		log.Println(err)
	}

	//launchRouters()
	//launchSax()
	//launchSoftware()
	//launchSim()

	//## old ##
	////runSim()

	/*

		for _, e := range design.Elements {
			switch e.(type) {
			case addie.Computer:
				c := e.(addie.Computer)
				args := strings.Split(strings.TrimPrefix(c.SSHC(user, design.Name), "ssh"), " ")
				args = append(args, []string{"touch", "addieWasHere"}...)
				log.Printf("%s: %s", c.Name, c.SSHC(user, design.Name))
				log.Println(args[1:])
				log.Println(len(args[1:]))

				cmd := exec.Command("ssh", args[1:]...)
				err := cmd.Run()
				if err != nil {
					log.Println("addie could not touch " + c.Name)
					log.Println(err)
				}

			case addie.Router:
				r := e.(addie.Router)
				log.Printf("%s: %s", r.Name, r.SSHC(user, design.Name))
			case addie.Sax:
				s := e.(addie.Sax)
				log.Printf("%s: %s", s.Name, s.SSHC(user, design.Name))
			}
		}

		for i := 0; i < kryClusterSize; i++ {

			ksshc := fmt.Sprintf("ssh -A -t %s@users.isi.deterlab.net ssh -A kry%d.%s-%s.cypress",
				user, i, user, design.Name)

			log.Printf("kry%d: %s", i, ksshc)
		}
	*/

	w.Write([]byte("ok"))
}
func onDeMaterialize(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	log.Println("dematerializing")
	rr, err := spi.RemoveRealization(user + ":" + design.Name + "-cypress:cypress")
	if err != nil {
		log.Println("spi call to remove realization failed")
		log.Println(err)
		w.WriteHeader(500)
		return
	}
	if rr.Return != true {
		log.Println("the badness happened with the spi call to remove realization")
		w.WriteHeader(500)
		return
	}

	dmat := false
	for dmat == false {
		// check dematerialization status
		time.Sleep(1 * time.Second)
		log.Println("checking realization shutdown")
		ms, err := spi.ViewRealizations(user, ".*"+design.Name+".*")
		if err != nil {
			log.Println("[dmat] spi call to get realizations failed")
			log.Println(err)
			w.WriteHeader(500)
			return
		}
		if len(ms.Return) == 0 {
			log.Println("dematerialization finished")
			dmat = true
		}
	}

	log.Println("removing experiment")
	rx, err := spi.RemoveExperiment(user + ":" + design.Name)
	if err != nil {
		log.Println("spi call to remove experiment failed")
		log.Println(err)
		w.WriteHeader(500)
		return
	}
	if rx.Return != true {
		log.Println("the badness happened with the spi call to remove experiment")
		w.WriteHeader(500)
		return
	}

}

func onMaterialize(w http.ResponseWriter, r *http.Request,
	ps httprouter.Params) {
	log.Println("addie materializing experiment")

	topDL, err := ioutil.ReadFile(topdlFileName())
	if err != nil {
		log.Println("unable to read topdl file :" + topdlFileName())
		w.WriteHeader(500)
		return
	}

	// Get active realizations
	log.Println("checking to see if we are already materialized")
	ms, err := spi.ViewRealizations(user, ".*"+design.Name+".*")
	if err != nil {
		log.Println("spi call to get realizations failed")
		log.Println(err)
		w.WriteHeader(500)
		return
	}
	if len(ms.Return) > 0 {
		msg := "the design is already materialize, de-materialize first to re-materialize"
		log.Println(msg)
		w.WriteHeader(400)
		fmt.Fprintf(w, msg)
		return
	}

	//++ Create
	log.Println("creating experiment")
	response, err := spi.CreateExperiment(user+":"+design.Name, user, string(topDL),
		false)
	if err != nil {
		log.Println("spi call to create experiment failed")
		log.Println(err)
		w.WriteHeader(500)
		return
	}
	if response != nil && response.Return != true {
		log.Println("the badness happend with the spi call to create experiment")
		w.WriteHeader(500)
		return
	}

	//~~ Realize
	log.Println("realizing experiment")
	mresponse, err := spi.RealizeExperiment(user+":"+design.Name, "cypress:cypress", user)
	if err != nil {
		log.Println("spi call to realize experiment failed")
		log.Println(err)
		w.WriteHeader(500)
		return
	}

	log.Println("realization response")
	log.Println(mresponse.Return)

	w.Write([]byte("ok"))
}

func onMstate(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	log.Println("addie fetching materialization state")

	ms, err := spi.ViewRealizations(user, ".*"+design.Name+".*")
	if err != nil {
		log.Println("spi call to get realizations failed")
		log.Println(err)
		w.WriteHeader(500)
		return
	}

	var js []byte
	if len(ms.Return) > 0 {
		js, err = json.Marshal(ms.Return[0])
	} else {
		js = []byte("[]")
	}

	if err != nil {
		log.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)

}

func onModelIco(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	log.Println("addie receiving model icon")
	err := r.ParseMultipartForm(50 * 1024 * 1024)
	if err != nil {
		log.Println("parse form failed")
		log.Println(err)
	}
	mdl := r.MultipartForm.Value["modelName"][0]
	log.Printf("model: %s", mdl)
	log.Printf("file: %s", r.MultipartForm.File["modelIco"][0].Filename)

	f, err := r.MultipartForm.File["modelIco"][0].Open()
	if err != nil {
		log.Println("error opening icon file")
		log.Println(err)
	}

	content, err := ioutil.ReadAll(f)
	if err != nil {
		log.Println("could not read file")
		log.Println(err)
	}

	log.Printf("icon file size: %d", len(content))

	uuid := uuid.NewV4()
	fn := "/tmp/" + uuid.String() + "_" + mdl + ".png"
	m, ok := userModels[mdl]
	if ok {
		m.Icon = fn
		userModels[mdl] = m
		db.UpdateModel(m.Name, m, user)
	}

	log.Println("saving icon " + fn)
	err = ioutil.WriteFile(fn, content, 0644)
	if err != nil {
		log.Println(err)
		log.Println("failed to save icon file")
	}

	err = sshCmd("web", "mkdir -p /cypress/web/ico/"+user)
	if err != nil {
		log.Println(err)
		log.Println("could not create user model icon directory on webserver")
	}

	err = syncFile(fn, "web:/cypress/web/ico/"+user+"/"+mdl+".png")
	if err != nil {
		log.Println(err)
		log.Println("could not sync model icon to user icon dir on webserver")
	}

	err = os.Remove(fn)
	if err != nil {
		log.Printf("could not remove temporary icon file %s", fn)
	}

	//log.Println(r.MultipartForm)

}

func onRawData(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	log.Println("getting raw data")

	data, err := ioutil.ReadFile(userDir() + "/" + design.Name +
		".cypk/cnode0.results")
	if err != nil {
		log.Println("could not read results")
		log.Println(err)
		w.WriteHeader(500)
		return
	}

	w.Write([]byte(data))
}

//TODO ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
//The way to do this is to have 1 addie instance and run (user,design) handler
//pairs as goroutines
func listen() {

	router := httprouter.New()
	router.POST("/"+design.Name+"/design/update", onUpdate)
	router.POST("/"+design.Name+"/design/delete", onDelete)
	router.GET("/"+design.Name+"/design/read", onRead)
	router.GET("/"+design.Name+"/design/compile", onCompile)
	router.GET("/"+design.Name+"/design/run", onRun)
	router.GET("/"+design.Name+"/design/materialize", onMaterialize)
	router.GET("/"+design.Name+"/design/dematerialize", onDeMaterialize)
	router.POST("/"+design.Name+"/design/modelIco", onModelIco)
	router.GET("/"+design.Name+"/design/mstate", onMstate)
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
