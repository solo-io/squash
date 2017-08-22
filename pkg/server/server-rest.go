package server

import (
	"context"
	"errors"
	"fmt"
	"time"

	"math/rand"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/davecgh/go-spew/spew"
	"github.com/go-openapi/runtime/middleware"
	"github.com/solo-io/squash/pkg/models"
	"github.com/solo-io/squash/pkg/platforms"
	"github.com/solo-io/squash/pkg/restapi/operations/debugconfig"
	"github.com/solo-io/squash/pkg/restapi/operations/debugsessions"
)

type watchedDebugConfig struct {
	config  models.DebugConfig
	session chan models.DebugSession
	cancel  context.CancelFunc
}

type RestHandler struct {
	debugConfigs        map[string]watchedDebugConfig
	debugConfigsMapLock sync.RWMutex

	nodeConfigs        map[string]chan models.DebugConfig
	nodeConfigsMaplock sync.RWMutex

	serviceWatcher   platforms.ServiceWatcher
	containerLocator platforms.ContainerLocator
}

func NewRestHandler(serviceWatcher platforms.ServiceWatcher, containerLocator platforms.ContainerLocator) *RestHandler {
	return &RestHandler{
		debugConfigs:     make(map[string]watchedDebugConfig),
		nodeConfigs:      make(map[string]chan models.DebugConfig),
		serviceWatcher:   serviceWatcher,
		containerLocator: containerLocator,
	}
}

func (r *RestHandler) AddDebugConfig(params debugconfig.AddDebugConfigParams) middleware.Responder {

	dbgcfg := *params.Body
	id := fmt.Sprintf("%d", rand.Int31())
	dbgcfg.ID = id
	dbgcfg.Active = true

	logger := log.WithField("id", id)

	logger.WithField("dbgcfg", spew.Sdump(dbgcfg)).Info("AddDebugConfig")

	switch *dbgcfg.Attachment.Type {

	case models.AttachmentTypeService:

		if dbgcfg.Immediately {
			logger.WithField("dbgcfg", spew.Sdump(dbgcfg)).Warn("AddDebugConfig rejecting bad service request")
			return debugconfig.NewAddDebugConfigBadRequest()
		}

		cancel, err := r.watchService(dbgcfg)
		if err != nil {
			logger.WithFields(log.Fields{"dbgcfg": spew.Sdump(dbgcfg), "err": err}).Error("AddDebugConfig watch service failed")

			return debugconfig.NewAddDebugConfigNotFound()
		}

		logger.WithFields(log.Fields{"dbgcfg": spew.Sdump(dbgcfg)}).Info("AddDebugConfig adding service")
		r.debugConfigsMapLock.Lock()
		r.debugConfigs[id] = watchedDebugConfig{
			config:  dbgcfg,
			cancel:  cancel,
			session: make(chan models.DebugSession, 1), // <- allow one session
		}
		r.debugConfigsMapLock.Unlock()

	case models.AttachmentTypeContainer:
		logger.Info("AddDebugConfig searching for container attachment")

		node, err := r.findContainerNode(params.HTTPRequest.Context(), *dbgcfg.Attachment.Name)
		if (err != nil) || (node == "") {
			return debugconfig.NewAddDebugConfigNotFound()
		}

		logger.WithFields(log.Fields{"dbgcfg": spew.Sdump(dbgcfg)}).Info("AddDebugConfig adding container")

		r.debugConfigsMapLock.Lock()
		r.debugConfigs[id] = watchedDebugConfig{
			config:  dbgcfg,
			session: make(chan models.DebugSession, 1),
		}
		r.debugConfigsMapLock.Unlock()

		select {
		case r.getNodeChan(node) <- dbgcfg:
		default:
			logger.Error("Node channel full!")
			return debugconfig.NewAddDebugConfigServiceUnavailable()
		}
	default:
		return debugconfig.NewAddDebugConfigBadRequest()

	}

	return debugconfig.NewAddDebugConfigCreated().WithPayload(&dbgcfg)
}

func (r *RestHandler) findContainerNode(ctx context.Context, container string) (string, error) {

	if r.containerLocator == nil {
		return "", errors.New("no container locator provided")
	}
	c, err := r.containerLocator.Locate(ctx, container)
	if err != nil {
		return "", err
	}
	if c.Node == "" {
		return "", errors.New("container not assigned to node")
	}

	return c.Node, nil
}

func (r *RestHandler) PutDebugSession(params debugsessions.PutDebugSessionParams) middleware.Responder {

	dbgcfgid := params.DebugConfigID
	dbgsession := *params.Body

	logger := log.WithFields(log.Fields{"dbgsession": dbgsession, "dbgcfgid": dbgcfgid})
	logger.Info("PutDebugSession called!")

	r.debugConfigsMapLock.RLock()
	watchedcfg, ok := r.debugConfigs[dbgcfgid]
	r.debugConfigsMapLock.RUnlock()

	if !ok {
		logger.Error("PutDebugSession dbgconfig not found!")
		return debugsessions.NewPutDebugSessionNotFound()
	}
	if watchedcfg.session == nil {
		logger.Info("PutDebugSession context is done")
		return debugsessions.NewPutDebugSessionPreconditionFailed()
	}

	var sessionchan chan models.DebugSession
	var wasactive bool
	var cancelfunc func()
	r.debugConfigsMapLock.Lock()
	if cfg, ok := r.debugConfigs[dbgcfgid]; ok {
		switch *watchedcfg.config.Attachment.Type {
		case models.AttachmentTypeContainer:
		// nothing here
		case models.AttachmentTypeService:
			// Cancel the service watch
			if cfg.cancel != nil {
				cancelfunc = cfg.cancel
			}
			cfg.cancel = nil
		}
		wasactive = cfg.config.Active
		cfg.config.Active = false
		sessionchan = cfg.session
		r.debugConfigs[dbgcfgid] = cfg
	}
	r.debugConfigsMapLock.Unlock()

	if cancelfunc != nil {
		cancelfunc()
	}

	if wasactive == false {
		logger.Info("PutDebugSession context is done")
		return debugsessions.NewPutDebugSessionPreconditionFailed()
	}

	select {
	case sessionchan <- dbgsession: // if i did the math right, this will not block
		close(sessionchan)
		logger.Info("PutDebugSession dbgsession created!")
		return debugsessions.NewPutDebugSessionCreated().WithPayload(&dbgsession)
	default: // <- but just in case
		logger.Error("session chan block - this should never happen!")
		return debugsessions.NewPutDebugSessionPreconditionFailed()
	}

}

func (r *RestHandler) DeleteDebugConfig(params debugconfig.DeleteDebugConfigParams) middleware.Responder {

	log.WithFields(log.Fields{"DebugConfigID": params.DebugConfigID}).Info("DeleteDebugConfig called!")

	r.debugConfigsMapLock.Lock()
	defer r.debugConfigsMapLock.Unlock()

	id := params.DebugConfigID
	if cfg, ok := r.debugConfigs[id]; ok {
		if cfg.cancel != nil {
			cfg.cancel()
		}

		if cfg.config.Active {
			close(cfg.session)
		}

		delete(r.debugConfigs, id)
		return debugconfig.NewDeleteDebugConfigOK()
	}
	return debugconfig.NewDeleteDebugConfigNotFound()
}

func (r *RestHandler) GetDebugConfig(params debugconfig.GetDebugConfigParams) middleware.Responder {
	log.WithFields(log.Fields{"DebugConfigID": params.DebugConfigID}).Info("GetDebugConfig called!")
	r.debugConfigsMapLock.Lock()
	defer r.debugConfigsMapLock.Unlock()

	id := params.DebugConfigID
	if cfg, ok := r.debugConfigs[id]; ok {
		return debugconfig.NewGetDebugConfigOK().WithPayload(&cfg.config)
	}
	return debugconfig.NewGetDebugConfigNotFound()
}

func (r *RestHandler) GetDebugConfigs(params debugconfig.GetDebugConfigsParams) middleware.Responder {
	log.Info("GetDebugConfigs called!")

	var debugconfigs []*models.DebugConfig
	r.debugConfigsMapLock.RLock()
	defer r.debugConfigsMapLock.RUnlock()

	for _, cfg := range r.debugConfigs {
		config := cfg.config
		debugconfigs = append(debugconfigs, &config)
	}
	return debugconfig.NewGetDebugConfigsOK().WithPayload(debugconfigs)
}
func (r *RestHandler) PopContainerToDebug(params debugconfig.PopContainerToDebugParams) middleware.Responder {
	node := params.Node
	logger := log.WithFields(log.Fields{"node": node})
	logger.Info("PopContainerToDebug called!")
	select {
	case dbgconfig := <-r.getNodeChan(node):
		log.WithFields(log.Fields{"node": node, "dbgconfig": dbgconfig}).Info("PopContainerToDebug got debug config!")
		return debugconfig.NewPopContainerToDebugOK().WithPayload(&dbgconfig)

	case <-params.HTTPRequest.Context().Done():
		logger.Warn("PopContainerToDebug context is done")
		return debugsessions.NewPutDebugSessionPreconditionFailed()

	case <-time.After(10 * time.Second):
		logger.Debug("PopContainerToDebug timeout waiting for debug config")
		return debugconfig.NewPopContainerToDebugRequestTimeout()
	}
}
func (r *RestHandler) PopDebugSession(params debugsessions.PopDebugSessionParams) middleware.Responder {

	logger := log.WithField("DebugConfigID", params.DebugConfigID)
	logger.Info("PopDebugSession")
	dbgcfgid := params.DebugConfigID

	timeout := 10 * time.Second
	if params.XTimeout != nil {
		timeout = time.Duration((*params.XTimeout) * float64(time.Second))
	}

	var sessionchan chan models.DebugSession
	r.debugConfigsMapLock.RLock()
	watchedcfg, ok := r.debugConfigs[dbgcfgid]
	if ok {
		sessionchan = watchedcfg.session
	}
	r.debugConfigsMapLock.RUnlock()

	if sessionchan == nil {
		// should never happen
		return debugsessions.NewPopDebugSessionNotFound()
	}

	select {
	case dbgsession, ok := <-sessionchan:
		if ok {
			logger.WithField("dbgsession", dbgsession).Info("PopDebugSession: config found")
			return debugsessions.NewPopDebugSessionOK().WithPayload(&dbgsession)
		}
		// TODO: return a better error

	case <-params.HTTPRequest.Context().Done():
		logger.Warn("PopDebugSessionRequest context is done")
		return debugsessions.NewPopDebugSessionRequestTimeout()
	case <-time.After(timeout):
		return debugsessions.NewPopDebugSessionRequestTimeout()
	}
	log.WithField("dbgcfgid", dbgcfgid).Error("config not found - PopDebugSession")
	return debugsessions.NewPopDebugSessionNotFound()
}

func (r *RestHandler) UpdateDebugConfig(params debugconfig.UpdateDebugConfigParams) middleware.Responder {
	r.debugConfigsMapLock.Lock()
	defer r.debugConfigsMapLock.Unlock()

	id := params.DebugConfigID
	if cfg, ok := r.debugConfigs[id]; ok {
		newcfg := cfg.config

		newcfg.Breakpoints = params.Body.Breakpoints
		newcfg.Image = params.Body.Image

		cfg.config = newcfg
		r.debugConfigs[id] = cfg
		return debugconfig.NewUpdateDebugConfigOK().WithPayload(&cfg.config)
	}
	return debugconfig.NewUpdateDebugConfigNotFound()

}

func (r *RestHandler) watchService(dbg models.DebugConfig) (context.CancelFunc, error) {

	if r.serviceWatcher == nil {
		log.Warn("Service watcher not defined. consider configuring cluster?")
		return nil, errors.New("service watcher not defined")
	}

	ctx, cancel := context.WithCancel(context.Background())
	c, err := r.serviceWatcher.WatchService(ctx, *dbg.Attachment.Name)

	if err != nil {
		cancel()
		return nil, err
	}

	go func() {
		for container := range c {
			if container.Image == *dbg.Image {
				log.WithFields(log.Fields{"servicecontainer": c, "dbgconfig": spew.Sdump(dbg)}).Debug("Found container to debug. sending to node")
				var cotainerdbgcfg models.DebugConfig
				cotainerdbgcfg = dbg
				s := models.AttachmentTypeContainer
				cotainerdbgcfg.Attachment = &models.Attachment{
					Name: &container.Name,
					Type: &s,
				}
				r.getNodeChan(container.Node) <- cotainerdbgcfg
			} else {
				log.WithFields(log.Fields{"servicecontainer": spew.Sdump(c), "dbgconfig": spew.Sdump(dbg)}).Debug("Service container doesn't match dbgconfig image")

			}

		}
	}()
	return cancel, nil
}

func (r *RestHandler) getNodeChan(node string) chan models.DebugConfig {
	r.nodeConfigsMaplock.RLock()
	c, ok := r.nodeConfigs[node]
	r.nodeConfigsMaplock.RUnlock()

	if ok {
		return c
	}
	// slow path
	r.nodeConfigsMaplock.Lock()
	defer r.nodeConfigsMaplock.Unlock()
	// check again, incase there was a race
	c, ok = r.nodeConfigs[node]

	if ok {
		return c
	}

	c = make(chan models.DebugConfig, 10)
	r.nodeConfigs[node] = c
	return c

}
