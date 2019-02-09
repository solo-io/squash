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

		// compose an address with the port
		addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("localhost:%v", port))
		Expect(err).To(BeNil())

		// ensure that the port is actually open
		listener, err := net.ListenTCP("tcp", addr)
		Expect(err).To(BeNil())

		// cleanup
		err = listener.Close()
		Expect(err).To(BeNil())
	})
})
