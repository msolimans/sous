package server

import (
	"net/http"
	"net/http/pprof"
	"os"

	"github.com/opentable/sous/config"
	"github.com/opentable/sous/ext/storage"
	"github.com/opentable/sous/lib"
	"github.com/opentable/sous/util/logging"
	"github.com/opentable/sous/util/restful"
	"github.com/pkg/errors"
)

type (
	logSet interface {
		Vomitf(format string, a ...interface{})
		Debugf(format string, a ...interface{})
		Warnf(format string, a ...interface{})
	}

	userExtractor struct{}
)

type (
	// ComponentLocator is a service locator for the Sous components that server
	// endpoints need to function.
	ComponentLocator struct {
		logging.LogSink
		*config.Config
		sous.Inserter
		sous.StateManager
		ResolveFilter *sous.ResolveFilter
		*sous.AutoResolver
	}
)

func (ctx ComponentLocator) liveState() *sous.State {
	state, err := ctx.StateManager.ReadState()
	if os.IsNotExist(errors.Cause(err)) || storage.IsGSMError(err) {
		ctx.Warnf("error reading state:", err)
		ctx.Warnf("defaulting to empty state")
		return sous.NewState()
	}
	if err != nil {
		return nil
	}
	return state
}

func (userExtractor) GetUser(req *http.Request) ClientUser {
	return ClientUser{
		Name:  req.Header.Get("Sous-User-Name"),
		Email: req.Header.Get("Sous-User-Email"),
	}
}

// Run starts a server up.
func Run(laddr string, handler http.Handler) error {
	s := &http.Server{Addr: laddr, Handler: handler}
	return s.ListenAndServe()
}

// Handler builds the http.Handler for the Sous server httprouter.
func Handler(sc ComponentLocator, metrics http.Handler, ls logSet) http.Handler {
	handler := mux(sc, ls)
	addMetrics(handler, metrics)
	return handler
}

// Handler builds the http.Handler for the Sous server httprouter.
func ProfilingHandler(sc ComponentLocator, metrics http.Handler, ls logSet) http.Handler {
	handler := mux(sc, ls)
	addMetrics(handler, metrics)
	addProfiling(handler)
	return handler
}

func mux(sc ComponentLocator, ls logSet) *http.ServeMux {
	router := routemap(sc).BuildRouter(ls)

	handler := http.NewServeMux()
	handler.Handle("/", router)
	return handler
}

func routemap(context ComponentLocator) *restful.RouteMap {
	return &restful.RouteMap{
		{"gdm", "/gdm", newGDMResource(context)},
		{"defs", "/defs", newStateDefResource(context)},
		{"manifest", "/manifest", newManifestResource(context)},
		{"artifact", "/artifact", newArtifactResource(context)},
		{"status", "/status", newStatusResource(context)},
		{"servers", "/servers", newServerListResource(context)},
	}
}

func addMetrics(handler *http.ServeMux, metrics http.Handler) {
	handler.Handle("/debug/metrics", metrics)
}

func addProfiling(handler *http.ServeMux) {
	handler.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
	handler.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
	handler.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
	handler.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
	handler.Handle("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))
}
