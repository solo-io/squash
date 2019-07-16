package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/solo-io/go-utils/contextutils"
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

	contextutils.LoggerFrom(context.TODO()).Fatal(http.ListenAndServe(":8080", nil))
}

func view(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello")
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello again")
}
