package store

import (
	"context"
	"github.com/pkg/errors"
	"github.com/samkreter/devopshelper/pkg/types"
)

var (
	// ErrNotFound the error is not found
	ErrNotFound = errors.New("record not found")
)


type ReviewerStore interface {
	PopLRUReviewer(ctx context.Context, alias []string) (*types.Reviewer, error)
	GetLRUReviewer(ctx context.Context, alias []string) (*types.Reviewer, error)
	AddReviewer(ctx context.Context, reviewer *types.Reviewer) error
	GetReviewer(ctx context.Context, alias string) (*types.Reviewer, error)
	GetReviewerByADOID(ctx context.Context, adoID string) (*types.Reviewer, error)
	UpdateReviewer(ctx context.Context, reviewer *types.Reviewer) error
}

type TeamStore interface {
	AddTeam(ctx context.Context, team *types.Team) error
	GetTeam(ctx context.Context, alias string) (*types.Team, error)
	UpdateTeam(ctx context.Context, team *types.Team) error
}

// RepositoryStore holds information for a repository
type RepositoryStore interface {
	AddRepository(ctx context.Context, repo *types.Repository) error
	UpdateRepository(ctx context.Context, id string, repository *types.Repository) error
	DeleteRepository(ctx context.Context, id string) error
	GetRepositoryByID(ctx context.Context, id string) (*types.Repository, error)
	GetAllRepositories(ctx context.Context) ([]*types.Repository, error)
	GetRepositoryByName(ctx context.Context, name, project string) (*types.Repository, error)
}