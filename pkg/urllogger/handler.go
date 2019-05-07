package urllogger

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

func BasicHandlerFunction(w http.ResponseWriter, r *http.Request) {
	b := []byte{}
	bb := bytes.NewBuffer(b)
	if _, err := io.Copy(bb, r.Body); err != nil {
		printErr(err)
	}
	if err := r.Body.Close(); err != nil {
		printErr(err)
	}
	fmt.Print(bb.String())
}

func printErr(e error) {
	fmt.Printf("Error in spooler: %v\n", e)
}
