package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

// Helper program to gather logs from pods that crash before you can read their logs
func main() {
	http.HandleFunc("/", func(writer http.ResponseWriter, r *http.Request) {
		b := []byte{}
		bb := bytes.NewBuffer(b)
		io.Copy(bb, r.Body)
		r.Body.Close()
		fmt.Println(bb.String())
	})
	http.ListenAndServe(":8080", nil)
}
