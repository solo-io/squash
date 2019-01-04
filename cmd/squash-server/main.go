package main

import (
	"fmt"

	"github.com/solo-io/squash/pkg/cmd/server"
)

// this program is for testing only, at this point
func main() {
	if err := server.ServerCmd(); err != nil {
		fmt.Println(err)
	}
}
