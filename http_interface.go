package gotrade

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type HttpServer struct {
	strategies map[string]Strategy
	port       int
}

type StrategyStatus struct {
	Name   string `json:"name"`
	Status bool   `json:"status"`
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
	http.HandleFunc("/reload", hs.reload)
	log.Printf("http serve at port %d", hs.port)
	http.ListenAndServe(fmt.Sprintf(":%d", hs.port), nil)
}

func (hs *HttpServer) list(w http.ResponseWriter, r *http.Request) {
	strategies := make([]StrategyStatus, 0)
	for name, strategy := range hs.strategies {
		strategies = append(strategies, StrategyStatus{name, strategy.Status()})
	}
	strategiesStr, _ := json.Marshal(strategies)
	w.Write(strategiesStr)
}

func (hs *HttpServer) pause(w http.ResponseWriter, r *http.Request) {
	name := r.PostFormValue("name")
	strategy, ok := hs.strategies[name]
	if !ok {
		w.WriteHeader(404)
		return
	}
	strategy.Pause()
	w.Write([]byte("true"))
	return
}

func (hs *HttpServer) start(w http.ResponseWriter, r *http.Request) {
	name := r.PostFormValue("name")
	strategy, ok := hs.strategies[name]
	if !ok {
		w.WriteHeader(404)
		return
	}
	strategy.Start()
	w.Write([]byte("true"))
	return
}

func (hs *HttpServer) reload(w http.ResponseWriter, r *http.Request) {
	name := r.PostFormValue("name")
	strategy, ok := hs.strategies[name]
	if !ok {
		w.WriteHeader(404)
		return
	}
	strategy.Reload()
	w.Write([]byte("true"))
	return
}
