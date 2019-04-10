package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/samkreter/go-core/log"
	"github.com/samkreter/vstsautoreviewer/pkg/store"
	"github.com/samkreter/vstsautoreviewer/pkg/types"
	"github.com/samkreter/vstsautoreviewer/pkg/utils"
)

// GetReviewerGroupToRepository gets a single reviewer group from a repository
func (s *Server) GetReviewerGroupToRepository(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	vars := mux.Vars(req)
	logger := log.G(ctx)

	repoName := vars["repository"]
	if repoName == "" {
		errMsg := "repository name missing from request"
		logger.Errorf("getReviewerGroupToRepository %v", errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	projectName := vars["project"]
	if projectName == "" {
		errMsg := "project name missing from request"
		logger.Errorf("getReviewerGroupToRepository: %v", errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	reviewGroupName := vars["reviewGroup"]
	if reviewGroupName == "" {
		http.Error(w, "reviewGroup missing from request", http.StatusBadRequest)
		return
	}

	repo, err := s.RepoStore.GetRepositoryByName(ctx, repoName, projectName)
	if err != nil {
		if err == store.ErrNotFound {
			http.Error(w, fmt.Sprintf("repository '%s/%s' not found", projectName, repoName), http.StatusBadRequest)
			return
		}
		logger.Errorf(fmt.Sprintf("database error: '%v'", err))
		http.Error(w, fmt.Sprintf("database error: '%v'", err), http.StatusInternalServerError)
	}

	reviewerGroup, ok := repo.ReviewerGroups[reviewGroupName]
	if !ok {
		http.Error(w, fmt.Sprintf("reviewer group '%s' not found", reviewGroupName), http.StatusBadRequest)
		return
	}

	err = json.NewEncoder(w).Encode(reviewerGroup)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to marshal request: '%v'", err), http.StatusInternalServerError)
		return
	}
}

// DeleteReviewerToRepository deletes a single reviewer from a repository
func (s *Server) DeleteReviewerToRepository(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	vars := mux.Vars(req)
	logger := log.G(ctx)

	currUser, err := getCurrentUser(ctx)
	if err != nil {
		logger.Errorf("DeleteReviewerToRepository: %v", err)
		http.Error(w, "failed to get current user", http.StatusBadRequest)
		return
	}

	repoName := vars["repository"]
	if repoName == "" {
		errMsg := "repository name missing from request"
		logger.Errorf("deleteReviewerToRepository: %v", errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	projectName := vars["project"]
	if projectName == "" {
		errMsg := "project name missing from request"
		logger.Errorf("deleteReviewerToRepository: %v", errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	reviewGroupName := vars["reviewGroup"]
	if reviewGroupName == "" {
		http.Error(w, "reviewGroup missing from request", http.StatusBadRequest)
		return
	}

	reviewerAlias := vars["reviewerAlias"]
	if reviewGroupName == "" {
		http.Error(w, "reviewGroup missing from request", http.StatusBadRequest)
		return
	}

	repo, err := s.RepoStore.GetRepositoryByName(ctx, repoName, projectName)
	if err != nil {
		if err == store.ErrNotFound {
			http.Error(w, fmt.Sprintf("repository '%s/%s' not found", projectName, repoName), http.StatusBadRequest)
			return
		}
		http.Error(w, fmt.Sprintf("database error: '%v'", err), http.StatusInternalServerError)
	}

	if !s.userHasWritePermission(currUser, repo.Owners) {
		http.Error(w, fmt.Sprintf("User %s does not have permission to delete a reviewer this repo.", currUser.Mail), http.StatusUnauthorized)
		return
	}

	reviewerGroup, ok := repo.ReviewerGroups[reviewGroupName]
	if !ok {
		http.Error(w, fmt.Sprintf("reviewer group '%s' not found", reviewGroupName), http.StatusBadRequest)
		return
	}

	for idx, reviewer := range reviewerGroup.Reviewers {
		if reviewer.Alias == reviewerAlias {
			// delete the element
			reviewerGroup.Reviewers = append(reviewerGroup.Reviewers[:idx], reviewerGroup.Reviewers[idx+1:]...)
		}
	}

	if err := s.RepoStore.UpdateRepository(ctx, repo.ID.Hex(), repo); err != nil {
		http.Error(w, fmt.Sprintf("failed to update repository: '%v'", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// AddReviewerToRepository adds a single reviewer to a repository
func (s *Server) AddReviewerToRepository(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	vars := mux.Vars(req)
	logger := log.G(ctx)

	currUser, err := getCurrentUser(ctx)
	if err != nil {
		logger.Errorf("AddReviewerToRepository: %v", err)
		http.Error(w, "failed to get current user", http.StatusBadRequest)
		return
	}

	repoName := vars["repository"]
	if repoName == "" {
		errMsg := "repository name missing from request"
		logger.Errorf("addReviewerToRepository: %v", errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	projectName := vars["project"]
	if projectName == "" {
		errMsg := "project name missing from request"
		logger.Errorf("addReviewerToRepository: %v", errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	reviewGroupName := vars["reviewGroup"]
	if reviewGroupName == "" {
		http.Error(w, "reviewGroup missing from request", http.StatusBadRequest)
		return
	}

	reviewerAlias := vars["reviewerAlias"]
	if reviewGroupName == "" {
		http.Error(w, "reviewGroup missing from request", http.StatusBadRequest)
		return
	}

	repo, err := s.RepoStore.GetRepositoryByName(ctx, repoName, projectName)
	if err != nil {
		if err == store.ErrNotFound {
			http.Error(w, fmt.Sprintf("repository '%s/%s' not found", projectName, repoName), http.StatusBadRequest)
			return
		}
		http.Error(w, fmt.Sprintf("database error: '%v'", err), http.StatusInternalServerError)
	}

	if !s.userHasWritePermission(currUser, repo.Owners) {
		http.Error(w, fmt.Sprintf("User %s does not have permission to add a reviewer this repo.", currUser.Mail), http.StatusUnauthorized)
		return
	}

	reviewerGroup, ok := repo.ReviewerGroups[reviewGroupName]
	if !ok {
		http.Error(w, fmt.Sprintf("reviewer group '%s' not found", reviewGroupName), http.StatusBadRequest)
		return
	}

	reviewer, err := utils.GetReviwerFromAlias(reviewerAlias, s.vstsClient.RestClient)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to retrive vsts id with error: '%v'", err), http.StatusInternalServerError)
	}

	reviewerGroup.Reviewers = append(reviewerGroup.Reviewers, reviewer)

	if err := s.RepoStore.UpdateRepository(ctx, repo.ID.Hex(), repo); err != nil {
		http.Error(w, fmt.Sprintf("failed to update repository: '%v'", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// AddBaseGroupToRepository adds all reviewers from a base group to a repository reviewer list
func (s *Server) AddBaseGroupToRepository(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	vars := mux.Vars(req)
	logger := log.G(ctx)

	currUser, err := getCurrentUser(ctx)
	if err != nil {
		logger.Errorf("AddBaseGroupToRepository: %v", err)
		http.Error(w, "failed to get current user", http.StatusBadRequest)
		return
	}

	repoName := vars["repository"]
	if repoName == "" {
		errMsg := "repository name missing from request"
		logger.Errorf("addBaseGroupToRepository: %v", errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	projectName := vars["project"]
	if projectName == "" {
		errMsg := "project name missing from request"
		logger.Errorf("addBaseGroupToRepository: %v", errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	baseGroupName := vars["baseGroupName"]
	if baseGroupName == "" {
		http.Error(w, ErrBaseGroupNameNotFound.Error(), http.StatusBadRequest)
		return
	}

	baseGroup, err := s.RepoStore.GetBaseGroupByName(ctx, baseGroupName)
	if err != nil {
		if err == store.ErrNotFound {
			http.Error(w, fmt.Sprintf("base group '%s' not found", baseGroupName), http.StatusBadRequest)
			return
		}
		http.Error(w, fmt.Sprintf("database error: '%v'", err), http.StatusInternalServerError)
	}

	repo, err := s.RepoStore.GetRepositoryByName(ctx, repoName, projectName)
	if err != nil {
		if err == store.ErrNotFound {
			http.Error(w, fmt.Sprintf("repository '%s/%s' not found", projectName, repoName), http.StatusBadRequest)
			return
		}
		http.Error(w, fmt.Sprintf("database error: '%v'", err), http.StatusInternalServerError)
	}

	if !s.userHasWritePermission(currUser, repo.Owners) {
		http.Error(w, fmt.Sprintf("User %s does not have permission to add a base group this repo.", currUser.Mail), http.StatusUnauthorized)
		return
	}

	// Add the base group into the repos reviewer groups
	for reviewGroupName, reviewers := range baseGroup.ReviewerGroups {
		repo.ReviewerGroups[reviewGroupName].Reviewers = append(
			repo.ReviewerGroups[reviewGroupName].Reviewers,
			reviewers.Reviewers...)
	}

	if err := s.RepoStore.UpdateRepository(ctx, repo.ID.Hex(), repo); err != nil {
		http.Error(w, fmt.Sprintf("failed to update repository: '%v'", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// DisableRepository disables a repository from being reviewed
func (s *Server) DisableRepository(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	vars := mux.Vars(req)
	logger := log.G(ctx)

	repoName := vars["repository"]
	if repoName == "" {
		errMsg := "repository name missing from request"
		logger.Errorf("DisableRepository: %v", errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	projectName := vars["project"]
	if projectName == "" {
		errMsg := "project name missing from request"
		logger.Errorf("DisableRepository: %v", errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	repo, err := s.RepoStore.GetRepositoryByName(ctx, repoName, projectName)
	if err != nil {
		if err == store.ErrNotFound {
			http.Error(w, fmt.Sprintf("repository '%s/%s' not found", projectName, repoName), http.StatusBadRequest)
			return
		}
		http.Error(w, fmt.Sprintf("database error: '%v'", err), http.StatusInternalServerError)
	}

	repo.Enabled = false

	if err := s.RepoStore.UpdateRepository(ctx, repo.ID.Hex(), repo); err != nil {
		http.Error(w, fmt.Sprintf("failed to update repository: '%v'", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// EnableRepository enables a repository for review
func (s *Server) EnableRepository(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	vars := mux.Vars(req)
	logger := log.G(ctx)

	repoName := vars["repository"]
	if repoName == "" {
		errMsg := "repository name missing from request"
		logger.Errorf("EnableRepository: %v", errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	projectName := vars["project"]
	if projectName == "" {
		errMsg := "project name missing from request"
		logger.Errorf("EnableRepository: %v", errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	repo, err := s.RepoStore.GetRepositoryByName(ctx, repoName, projectName)
	if err != nil {
		if err == store.ErrNotFound {
			http.Error(w, fmt.Sprintf("repository '%s/%s' not found", projectName, repoName), http.StatusBadRequest)
			return
		}
		http.Error(w, fmt.Sprintf("database error: '%v'", err), http.StatusInternalServerError)
	}

	repo.Enabled = true

	if err := s.RepoStore.UpdateRepository(ctx, repo.ID.Hex(), repo); err != nil {
		http.Error(w, fmt.Sprintf("failed to update repository: '%v'", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// DeleteRepository removes a repository
func (s *Server) DeleteRepository(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	ctx := req.Context()
	logger := log.G(ctx)

	currUser, err := getCurrentUser(ctx)
	if err != nil {
		logger.Errorf("DeleteRepository: %v", err)
		http.Error(w, "failed to get current user", http.StatusBadRequest)
		return
	}

	repoName := vars["repository"]
	if repoName == "" {
		errMsg := "repository name missing from request"
		logger.Errorf("deleteRepository: %v", errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	projectName := vars["project"]
	if projectName == "" {
		errMsg := "project name missing from request"
		logger.Errorf("deleteRepository: %v", errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	repo, err := s.RepoStore.GetRepositoryByName(ctx, repoName, projectName)
	if err != nil {
		if err == store.ErrNotFound {
			http.Error(w, fmt.Sprintf("repository %s/%s not found", projectName, repoName), http.StatusBadRequest)
			return
		}
		http.Error(w, fmt.Sprintf("database error: '%v'", err), http.StatusInternalServerError)
	}

	if !s.userHasWritePermission(currUser, repo.Owners) {
		http.Error(w, fmt.Sprintf("User %s does not have permission to delete this repo.", currUser.Mail), http.StatusUnauthorized)
		return
	}

	if err := s.RepoStore.DeleteRepository(ctx, repo.ID.Hex()); err != nil {
		http.Error(w, fmt.Sprintf("failed to delete repository: '%v'", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// PostRepository creates a new repository
func (s *Server) PostRepository(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	logger := log.G(ctx)

	currUser, err := getCurrentUser(ctx)
	if err != nil {
		logger.Errorf("PostRepository: %v", err)
		http.Error(w, "failed to get current user", http.StatusBadRequest)
		return
	}

	repo, err := getRepositoryFromBody(req)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to parse repository body: '%v'", err), http.StatusBadRequest)
		return
	}

	logger.Infof("Got post repo request: %+v", repo)

	err = validateRepository(repo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	_, err = s.RepoStore.GetRepositoryByName(ctx, repo.Name, repo.ProjectName)
	if err != nil {
		if err == store.ErrNotFound {
			// If createing a new repo, make the creator the owner
			if len(repo.Owners) == 0 {
				repo.Owners = []string{currUser.Mail}
			}

			now := time.Now().UTC()
			repo.Created = &now
			repo.Updated = &now

			// process the reviewers and get their vsts ids
			if err := s.processReviewers(repo.ReviewerGroups); err != nil {
				httpError(ctx, w, err.Error(), http.StatusBadRequest)
				return
			}

			if err := s.RepoStore.AddRepository(ctx, repo); err != nil {
				http.Error(w, fmt.Sprintf("failed to add repository: '%v", err), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusCreated)
			return
		}
		http.Error(w, fmt.Sprintf("database error: '%v'", err), http.StatusInternalServerError)
	}

	http.Error(w, fmt.Sprintf("repository %s/%s already exists", repo.ProjectName, repo.Name), http.StatusBadRequest)
	w.WriteHeader(http.StatusCreated)
	return
}

func (s *Server) processReviewers(reviewerGroups types.ReviewerGroups) error {
	for _, group := range reviewerGroups {
		reviewers := make([]*types.Reviewer, 0)
		for _, reviewer := range group.Reviewers {
			// If the ID is already populated, continue
			if reviewer.ID != "" {
				reviewers = append(reviewers, reviewer)
				continue
			}

			fullReviewer, err := utils.GetReviwerFromAlias(reviewer.Alias, s.vstsClient.RestClient)
			if err != nil {
				log.G(context.TODO()).Errorf("failed to validate reviewer '%s' with err: '%v'", reviewer.Alias, err)
				continue
			}

			reviewers = append(reviewers, fullReviewer)
		}
		group.Reviewers = reviewers
	}
	return nil
}

// PutRepository updates a currently availble repository
func (s *Server) PutRepository(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	ctx := req.Context()
	logger := log.G(ctx)

	currUser, err := getCurrentUser(ctx)
	if err != nil {
		logger.Errorf("PutRepository: %v", err)
		http.Error(w, "failed to get current user", http.StatusBadRequest)
		return
	}

	repoName := vars["repository"]
	if repoName == "" {
		errMsg := "repository name missing from request"
		logger.Errorf("PutRepository: %v", errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	projectName := vars["project"]
	if projectName == "" {
		errMsg := "project name missing from request"
		logger.Errorf("PutRepository: %v", errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	repo, err := getRepositoryFromBody(req)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to parse repository body: '%v'", err), http.StatusBadRequest)
		return
	}

	repoCurr, err := s.RepoStore.GetRepositoryByName(ctx, repoName, projectName)
	if err != nil {
		if err == store.ErrNotFound {
			http.Error(w, fmt.Sprintf("repository %s/%s does not exist", projectName, repoName), http.StatusBadRequest)
			return
		}
		http.Error(w, fmt.Sprintf("database error: '%v'", err), http.StatusInternalServerError)
	}

	// process the reviewers and get their vsts ids
	if err := s.processReviewers(repo.ReviewerGroups); err != nil {
		httpError(ctx, w, err.Error(), http.StatusBadRequest)
		return
	}

	if !s.userHasWritePermission(currUser, repo.Owners) {
		http.Error(w, fmt.Sprintf("User %s does not have write permission for this repo.", currUser.Mail), http.StatusUnauthorized)
		return
	}

	now := time.Now().UTC()
	repo.Updated = &now

	err = s.RepoStore.UpdateRepository(ctx, repoCurr.ID.Hex(), repo)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to update repository: '%v'", err), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
	return
}

// GetRepository gets a single repository
func (s *Server) GetRepository(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	vars := mux.Vars(req)
	logger := log.G(ctx)

	repoName := vars["repository"]
	if repoName == "" {
		errMsg := "repository name missing from request"
		logger.Errorf("GetRepository: %v", errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	projectName := vars["project"]
	if projectName == "" {
		errMsg := "project name missing from request"
		logger.Errorf("GetRepository: %v", errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	repo, err := s.RepoStore.GetRepositoryByName(ctx, repoName, projectName)
	if err != nil {
		if err == store.ErrNotFound {
			http.Error(w, fmt.Sprintf("repository: '%s/%s' not found", projectName, repoName), http.StatusBadRequest)
			return
		}
		http.Error(w, fmt.Sprintf("database error: '%v'", err), http.StatusInternalServerError)
	}

	err = json.NewEncoder(w).Encode(repo)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to marshal request: '%v'", err), http.StatusInternalServerError)
		return
	}
}

// GetRepositoryPerProject gets all repositories from a project
func (s *Server) GetRepositoryPerProject(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	vars := mux.Vars(req)

	projectName := vars["project"]
	if projectName == "" {
		errMsg := "project name missing from request"
		httpError(ctx, w, errMsg, http.StatusBadRequest)
		return
	}

	repos, err := s.RepoStore.GetAllRepositories(ctx)
	if err != nil {
		httpError(ctx, w, fmt.Sprintf("failed to get repositories: '%v'", err), http.StatusInternalServerError)
		return
	}

	var result []*types.Repository
	for _, repo := range repos {
		if repo.ProjectName == projectName {
			result = append(result, repo)
		}
	}

	err = json.NewEncoder(w).Encode(result)
	if err != nil {
		httpError(ctx, w, fmt.Sprintf("failed to marshal request: '%v'", err), http.StatusInternalServerError)
		return
	}
}

// GetRepositories gets all avaible repositories
func (s *Server) GetRepositories(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	repos, err := s.RepoStore.GetAllRepositories(ctx)
	if err != nil {
		httpError(ctx, w, fmt.Sprintf("failed to get repositories: '%v'", err), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(repos)
	if err != nil {
		httpError(ctx, w, fmt.Sprintf("failed to marshal request: '%v'", err), http.StatusInternalServerError)
		return
	}
}

func validateRepository(repo *types.Repository) error {
	if repo.Name == "" {
		return fmt.Errorf("name is a required field for repository")
	}

	if repo.ProjectName == "" {
		return fmt.Errorf("ProjectName is a required field for repository")
	}

	return nil
}

func getRepositoryFromBody(req *http.Request) (*types.Repository, error) {
	var repo types.Repository
	err := json.NewDecoder(req.Body).Decode(&repo)
	if err != nil {
		return nil, err
	}

	return &repo, nil
}

func getReviewerFromBody(req *http.Request) (*types.Reviewer, error) {
	var reviewer types.Reviewer
	err := json.NewDecoder(req.Body).Decode(&reviewer)
	if err != nil {
		return nil, err
	}

	return &reviewer, nil
}

func getCurrentUser(ctx context.Context) (*types.GraphUser, error) {
	user, ok := ctx.Value(userInfoContextKey).(*types.GraphUser)
	if !ok {
		return nil, errors.New("no current user in the context")
	}

	return user, nil
}

func httpError(ctx context.Context, w http.ResponseWriter, msg string, code int) {
	log.G(ctx).Error(msg)
	http.Error(w, msg, code)
}

func (s *Server) userHasWritePermission(user *types.GraphUser, owners []string) bool {
	// Allow full access for admins
	if contains(s.Admins, user.Mail) {
		return true
	}

	return contains(owners, user.Mail)
}

func contains(list []string, key string) bool {
	for _, elem := range list {
		if elem == key {
			return true
		}
	}

	return false
}
