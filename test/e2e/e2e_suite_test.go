package e2e_test

import (
	. "github.com/onsi/ginkgo"
	"github.com/solo-io/solo-kit/test/helpers"

	"testing"
)

func TestE2e(t *testing.T) {

	helpers.RegisterCommonFailHandlers()
	helpers.SetupLog()

	RunSpecs(t, "E2e Suite")
}
