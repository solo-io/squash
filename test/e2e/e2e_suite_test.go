package e2e_test

import (
	"fmt"
	"math/rand"
	"time"

	. "github.com/onsi/ginkgo"
	"github.com/solo-io/solo-kit/test/helpers"
	"github.com/solo-io/squash/test/testutils"

	"testing"
)

func TestE2e(t *testing.T) {

	helpers.RegisterCommonFailHandlers()
	helpers.SetupLog()

	RunSpecs(t, "E2e Squash Suite")
}

var _ = BeforeSuite(func() {
	testutils.DeclareTestConditions()

	seed := time.Now().UnixNano()
	fmt.Printf("rand seed: %v\n", seed)
	rand.Seed(seed)
})
