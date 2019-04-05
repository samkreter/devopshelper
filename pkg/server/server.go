package server

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/samkreter/go-core/httputil"
	"github.com/samkreter/go-core/log"
	vsts "github.com/samkreter/vsts-goclient/client"
	"github.com/samkreter/vstsautoreviewer/pkg/store"
)

const (
	defaultAddr = "localhost:8080"
	GraphURI    = "https://graph.microsoft.com/v1.0/me"
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
	router.HandleFunc("/api/repositories", s.handlePostRepository).Methods("POST")
	router.HandleFunc("/api/projects/{project}/repositories", s.handleGetRepositoryPerProject).Methods("GET")
	router.HandleFunc("/api/projects/{project}/repositories/{repository}", s.handleGetRepository).Methods("GET")
	router.HandleFunc("/api/projects/{project}/repositories/{repository}", s.handlePutRepository).Methods("PUT")
	router.HandleFunc("/api/projects/{project}/repositories/{repository}", s.handleDeleteRepository).Methods("DELETE")

	// Reviewers
	router.HandleFunc("/api/projects/{project}/repositories/{repository}/reviewGroups/{reviewGroup}/reviewers", s.handleGetReviewerGroupToRepository).Methods("GET")
	router.HandleFunc("/api/projects/{project}/repositories/{repository}/reviewGroups/{reviewGroup}/reviewers", s.handleAddReviewerToRepository).Methods("POST")
	router.HandleFunc("/api/projects/{project}/repositories/{repository}/reviewGroups/{reviewGroup}/reviewers/{reviewerAlias}", s.handleDeleteReviewerToRepository).Methods("DELETE")

	// Repos Base Groups
	router.HandleFunc("/api/projects/{project}/repositories/{repository}/basegroup/{basegroupName}", s.handleAddBaseGroupToRepository).Methods("PUT")

	// Enable
	router.HandleFunc("/api/projects/{project}/repositories/{repository}/enable", s.handleEnableRepository).Methods("POST")
	router.HandleFunc("/api/projects/{project}/repositories/{repository}/disable", s.handleDisableRepository).Methods("POST")

	tracingRouter := httputil.SetUpHandler(router, &httputil.HandlerConfig{
		CorrelationEnabled: true,
		LoggingEnabled:     true,
		TracingEnabled:     true,
	})

	log.G(context.TODO()).WithField("address: ", s.Addr).Info("Starting Frontend Server:")
	log.G(context.TODO()).Fatal(http.ListenAndServe(s.Addr, tracingRouter))
}

func authenticate(accessToken string) (*GraphUser, error) {
	req, err := http.NewRequest("GET", GraphURI, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var user GraphUser
	err = json.NewDecoder(resp.Body).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

type GraphUser struct {
	OdataContext      string      `json:"@odata.context"`
	BusinessPhones    []string    `json:"businessPhones"`
	DisplayName       string      `json:"displayName"`
	GivenName         string      `json:"givenName"`
	JobTitle          string      `json:"jobTitle"`
	Mail              string      `json:"mail"`
	MobilePhone       interface{} `json:"mobilePhone"`
	OfficeLocation    string      `json:"officeLocation"`
	PreferredLanguage interface{} `json:"preferredLanguage"`
	Surname           string      `json:"surname"`
	UserPrincipalName string      `json:"userPrincipalName"`
	ID                string      `json:"id"`
}