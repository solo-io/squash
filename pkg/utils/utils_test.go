package utils_test

import (
	"fmt"
	"net"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/squash/pkg/utils"
)

var _ = Describe("get a free port", func() {
	It("port should be free", func() {
		port := 0
		err := utils.FindAnyFreePort(&port)
		Expect(err).To(BeNil())

		listener := listenOnPort(port, true)

		// cleanup
		err = listener.Close()
		Expect(err).To(BeNil())
	})
})

var _ = Describe("check status of port", func() {
	It("should handle free and non-free ports", func() {
		port := 0
		err := utils.FindAnyFreePort(&port)
		Expect(err).To(BeNil())
		listener := &net.TCPListener{}

		// should not error on free port
		err = utils.ExpectPortToBeFree(port)
		Expect(err).To(BeNil())
		listener = listenOnPort(port, true)

		// should error on non-free port
		err = utils.ExpectPortToBeFree(port)
		Expect(err).NotTo(BeNil())

		// cleanup
		err = listener.Close()
		Expect(err).To(BeNil())
	})
})

func listenOnPort(port int, shouldPass bool) *net.TCPListener {

	// compose an address with the port
	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("localhost:%v", port))
	Expect(err).To(BeNil())

	// ensure that the port is actually open
	listener, err := net.ListenTCP("tcp", addr)
	if shouldPass {
		Expect(err).To(BeNil())
	} else {
		Expect(err).NotTo(BeNil())
	}

	return listener
}
