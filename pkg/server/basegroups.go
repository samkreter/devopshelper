package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/samkreter/vstsautoreviewer/pkg/store"
	"github.com/samkreter/vstsautoreviewer/pkg/types"
)

var (
	// ErrBaseGroupNameNotFound the base group name not found in the request
	ErrBaseGroupNameNotFound = errors.New("missing baseGroupName in reqeust")
)

func (s *Server) handleDeleteBaseGroup(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	ctx := context.Background()

	baseGroupName := vars["baseGroupName"]
	if baseGroupName == "" {
		http.Error(w, ErrBaseGroupNameNotFound.Error(), http.StatusBadRequest)
		return
	}

	baseGroup, err := s.RepoStore.GetBaseGroupByName(ctx, baseGroupName)
	if err != nil {
		if err == store.ErrNotFound {
			w.WriteHeader(http.StatusOK)
			return
		}
		http.Error(w, fmt.Sprintf("database error: '%v'", err), http.StatusInternalServerError)
	}

	if err := s.RepoStore.DeleteBaseGroup(ctx, baseGroup.ID.Hex()); err != nil {
		http.Error(w, fmt.Sprintf("failed to delete basegroupd: '%v'", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) handlePutBaseGroup(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	ctx := context.Background()

	baseGroupName := vars["baseGroupName"]
	if baseGroupName == "" {
		http.Error(w, ErrBaseGroupNameNotFound.Error(), http.StatusBadRequest)
		return
	}

	baseGroup, err := getBaseGroupFromBody(req)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to parse base group body: '%v'", err), http.StatusBadRequest)
		return
	}

	originalBaseGroup, err := s.RepoStore.GetBaseGroupByName(ctx, baseGroupName)
	if err != nil {
		if err == store.ErrNotFound {
			if err := s.RepoStore.AddBaseGroup(ctx, baseGroup.Name, baseGroup); err != nil {
				http.Error(w, fmt.Sprintf("failed to add base group: '%v", err), http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
			return
		}
		http.Error(w, fmt.Sprintf("database error: '%v'", err), http.StatusInternalServerError)
	}

	err = s.RepoStore.UpdateBaseGroup(ctx, originalBaseGroup.ID.Hex(), baseGroup)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to update base group: '%v'", err), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
	return
}

func getBaseGroupFromBody(req *http.Request) (*types.BaseGroup, error) {
	var baseGroup types.BaseGroup
	err := json.NewDecoder(req.Body).Decode(&baseGroup)
	if err != nil {
		return nil, err
	}

	return &baseGroup, nil
}

func (s *Server) handleGetBaseGroup(w http.ResponseWriter, req *http.Request) {
	ctx := context.Background()
	vars := mux.Vars(req)

	baseGroupName := vars["baseGroupName"]
	if baseGroupName == "" {
		http.Error(w, ErrBaseGroupNameNotFound.Error(), http.StatusBadRequest)
		return
	}

	baseGroup, err := s.RepoStore.GetBaseGroupByName(ctx, baseGroupName)
	if err != nil {
		if err == store.ErrNotFound {
			http.Error(w, fmt.Sprintf("base group: '%s' not found", baseGroupName), http.StatusBadRequest)
			return
		}
		http.Error(w, fmt.Sprintf("database error: '%v'", err), http.StatusInternalServerError)
	}

	err = json.NewEncoder(w).Encode(baseGroup)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to marshal request: '%v'", err), http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleGetBaseGroups(w http.ResponseWriter, req *http.Request) {
	ctx := context.Background()

	baseGroups, err := s.RepoStore.GetAllBaseGroups(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get base groups: '%v'", err), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(baseGroups)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to marshal request: '%v'", err), http.StatusInternalServerError)
		return
	}
}
