package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/solo-io/squash/pkg/squash"

	dbgutils "github.com/solo-io/squash/pkg/debuggers/utils"
	"github.com/solo-io/squash/pkg/platforms/kubernetes"
	"github.com/solo-io/squash/pkg/version"
)

func main() {
	log.SetLevel(log.DebugLevel)

	customFormatter := new(log.TextFormatter)
	log.SetFormatter(customFormatter)

	log.Infof("squash started %v, %v", version.Version, version.TimeStamp)

	mustGetContainerProcessLocator()
	err := squash.RunSquashAgent(dbgutils.GetParticularDebugger)
	log.WithError(err).Fatal("Error running debug bridge")

}

// The debugging pod needs to be able to get a container process
// This function is a way to fail early (from the squash pod) if the running
// version of Kubernetes does not support the needed API.
func mustGetContainerProcessLocator() {
	_, err := kubernetes.NewContainerProcess()
	if err != nil {
		_, err := kubernetes.NewCRIContainerProcessAlphaV1()
		if err != nil {
			log.WithError(err).Fatal("Cannot get container process locator")
		}
	}
}
