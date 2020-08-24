package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	adogit "github.com/microsoft/azure-devops-go-api/azuredevops/git"
    adoidentity "github.com/microsoft/azure-devops-go-api/azuredevops/identity"
	"github.com/rs/cors"
	"github.com/samkreter/go-core/httputil"
	"github.com/samkreter/go-core/log"

	"github.com/samkreter/devopshelper/pkg/store"
	"github.com/samkreter/devopshelper/pkg/types"
)

type contextKey string

const (
	userInfoContextKey = contextKey("userinfo")

	defaultAddr         = "localhost:8080"
	currentUserLogField = "currentUser"
	// GraphURI uri to grab the currently logged in users identity
	GraphURI = "https://graph.microsoft.com/v1.0/me"
)

// Options to for starter the apiserver
type Options struct {
	AllowCORS bool
	Addr      string
	Admins    []string
}

// Server holds configuration for the server
type Server struct {
	AdoGitClient adogit.Client
	AdoIdentityClient adoidentity.Client
	RepoStore  store.RepositoryStore
	Options    *Options
}

// NewServer creates a new server
func NewServer(adoGitClient adogit.Client, adoIdentityClient adoidentity.Client, repoStore store.RepositoryStore, o *Options) (*Server, error) {
	if o.Addr == "" {
		o.Addr = defaultAddr
	}

	log.G(context.TODO()).Infof("Adding admins: '%s'", strings.Join(o.Admins, ", "))

	return &Server{
		AdoGitClient: adoGitClient,
		AdoIdentityClient: adoIdentityClient,
		RepoStore:  repoStore,
		Options:    o,
	}, nil
}

// Run start the frontend server
func (s *Server) Run() {
	router := mux.NewRouter()

	router.Handle("/", http.FileServer(http.Dir("static")))

	// Base Groups
	router.HandleFunc("/api/basegroups", s.GetBaseGroups).Methods("GET")
	router.HandleFunc("/api/basegroups/{baseGroupName}", s.GetBaseGroup).Methods("GET")
	router.HandleFunc("/api/basegroups/{baseGroupName}", s.PutBaseGroup).Methods("PUT")
	router.HandleFunc("/api/basegroups/{baseGroupName}", s.DeleteBaseGroup).Methods("DELETE")

	// Repos
	router.HandleFunc("/api/repositories", s.GetRepositories).Methods("GET")
	router.HandleFunc("/api/repositories", s.PostRepository).Methods("POST")
	router.HandleFunc("/api/projects/{project}/repositories", s.GetRepositoryPerProject).Methods("GET")
	router.HandleFunc("/api/projects/{project}/repositories/{repository}", s.GetRepository).Methods("GET")
	router.HandleFunc("/api/projects/{project}/repositories/{repository}", s.PutRepository).Methods("PUT")
	router.HandleFunc("/api/projects/{project}/repositories/{repository}", s.DeleteRepository).Methods("DELETE")

	// Reviewers
	router.HandleFunc("/api/projects/{project}/repositories/{repository}/reviewGroups/{reviewGroup}/reviewers", s.GetReviewerGroupToRepository).Methods("GET")
	router.HandleFunc("/api/projects/{project}/repositories/{repository}/reviewGroups/{reviewGroup}/reviewers", s.AddReviewerToRepository).Methods("POST")
	router.HandleFunc("/api/projects/{project}/repositories/{repository}/reviewGroups/{reviewGroup}/reviewers/{reviewerAlias}", s.DeleteReviewerToRepository).Methods("DELETE")

	// Repos Base Groups
	router.HandleFunc("/api/projects/{project}/repositories/{repository}/basegroup/{basegroupName}", s.AddBaseGroupToRepository).Methods("PUT")

	// Enable
	router.HandleFunc("/api/projects/{project}/repositories/{repository}/enable", s.EnableRepository).Methods("POST")
	router.HandleFunc("/api/projects/{project}/repositories/{repository}/disable", s.DisableRepository).Methods("POST")
	router.PathPrefix("/").HandlerFunc(s.catchAllHandler)

	// Add authentication handler
	handler := AuthMiddleware(router)

	handler = httputil.SetUpHandler(handler, &httputil.HandlerConfig{
		CorrelationEnabled: true,
		LoggingEnabled:     true,
		TracingEnabled:     true,
	})

	// allow cors, for frontend access
	// TODO: make cors more restrivtive
	if s.Options.AllowCORS {
		log.G(context.TODO()).Info("Enabling CORS")
		handler = cors.AllowAll().Handler(handler)
	}

	log.G(context.TODO()).WithField("address: ", s.Options.Addr).Info("Starting Frontend Server:")
	log.G(context.TODO()).Fatal(http.ListenAndServe(s.Options.Addr, handler))
}

func (s *Server) catchAllHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "API route not found.")
}

// AuthMiddleware only allows users in the security group and adds the user into the request context
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		authHeader := req.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "no Authorization header found", http.StatusUnauthorized)
			return
		}

		ctx := req.Context()
		logger := log.G(ctx)

		user, err := getAuthenticatedUser(authHeader)
		if err != nil {
			logger.Errorf("failed to auth user with err: '%v'", err)
			http.Error(w, "authentication was invalid", http.StatusUnauthorized)
			return
		}

		// Add the current user to the log fields
		ctx = log.WithLogger(ctx, logger.WithField(currentUserLogField, user.Mail))

		ctx = context.WithValue(ctx, userInfoContextKey, user)

		logger.Infof("using logged in user: '%s'", user.Mail)

		next.ServeHTTP(w, req.WithContext(ctx))
	})
}

func getAuthenticatedUser(authToken string) (*types.GraphUser, error) {
	req, err := http.NewRequest("GET", GraphURI, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", authToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("graph returned non 200 status code: '%d'", resp.StatusCode)
	}

	defer resp.Body.Close()

	var user types.GraphUser
	err = json.NewDecoder(resp.Body).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
