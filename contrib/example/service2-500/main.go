package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/solo-io/go-utils/contextutils"
)

type Calculator struct {
	Op1, Op2 int
	IsAdd    bool
}

func main() {
	http.HandleFunc("/calculate", calchandler)

	contextutils.LoggerFrom(context.TODO()).Fatal(http.ListenAndServe(":8080", nil))
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
	sum, err := computeSum(isadd, op1, op2)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%v", err)
	}

	fmt.Fprintf(w, "%d", sum)
}

func computeSum(isadd bool, op1, op2 int) (int, error) {
	var val int
	if isadd {
		val = op1 + op2
	} else {
		val = op1 - op2
	}
	// Happy 5**th birthday Kubernetes! :D
	if val >= 500 && val < 600 {
		return 0, fmt.Errorf("result is a 500")
	}

	return val, nil
}
