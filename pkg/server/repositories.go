package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/samkreter/vstsautoreviewer/pkg/store"
	"github.com/samkreter/vstsautoreviewer/pkg/types"
	"github.com/samkreter/vstsautoreviewer/pkg/utils"
)

func (s *Server) handleGetReviewerGroupToRepository(w http.ResponseWriter, req *http.Request) {
	ctx := context.Background()
	vars := mux.Vars(req)

	repoName := vars["repository"]
	if repoName == "" {
		http.Error(w, "repository name missing from request", http.StatusBadRequest)
		return
	}

	projectName := vars["project"]
	if repoName == "" {
		http.Error(w, "project name missing from request", http.StatusBadRequest)
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

func (s *Server) handleDeleteReviewerToRepository(w http.ResponseWriter, req *http.Request) {
	ctx := context.Background()
	vars := mux.Vars(req)

	repoName := vars["repository"]
	if repoName == "" {
		http.Error(w, "repository name missing from request", http.StatusBadRequest)
		return
	}

	projectName := vars["project"]
	if repoName == "" {
		http.Error(w, "project name missing from request", http.StatusBadRequest)
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

func (s *Server) handleAddReviewerToRepository(w http.ResponseWriter, req *http.Request) {
	ctx := context.Background()
	vars := mux.Vars(req)

	repoName := vars["repository"]
	if repoName == "" {
		http.Error(w, "repository name missing from request", http.StatusBadRequest)
		return
	}

	projectName := vars["project"]
	if repoName == "" {
		http.Error(w, "project name missing from request", http.StatusBadRequest)
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

func (s *Server) handleAddBaseGroupToRepository(w http.ResponseWriter, req *http.Request) {
	ctx := context.Background()
	vars := mux.Vars(req)

	repoName := vars["repository"]
	if repoName == "" {
		http.Error(w, "repository name missing from request", http.StatusBadRequest)
		return
	}

	projectName := vars["project"]
	if repoName == "" {
		http.Error(w, "project name missing from request", http.StatusBadRequest)
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

func (s *Server) handleDisableRepository(w http.ResponseWriter, req *http.Request) {
	ctx := context.Background()
	vars := mux.Vars(req)

	repoName := vars["repository"]
	if repoName == "" {
		http.Error(w, "repository name missing from request", http.StatusBadRequest)
		return
	}

	projectName := vars["project"]
	if repoName == "" {
		http.Error(w, "project name missing from request", http.StatusBadRequest)
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

func (s *Server) handleEnableRepository(w http.ResponseWriter, req *http.Request) {
	ctx := context.Background()
	vars := mux.Vars(req)

	repoName := vars["repository"]
	if repoName == "" {
		http.Error(w, "repository name missing from request", http.StatusBadRequest)
		return
	}

	projectName := vars["project"]
	if repoName == "" {
		http.Error(w, "project name missing from request", http.StatusBadRequest)
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

func (s *Server) handleDeleteRepository(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	ctx := context.Background()

	repoName := vars["repository"]
	if repoName == "" {
		http.Error(w, "repository name missing from request", http.StatusBadRequest)
		return
	}

	projectName := vars["project"]
	if repoName == "" {
		http.Error(w, "project name missing from request", http.StatusBadRequest)
		return
	}

	repo, err := s.RepoStore.GetRepositoryByName(ctx, repoName, projectName)
	if err != nil {
		if err == store.ErrNotFound {
			w.WriteHeader(http.StatusOK)
			return
		}
		http.Error(w, fmt.Sprintf("database error: '%v'", err), http.StatusInternalServerError)
	}

	if err := s.RepoStore.DeleteRepository(ctx, repo.ID.Hex()); err != nil {
		http.Error(w, fmt.Sprintf("failed to delete repository: '%v'", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) handlePostRepository(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	ctx := context.Background()

	repoName := vars["repository"]
	if repoName == "" {
		http.Error(w, "repository name missing from request", http.StatusBadRequest)
		return
	}

	projectName := vars["project"]
	if repoName == "" {
		http.Error(w, "project name missing from request", http.StatusBadRequest)
		return
	}

	repo, err := getRepositoryFromBody(req)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to parse repository body: '%v'", err), http.StatusBadRequest)
		return
	}

	err = validateRepository(repo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	_, err = s.RepoStore.GetRepositoryByName(ctx, repoName, projectName)
	if err != nil {
		if err == store.ErrNotFound {
			if err := s.RepoStore.AddRepository(ctx, repo); err != nil {
				http.Error(w, fmt.Sprintf("failed to add repository: '%v", err), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			return
		}
		http.Error(w, fmt.Sprintf("database error: '%v'", err), http.StatusInternalServerError)
	}

	http.Error(w, fmt.Sprintf("repository %s/%s already exists", projectName, repoName), http.StatusBadRequest)
	w.WriteHeader(http.StatusOK)
	return
}

func (s *Server) handlePutRepository(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	ctx := context.Background()

	repoName := vars["repository"]
	if repoName == "" {
		http.Error(w, "repository name missing from request", http.StatusBadRequest)
		return
	}

	projectName := vars["project"]
	if repoName == "" {
		http.Error(w, "project name missing from request", http.StatusBadRequest)
		return
	}

	repo, err := getRepositoryFromBody(req)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to parse repository body: '%v'", err), http.StatusBadRequest)
		return
	}

	_, err = s.RepoStore.GetRepositoryByName(ctx, repoName, projectName)
	if err != nil {
		if err == store.ErrNotFound {
			if err := s.RepoStore.AddRepository(ctx, repo); err != nil {
				http.Error(w, fmt.Sprintf("failed to add repository: '%v", err), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			return
		}
		http.Error(w, fmt.Sprintf("database error: '%v'", err), http.StatusInternalServerError)
	}

	err = s.RepoStore.UpdateRepository(ctx, repo.ID.Hex(), repo)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to update repository: '%v'", err), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
	return
}

func (s *Server) handleGetRepository(w http.ResponseWriter, req *http.Request) {
	ctx := context.Background()
	vars := mux.Vars(req)

	repoName := vars["repository"]
	if repoName == "" {
		http.Error(w, "repository name missing from request", http.StatusBadRequest)
		return
	}

	projectName := vars["project"]
	if repoName == "" {
		http.Error(w, "project name missing from request", http.StatusBadRequest)
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

func (s *Server) handleGetRepositoryPerProject(w http.ResponseWriter, req *http.Request) {
	ctx := context.Background()
	vars := mux.Vars(req)

	projectName := vars["project"]
	if projectName == "" {
		http.Error(w, "project name missing from request", http.StatusBadRequest)
		return
	}

	repos, err := s.RepoStore.GetAllRepositories(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get repositories: '%v'", err), http.StatusInternalServerError)
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
		http.Error(w, fmt.Sprintf("failed to marshal request: '%v'", err), http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleGetRepositories(w http.ResponseWriter, req *http.Request) {
	ctx := context.Background()

	repos, err := s.RepoStore.GetAllRepositories(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get repositories: '%v'", err), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(repos)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to marshal request: '%v'", err), http.StatusInternalServerError)
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
