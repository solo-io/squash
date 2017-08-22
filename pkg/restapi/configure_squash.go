package restapi

import (
	"crypto/tls"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"k8s.io/client-go/rest"

	errors "github.com/go-openapi/errors"
	runtime "github.com/go-openapi/runtime"
	"github.com/go-openapi/swag"
	graceful "github.com/tylerb/graceful"

	"github.com/solo-io/squash/pkg/platforms"
	"github.com/solo-io/squash/pkg/platforms/debug"
	"github.com/solo-io/squash/pkg/platforms/kubernetes"
	"github.com/solo-io/squash/pkg/restapi/operations"
	"github.com/solo-io/squash/pkg/restapi/operations/debugconfig"
	"github.com/solo-io/squash/pkg/restapi/operations/debugsessions"
	"github.com/solo-io/squash/pkg/server"
)

// This file is safe to edit. Once it exists it will not be overwritten

//go:generate swagger generate server --target .. --name Squash --spec ../api.yaml

var (
	options = &struct {
		Cluster    string `long:"cluster" short:"c" description:"Cluster type. currently only kube"`
		KubeConfig string `long:"kubeurl" description:"Kube url (useful with kubectl proxy)"`
	}{"kube", ""}
)

func configureFlags(api *operations.SquashAPI) {

	api.CommandLineOptionsGroups = []swag.CommandLineOptionsGroup{
		swag.CommandLineOptionsGroup{
			LongDescription:  "cluster related operations",
			ShortDescription: "cluster",
			Options:          options,
		},
	}

}

func configureAPI(api *operations.SquashAPI) http.Handler {
	log.SetLevel(log.DebugLevel)
	// configure the api here
	api.ServeError = errors.ServeError

	// Set your custom logger if needed. Default one is log.Printf
	// Expected interface func(string, ...interface{})
	//
	// Example:
	// s.api.Logger = log.Printf

	api.JSONConsumer = runtime.JSONConsumer()

	api.JSONProducer = runtime.JSONProducer()

	var sw platforms.ServiceWatcher = nil
	var cl platforms.ContainerLocator = nil
	switch options.Cluster {
	case "kube":
		var config *rest.Config = nil
		if options.KubeConfig != "" {
			config = &rest.Config{Host: options.KubeConfig}
		}

		kube, err := kubernetes.NewKubeOperations(nil, config)
		if err != nil {
			log.Fatalln(err)
		}
		sw = kube
		cl = kube
	case "debug":
		var d debug.DebugPlatform
		sw = &d
		cl = &d
	}

	handler := server.NewRestHandler(sw, cl)
	api.DebugconfigAddDebugConfigHandler = debugconfig.AddDebugConfigHandlerFunc(handler.AddDebugConfig)
	api.DebugsessionsPutDebugSessionHandler = debugsessions.PutDebugSessionHandlerFunc(handler.PutDebugSession)
	api.DebugconfigDeleteDebugConfigHandler = debugconfig.DeleteDebugConfigHandlerFunc(handler.DeleteDebugConfig)
	api.DebugconfigGetDebugConfigsHandler = debugconfig.GetDebugConfigsHandlerFunc(handler.GetDebugConfigs)
	api.DebugconfigGetDebugConfigHandler = debugconfig.GetDebugConfigHandlerFunc(handler.GetDebugConfig)
	api.DebugconfigPopContainerToDebugHandler = debugconfig.PopContainerToDebugHandlerFunc(handler.PopContainerToDebug)
	api.DebugsessionsPopDebugSessionHandler = debugsessions.PopDebugSessionHandlerFunc(handler.PopDebugSession)
	api.DebugconfigUpdateDebugConfigHandler = debugconfig.UpdateDebugConfigHandlerFunc(handler.UpdateDebugConfig)

	api.ServerShutdown = func() {}

	return setupGlobalMiddleware(api.Serve(setupMiddlewares))
}

// The TLS configuration before HTTPS server starts.
func configureTLS(tlsConfig *tls.Config) {
	// Make all necessary changes to the TLS configuration here.
}

// As soon as server is initialized but not run yet, this function will be called.
// If you need to modify a config, store server instance to stop it individually later, this is the place.
// This function can be called multiple times, depending on the number of serving schemes.
// scheme value will be set accordingly: "http", "https" or "unix"
func configureServer(s *graceful.Server, scheme, addr string) {
}

// The middleware configuration is for the handler executors. These do not apply to the swagger.json document.
// The middleware executes after routing but before authentication, binding and validation
func setupMiddlewares(handler http.Handler) http.Handler {
	return handler
}

// The middleware configuration happens before anything, this middleware also applies to serving the swagger.json document.
// So this is a good place to plug in a panic handling middleware, logging and metrics
func setupGlobalMiddleware(handler http.Handler) http.Handler {
	return handler
}
