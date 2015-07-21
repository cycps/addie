package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cycps/addie"
	"github.com/julienschmidt/httprouter"
	_ "github.com/lib/pq"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
)

var cypdir = os.ExpandEnv("$HOME/.cypress")
var root addie.System
var logfile os.File
var db *sql.DB = nil

func initLogging() {
	err := os.MkdirAll(cypdir+"/log", 0755)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to create directory "+cypdir+"/log\n")
		fmt.Fprintf(os.Stderr, err.Error()+"\n")
		os.Exit(1)
	}

	logfile, err := os.OpenFile(cypdir+"/log/addie.log",
		os.O_RDWR|os.O_CREATE|os.O_APPEND, 0755)

	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to open "+cypdir+"/log/addie.log for writing\n")
		fmt.Fprintf(os.Stderr, err.Error()+"\n")
		os.Exit(1)
	}
	log.SetOutput(io.MultiWriter(logfile, os.Stdout))
}

func initDesign() {
	initLogging()
}

func exit(exitVal int) {
	log.Println("addie shutting down")
	logfile.Close()
	if r := recover(); r != nil {
		log.Println("shutdown is due to panic, panic info follows")
		panic(r)
	}
	os.Exit(exitVal)
}

func catchSignals() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for _ = range c {
			log.Println("Caught intterrupt signal, cleaning up and shutting down")
			exit(1)
		}
	}()
}

func dbConnect() error {
	log.Printf("Opening connecton to pgdb\n")
	var err error
	db, err = sql.Open("postgres", "postgres://postgres@192.168.1.201/cyp?sslmode=require")
	if err != nil {
		log.Println(err)
		return err
	}

	err = db.Ping()
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func dbStats() error {
	log.Println("Cypress DB stats:")
	rows, err := db.Query("SELECT relname, n_live_tup FROM pg_stat_user_tables")
	if err != nil {
		log.Println(err)
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var relname string
		var n_live_tup int
		if err := rows.Scan(&relname, &n_live_tup); err != nil {
			log.Println(err)
		}
		log.Printf("%s {%d}", relname, n_live_tup)
	}
	return err
}

func bakery(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	log.Println("[bakery][" + r.Method + "] hit")
	fmt.Fprint(w, "Do you know the muffin man?\n")
}

type UpdateMsg struct {
	Computers []addie.Computer
}

type DeleteMsg struct {
	Elements []addie.Element
}

type UpdateResult struct {
	Name string
	Sys  string
}

type FailedUpdateResult struct {
	UpdateResult
	Msg string
}

type AggUpdateResult struct {
	Result  string
	Details string
	Created []UpdateResult
	Updated []UpdateResult
	Deleted []UpdateResult
	Failed  []FailedUpdateResult
}

func writeUpdateResult(w http.ResponseWriter, r AggUpdateResult) {
	bs, err := json.Marshal(r)
	if err != nil {
		log.Println("error marshalling json response")
		log.Println(err)
	} else {
		w.Write(bs)
	}
}

func unpackUpdateMsg(r *http.Request) (*UpdateMsg, error) {
	log.Print("unpacking update message")

	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)

	msg := new(UpdateMsg)
	err := json.Unmarshal(buf.Bytes(), msg)
	if err != nil {
		log.Println(err)
		return nil, errors.New("unable to unpack update mesage\n" + buf.String())
	}
	return msg, nil
}

func unpackDeleteMsg(r *http.Request) (*DeleteMsg, error) {
	log.Print("unpacking delete message")

	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)

	msg := new(DeleteMsg)
	err := json.Unmarshal(buf.Bytes(), msg)
	if err != nil {
		log.Println(err)
		return nil, errors.New("unable to unpack delete message\n" + buf.String())
	}
	return msg, nil
}

func updateResponse(w http.ResponseWriter, r *AggUpdateResult) error {
	log.Println("writing client response")
	bs, err := json.Marshal(r)
	if err != nil {
		log.Println(err)
		return errors.New("error marshalling json response")
	}
	w.Write(bs)
	return nil
}

func dbComputerUpdate(c addie.Computer) error {
	q := fmt.Sprintf(
		"UPDATE computers SET os = '%s', start_script = '%s'"+
			"WHERE name = '%s' AND sys = '%s';",
		c.OS, c.Start_script, c.Name, c.Sys)

	_, err := db.Query(q)
	if err != nil {
		log.Println(err)
		return errors.New("computer update failed")
	}
	return nil
}

func dbComputerInsert(c addie.Computer) error {

	q0 := fmt.Sprintf("INSERT INTO network_hosts VALUES ('%s', '%s');",
		c.Name, c.Sys)

	q1 := fmt.Sprintf("INSERT INTO computers VALUES ('%s', '%s', '%s', '%s');",
		c.Name, c.Sys, c.OS, c.Start_script)

	_, err := db.Query(q0)
	if err != nil {
		log.Println(err)
		return errors.New("network_hosts insert failed")

	}
	_, err = db.Query(q1)
	if err != nil {
		log.Println(err)
		return errors.New("computers insert failed")
	}

	return nil
}

func updateComputer(c addie.Computer, r *AggUpdateResult) error {
	_c := root.FindComputer(c.Element)
	if _c == nil {
		log.Println("Added computer ", c)
		__c := root.AddComputer(c)
		if __c == nil {
			var fur FailedUpdateResult
			fur.Name = c.Name
			fur.Sys = c.Sys
			fur.Msg = "Non-existant system"
			r.Failed = append(r.Failed, fur)
		} else {
			err := dbComputerInsert(c)
			if err != nil {
				log.Println(err)
				return err
			}
			r.Created = append(r.Created, UpdateResult{c.Name, c.Sys})
		}
	} else {
		log.Println("Updated computer ", c)
		*_c = c
		dbComputerUpdate(c)
		r.Updated = append(r.Updated, UpdateResult{c.Name, c.Sys})
	}
	return nil
}

func doUpdate(u *UpdateMsg) (*AggUpdateResult, error) {

	r := new(AggUpdateResult)
	r.Result = "ok"

	for _, c := range u.Computers {
		err := updateComputer(c, r)
		if err != nil {
			r.Result = "failed"
			log.Println(err)
			break
		}
	}

	if len(r.Failed) > 0 {
		r.Result = "failed"
	}

	log.Println(root)

	return r, nil
}

func dbElementDelete(e addie.Element) error {

	q := fmt.Sprintf("DELETE FROM network_hosts WHERE name = '%s' AND sys = '%s';",
		e.Name, e.Sys)
	log.Println(q)

	_, err := db.Query(q)
	if err != nil {
		log.Println(err)
		return errors.New("element delete failed")
	}
	return nil
}

func deleteElement(e addie.Element, r *AggUpdateResult) error {
	err := root.DeleteElement(e)
	if err != nil {
		log.Println(err)
		var fur FailedUpdateResult
		fur.Name = e.Name
		fur.Sys = e.Sys
		fur.Msg = "Non-existant system"
		r.Failed = append(r.Failed, fur)
	} else {
		err := dbElementDelete(e)
		if err != nil {
			log.Println(err)
			return err
		}
		r.Deleted = append(r.Deleted, UpdateResult{e.Name, e.Sys})
	}
	return nil
}

func doDelete(d *DeleteMsg) (*AggUpdateResult, error) {
	r := new(AggUpdateResult)
	r.Result = "ok"

	for _, e := range d.Elements {
		err := deleteElement(e, r)
		if err != nil {
			r.Result = "failed"
			log.Println(err)
			break
		}
	}
	if len(r.Failed) > 0 {
		r.Result = "failed"
	}
	log.Println(root)

	return r, nil
}

func handleDesignUpdate(w http.ResponseWriter, r *http.Request,
	ps httprouter.Params) {
	xpid := ps.ByName("xpid")
	root.Name = xpid
	log.Printf("/design/%s", xpid)
	w.Header().Set("Content-Type", "application/json")

	//unpack message
	in, err := unpackUpdateMsg(r)
	if err != nil {
		log.Println(err)
		writeUpdateResult(w, AggUpdateResult{"failed", "malformed request",
			nil, nil, nil, nil})
		return
	}

	//do the update
	out, err := doUpdate(in)
	if err != nil {
		log.Println(err)
		writeUpdateResult(w, AggUpdateResult{"failed", "persistence error",
			nil, nil, nil, nil})
	}

	//send response
	updateResponse(w, out)
}

func handleDesignDelete(w http.ResponseWriter, r *http.Request,
	ps httprouter.Params) {
	xpid := ps.ByName("xpid")
	log.Printf("/design/%s/delete", xpid)
	w.Header().Set("Content-Type", "application/json")

	//unpack message
	in, err := unpackDeleteMsg(r)
	if err != nil {
		log.Println(err)
		writeUpdateResult(w, AggUpdateResult{"failed", "malformed request",
			nil, nil, nil, nil})
		return
	}

	out, err := doDelete(in)
	if err != nil {
		log.Println(err)
		writeUpdateResult(w, AggUpdateResult{"failed", "persistence error",
			nil, nil, nil, nil})
	}

	updateResponse(w, out)

	/*
		//TODO this is a thermonuclear baseline
		var _rt addie.System
		root = _rt
		root.Name = xpid
		q0 := "DELETE FROM computers ;"
		q1 := "DELETE FROM network_hosts ;"

		db.Query(q0)
		db.Query(q1)

		//TODO hardcode for test baseline
		var res AggUpdateResult
		res.Result = "ok"
		res.Deleted = append(res.Deleted, UpdateResult{"abby", "system47"})

		updateResponse(w, &res)
	*/
}

func handleRequests() {

	router := httprouter.New()
	router.POST("/design/:xpid", handleDesignUpdate)
	router.POST("/design/:xpid/delete", handleDesignDelete)
	router.GET("/bakery", bakery)
	router.POST("/bakery", bakery)

	log.Println("listening ...")
	log.Fatal(
		http.ListenAndServeTLS(":8080", cypdir+"/keys/cert.pem",
			cypdir+"/keys/key.pem", router))

}

func main() {

	defer exit(0)
	initDesign()
	catchSignals()

	log.Printf("Cypress Design Automator .... Go!\n")
	if dbConnect() != nil {
		exit(1)
	}
	/*
		if dbStats() != nil {
			exit(1)
		}
	*/

	handleRequests()

}
