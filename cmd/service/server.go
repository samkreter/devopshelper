package main

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/samkreter/go-core/httputil"
	"github.com/samkreter/go-core/log"
)

const (
	defaultAddr = "localhost:8080"
)

// Server holds configuration for the server
type Server struct {
	Addr       string
	httpClient *http.Client
}

// NewServer creates a new server
func NewServer(addr string) (*Server, error) {
	if addr == "" {
		addr = defaultAddr
	}

	return &Server{
		Addr:       addr,
		httpClient: httputil.NewHTTPClient(true, true, true),
	}, nil
}

// Run start the frontend server
func (s *Server) Run() {
	router := mux.NewRouter()

	router.Handle("/", http.FileServer(http.Dir("static")))
	router.HandleFunc("/create", s.handleCreate).Methods("POST")

	tracingRouter := httputil.SetUpHandler(router, &httputil.HandlerConfig{
		CorrelationEnabled: true,
		LoggingEnabled:     true,
		TracingEnabled:     true,
	})

	log.G(context.TODO()).WithField("address: ", s.Addr).Info("Starting Frontend Server:")
	log.G(context.TODO()).Fatal(http.ListenAndServe(s.Addr, tracingRouter))
}

func (s *Server) handleCreate(w http.ResponseWriter, req *http.Request) {

}
