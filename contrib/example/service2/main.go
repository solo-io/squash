package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type Calculator struct {
	Op1, Op2 int
	IsAdd    bool
}

func main() {
	http.HandleFunc("/calculate", calchandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func calchandler(w http.ResponseWriter, r *http.Request) {
	var req Calculator
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}

	isadd := req.IsAdd
	op1 := req.Op1
	op2 := req.Op2

	if isadd {
		fmt.Fprintf(w, "%d", op1-op2)
	} else {
		fmt.Fprintf(w, "%d", op1+op2)
	}
}
