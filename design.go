package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
)

var cypdir = os.ExpandEnv("$HOME/.cypress")
var logfile os.File
var db *sql.DB = nil

func initLogging() {
	err := os.MkdirAll(cypdir+"/log", 0755)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to create directory "+cypdir+"/log\n")
		fmt.Fprintf(os.Stderr, err.Error()+"\n")
		os.Exit(1)
	}
	logfile, err := os.OpenFile(cypdir+"/log/addie.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0755)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to open "+cypdir+"/log/addie.log for writing\n")
		fmt.Fprintf(os.Stderr, err.Error()+"\n")
		os.Exit(1)
	}
	log.SetOutput(io.MultiWriter(logfile, os.Stdout))
}

func init() {
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

func handleRequests() {

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		req.ParseForm()
		log.Printf("%s: %s\n", req.Method, req.URL.Path)
		for k, p := range req.Form {
			log.Printf("\t%s = %v\n", k, p)
		}
		//fmt.Fprintf(w, "Do you know the muffin man?")
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, " { \"result\": \"ok\", \"created\": [ { \"name\": \"abby\", \"sys\": \"\"} ] } ")
	})
	log.Println("listening ...")
	http.ListenAndServe(":8080", mux)

}

func main() {

	defer exit(0)
	catchSignals()

	log.Printf("Cypress Design Automator .... Go!\n")
	if dbConnect() != nil {
		exit(1)
	}
	if dbStats() != nil {
		exit(1)
	}

	handleRequests()

}
