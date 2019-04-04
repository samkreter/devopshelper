package server

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/samkreter/go-core/httputil"
	"github.com/samkreter/go-core/log"
	vsts "github.com/samkreter/vsts-goclient/client"
	"github.com/samkreter/vstsautoreviewer/pkg/store"
)

const (
	defaultAddr = "localhost:8080"
)

// Server holds configuration for the server
type Server struct {
	Addr       string
	vstsClient *vsts.Client
	httpClient *http.Client
	RepoStore  store.RepositoryStore
}

// NewServer creates a new server
func NewServer(addr string, repoStore store.RepositoryStore) (*Server, error) {
	if addr == "" {
		addr = defaultAddr
	}

	return &Server{
		Addr:       addr,
		RepoStore:  repoStore,
		httpClient: httputil.NewHTTPClient(true, true, true),
	}, nil
}

// Run start the frontend server
func (s *Server) Run() {
	router := mux.NewRouter()

	router.Handle("/", http.FileServer(http.Dir("static")))

	// Base Groups
	router.HandleFunc("/api/basegroups", s.handleGetBaseGroups).Methods("GET")
	router.HandleFunc("/api/basegroups/{baseGroupName}", s.handleGetBaseGroup).Methods("GET")
	router.HandleFunc("/api/basegroups/{basegroupName}", s.handlePutBaseGroup).Methods("PUT")
	router.HandleFunc("/api/basegroups/{basegroupName}", s.handleDeleteBaseGroup).Methods("DELETE")

	// Repos
	router.HandleFunc("/api/repositories", s.handleGetRepositories).Methods("GET")
	router.HandleFunc("/api/repositories/{repositoryID}", s.handleGetRepository).Methods("GET")
	router.HandleFunc("/api/repositories/{repositoryID}", s.handlePutRepository).Methods("PUT")
	router.HandleFunc("/api/repositories/{repositoryID}", s.handleDeleteRepository).Methods("DELETE")

	// Reviewers
	router.HandleFunc("/api/repositories/{repositoryID}/reviewGroups/{reviewGroup}/reviewers", s.handleGetReviewerGroupToRepository).Methods("GET")
	router.HandleFunc("/api/repositories/{repositoryID}/reviewGroups/{reviewGroup}/reviewers", s.handleAddReviewerToRepository).Methods("POST")
	router.HandleFunc("/api/repositories/{repositoryID}/reviewGroups/{reviewGroup}/reviewers/{reviewerAlias}", s.handleDeleteReviewerToRepository).Methods("DELETE")

	// Repos Base Groups
	router.HandleFunc("/api/repositories/{repositoryID}/basegroup/{basegroupName}", s.handleAddBaseGroupToRepository).Methods("PUT")

	// Enable
	router.HandleFunc("/api/repositories/{repositoryID}/enable", s.handleEnableRepository).Methods("POST")
	router.HandleFunc("/api/repositories/{repositoryID}/disable", s.handleDisableRepository).Methods("POST")

	tracingRouter := httputil.SetUpHandler(router, &httputil.HandlerConfig{
		CorrelationEnabled: true,
		LoggingEnabled:     true,
		TracingEnabled:     true,
	})

	log.G(context.TODO()).WithField("address: ", s.Addr).Info("Starting Frontend Server:")
	log.G(context.TODO()).Fatal(http.ListenAndServe(s.Addr, tracingRouter))
}
