package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

var ServiceToCall = "example-service2"

func main() {

	fmt.Println("starting app")
	potentialservice2 := os.Getenv("SERVICE2_URL")
	if potentialservice2 != "" {
		ServiceToCall = potentialservice2
	}

	http.HandleFunc("/calc", handler)
	http.HandleFunc("/", view)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func view(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello")
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello again")
}
