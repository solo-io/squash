package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"github.com/solo-io/go-utils/contextutils"
)

var ServiceToCall = "example-service2"

const form = `
<html>
<head>
<!-- Latest compiled and minified CSS -->
<link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.7/css/bootstrap.min.css" integrity="sha384-BVYiiSIFeK1dGmJRAkycuHAHRg32OmUcww7on3RYdg4Va+PmSTsz/K68vbdEjh4u" crossorigin="anonymous">

<!-- Optional theme -->
<link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.7/css/bootstrap-theme.min.css" integrity="sha384-rHyoN1iRsVXV4nD0JutlnGaslCJuC7uwjduW9SVrLvRYooPp2bWYgmgJQIXwl/Sp" crossorigin="anonymous">

<!-- Latest compiled and minified JavaScript -->
<script src="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.7/js/bootstrap.min.js" integrity="sha384-Tc5IQib027qvyjSMfHjOMaLkfuWVxZxUPnCJA7l2mCWNIpG9mGCD8wGNIcPD7Txa" crossorigin="anonymous"></script>
</head>
	<body style="background-color:black;color:white;">

		<main class="col-md-6 col-md-offset-3">
			<form action="calc" method="POST">

				<div class="form-group" style="text-align:center;">
					<BR><BR><H1>The great adding and subtracting app</H1><BR><BR>
				</div>
				<div class="form-group">
					<label for="op1">Operand 1:</label>
					<input type="text" name="op1" id="op1" value="%v" class="form-control">
				</div>
				<div class="form-group">	
					<label for="op2">Operand 2:</label>
					<input type="text" name="op2" id="op2" value="%v" class="form-control">
				</div>
				<div class="form-group">	
					<input type="radio" name="optype" id="optypeadd" checked="checked" value="add"> <label for="optypeadd">add</label>
					<input type="radio" name="optype" id="optype-subtract" value="subtract"> <label for="optype-subtract"> subtract</label>
				</div>
				<div class="form-group">	
					<input type="submit" value="Calculate" style="color: black; background-color: white; font-weight: bold;">
				</div>
			</form>

			<div>
				<H1><BR><BR>
					Result: %v %v %v %v %v
				</H1>
			</div>

		</main>
	</body>
</html>
`

func main() {

	potentialservice2 := os.Getenv("SERVICE2_URL")
	if potentialservice2 != "" {
		ServiceToCall = potentialservice2
	}

	http.HandleFunc("/calc", handler)
	http.HandleFunc("/", view)

	contextutils.LoggerFrom(context.TODO()).Fatal(http.ListenAndServe(":8080", nil))
}

func view(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, form, "", "", "", "", "", "", "")
}

type Calculator struct {
	Op1, Op2 int
	IsAdd    bool
}

func handler(w http.ResponseWriter, r *http.Request) {
	optype := r.FormValue("optype")
	op1, _ := strconv.Atoi(r.FormValue("op1"))
	op2, _ := strconv.Atoi(r.FormValue("op2"))

	isadd := optype == "add"
	calc := Calculator{
		Op1: op1, Op2: op2, IsAdd: isadd,
	}

	jsoncalc, _ := json.Marshal(calc)
	var jsoncalcreader io.Reader = bytes.NewReader(jsoncalc)
	resp, err := http.Post("http://"+ServiceToCall+"/calculate", "application/json", jsoncalcreader)
	if err != nil {
		fmt.Fprintf(w, form, "", "", "", "", "", "", err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(w, form, "", "", "", "", "", "", resp.Status)
		return
	}

	body, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	calcresult, _ := strconv.Atoi(string(body))

	if optype == "add" {
		fmt.Fprintf(w, form, op1, op2, op1, "+", op2, "=", calcresult)
	} else {
		fmt.Fprintf(w, form, op1, op2, op1, "-", op2, "=", calcresult)
	}
}
