package autoreviewer

import (
	"context"
	"github.com/samkreter/go-core/log"
	"path/filepath"
	"strings"

	adogit "github.com/microsoft/azure-devops-go-api/azuredevops/git"
	"github.com/pkg/errors"
)

const (
	PrefixComment = ";"
	PrefixNoNotify = "*"
	PrefixGroup = "; TEAM: "
)

var (
	errNotFound = errors.New("resource not found")
)

type PullRequest struct {
	adogit.GitPullRequest
}

type ReviewerGroup struct {
	Owners map[string]bool
	Teams  map[string]bool
}

// GetRequiredReviewerGroups gets all required reviewers from the owners files based on changes made in the PR.
// TODO: Use cache for finding the owners files.
func (pr *PullRequest) GetRequiredReviewerGroups(ctx context.Context,  client adogit.Client) ([]*ReviewerGroup, error) {
	ownerFilesMap  := map[string]*ReviewerGroup{}

	changePaths, err := pr.GetAllChanges(ctx, client)
	if err != nil {
		return nil, err
	}

	for _, path := range changePaths {
		pathDir := filepath.Dir(path)

		if ownerFilesMap[pathDir] == nil {
			ownersFile, err := getRelatedOwnersFile(ctx, client, pr.Repository.Id.String(), path)
			if err != nil {
				return nil, err
			}

			ownerFilesMap[pathDir] = newReviewerGroupFromOwnersFile(*ownersFile.Content)
		}
	}

	reviewerGroups := make([]*ReviewerGroup, len(ownerFilesMap))
	for _, reviewerGroup := range ownerFilesMap {
		reviewerGroups = append(reviewerGroups, reviewerGroup)
	}

	return reviewerGroups, nil
}

// getRelatedOwnersFile returns the owners file that impacts this change. This is the owners file that is
// the closest directory moving upwards.
func getRelatedOwnersFile(ctx context.Context, client adogit.Client, repoID, path string) (*adogit.GitItem, error) {
	dirPath := filepath.Dir(path)
	ownersPath := filepath.Join(dirPath, "owners.txt")

	ownersFile, err := getFileFromADO(ctx, client, repoID, ownersPath)
	if err != nil {
		switch {
		case errors.Is(err, errNotFound):
			if dirPath == "/" {
				return nil, errNotFound
			}
			return getRelatedOwnersFile(ctx, client, repoID, dirPath)
		default:
			return nil, err
		}
	}

	return ownersFile, nil
}

// GetAllChanges returns all changes from all iterations of the pull request
func (pr *PullRequest) GetAllChanges(ctx context.Context, client adogit.Client) ([]string, error) {
	logger := log.G(ctx)
	repositoryID := pr.Repository.Id.String()
	its, err := client.GetPullRequestIterations(ctx, adogit.GetPullRequestIterationsArgs{
		RepositoryId:  &repositoryID,
		PullRequestId: pr.PullRequestId,
	})
	if err != nil {
		return nil, ParseADOError(err)
	}

	iterationID := len(*its)
	var paths []string
	nextSkipToken := 0
	// TODO: Add timeout here
	for {
		changes, err := client.GetPullRequestIterationChanges(ctx, adogit.GetPullRequestIterationChangesArgs{
			RepositoryId:  &repositoryID,
			PullRequestId: pr.PullRequestId,
			IterationId:   &iterationID,
			Skip: &nextSkipToken,
		})
		if err != nil {
			return nil, ParseADOError(err)
		}

		localPaths, err := getChangePaths(ctx, *changes.ChangeEntries)
		if err != nil {
			return nil, err
		}

		if localPaths != nil {
			paths = append(paths, localPaths...)
		}

		if changes.NextSkip == nil || *changes.NextSkip == 0 {
			return paths, nil
		}
		nextSkipToken = *changes.NextSkip
		logger.Infof("Getting skip token Changes with skip token %d", nextSkipToken)
	}
}

// newReviewerGroupFromOwnersFile creates a reviewerGroup from an owners file
func newReviewerGroupFromOwnersFile(content string) *ReviewerGroup {
	lines := strings.Split(content, "\n")

	reviewerGroup := ReviewerGroup{
		Owners: map[string]bool{},
		Teams: map[string]bool{},
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		switch {
		// Parse the reviewer group
		case strings.HasPrefix(line, PrefixGroup):
			team := strings.TrimSpace(strings.TrimPrefix(line, PrefixGroup))
			reviewerGroup.Teams[team] = true
			continue

		// Parse owners with the no notify prefix
		case strings.HasPrefix(line, PrefixNoNotify):
			owner := strings.TrimSpace(strings.TrimPrefix(line, PrefixNoNotify))
			reviewerGroup.Owners[owner] = true
			continue

		// Ignore comments
		case strings.HasPrefix(line, PrefixComment):
			continue

		// Assume line without prefix is owner
		default:
			reviewerGroup.Owners[line] = true
		}
	}

	return &reviewerGroup
}


func toBoolPtr(val bool) *bool {
	return &val
}
