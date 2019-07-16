package main

import (
	"net/http"

	"github.com/solo-io/squash/pkg/urllogger"
)

// Helper program to gather logs from pods that crash before you can read their logs
func main() {
	http.HandleFunc("/", urllogger.BasicHandlerFunction)
	http.ListenAndServe(":8080", nil)
}
