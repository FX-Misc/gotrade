package gotrade

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type HttpServer struct {
	strategies map[string]Strategy
	port       int
}

func NewHttpServer(strategies map[string]Strategy, port int) *HttpServer {
	hs := new(HttpServer)
	hs.strategies = strategies
	hs.port = port
	return hs
}

func (hs *HttpServer) Serve() {
	http.HandleFunc("/", hs.list)
	http.HandleFunc("/pause", hs.pause)
	http.HandleFunc("/start", hs.start)
	http.ListenAndServe(fmt.Sprintf(":%d", hs.port), nil)
}

func (hs *HttpServer) list(w http.ResponseWriter, r *http.Request) {
	strategies := make(map[string]bool, 0)
	for name, strategy := range hs.strategies {
		strategies[name] = strategy.Status()
	}
	strategiesStr, _ := json.Marshal(strategies)
	w.Write(strategiesStr)
}

func (hs *HttpServer) pause(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	strategy, ok := hs.strategies[name]
	if !ok {
		w.WriteHeader(404)
		return
	}
	strategy.Pause()
	return
}

func (hs *HttpServer) start(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	strategy, ok := hs.strategies[name]
	if !ok {
		w.WriteHeader(404)
		return
	}
	strategy.Start()
	return
}

func (hs *HttpServer) reload(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	strategy, ok := hs.strategies[name]
	if !ok {
		w.WriteHeader(404)
		return
	}
	strategy.Reload()
	return
}
