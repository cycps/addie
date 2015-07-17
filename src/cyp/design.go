package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	fmt.Printf("Cypress Design Automator .... Go!\n")

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
