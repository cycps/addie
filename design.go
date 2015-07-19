package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"io"
	"log"
	"net/http"
	"os"
)

func main() {

	cypdir := os.ExpandEnv("$HOME/.cypress")

	err := os.MkdirAll(cypdir+"/log", 0755)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to create directory "+cypdir+"/log\n")
		fmt.Fprintf(os.Stderr, err.Error()+"\n")
		os.Exit(1)
	}
	f, err := os.OpenFile(cypdir+"/log/addie.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0755)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to open "+cypdir+"/log/addie.log for writing\n")
		fmt.Fprintf(os.Stderr, err.Error()+"\n")
		os.Exit(1)
	}
	defer f.Close()
	log.SetOutput(io.MultiWriter(f, os.Stdout))

	log.Printf("Cypress Design Automator .... Go!\n")
	log.Printf("Do you know the muffin man?\n")

	log.Printf("Opening connecton to pgdb\n")
	db, err := sql.Open("postgres", "postgres://postgres@192.168.1.201/cyp?sslmode=require")
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	rows, err := db.Query("SELECT relname, n_live_tup FROM pg_stat_user_tables")
	defer rows.Close()
	for rows.Next() {
		var relname string
		var n_live_tup int
		if err := rows.Scan(&relname, &n_live_tup); err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		log.Printf("%s {%d}", relname, n_live_tup)
	}

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
