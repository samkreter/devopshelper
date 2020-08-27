package autoreviewer

import (
	"fmt"
	"github.com/microsoft/azure-devops-go-api/azuredevops"
	"path/filepath"
	"context"
	"errors"
	"strings"

	adogit "github.com/microsoft/azure-devops-go-api/azuredevops/git"
)

const (
	PrefixComment = ";"
	PrefixNoNotify = "*"
	PrefixGroup = "; GROUP: "
)

var (
	errNotFound = errors.New("resource not found")
)

type PullRequest struct {
	adogit.GitPullRequest
}

type ReviewerGroup struct {
	Owners map[string]bool
	Group string
}

// GetReviewerGroups gets all required reviewers from the owners files based on file changes in the PR.
// TODO: Use cache for finding the owners files.
func (pr *PullRequest) GetReviewerGroups(ctx context.Context,  client adogit.Client) ([]*ReviewerGroup, error) {
	ownerFilesMap  := map[string]*ReviewerGroup{}

	changePaths, err := pr.GetAllChanges(ctx, client)
	if err != nil {
		return nil, err
	}

	for _, path := range changePaths {
		pathDir := filepath.Dir(path)
		if ownerFilesMap[pathDir] == nil {
			ownersFile, err := getOwnersFile(ctx, client, pr.Repository.Id.String(), path)
			if err != nil {
				return nil, err
			}

			ownerFilesMap[pathDir] = ParseOwnerFile(*ownersFile.Content)
		}
	}

	reviewerGroups := make([]*ReviewerGroup, len(ownerFilesMap))
	for _, reviewerGroup := range ownerFilesMap {
		reviewerGroups = append(reviewerGroups, reviewerGroup)
	}

	return reviewerGroups, nil
}

func getOwnersFile(ctx context.Context, client adogit.Client, repoID, path string) (*adogit.GitItem, error) {
	dirPath := filepath.Dir(path)
	ownersPath := filepath.Join(dirPath, "owners.txt")

	item, err := client.GetItem(ctx, adogit.GetItemArgs{
		RepositoryId:   &repoID,
		Path:           &ownersPath,
		IncludeContent: toBoolPtr(true),
	})
	if err != nil {
		switch ParseADOError(err) {
		case errNotFound:
			if dirPath == "/" {
				return nil, errNotFound
			}
			return getOwnersFile(ctx, client, repoID, dirPath)
		default:
			return nil, err
		}
	}

	return item, nil
}

func (pr *PullRequest) GetAllChanges(ctx context.Context, client adogit.Client) ([]string, error) {
	repositoryID := pr.Repository.Id.String()
	its, err := client.GetPullRequestIterations(ctx, adogit.GetPullRequestIterationsArgs{
		RepositoryId:  &repositoryID,
		PullRequestId: pr.PullRequestId,
	})
	if err != nil {
		return nil, err
	}

	iterationID := len(*its)
	changes, err := client.GetPullRequestIterationChanges(ctx, adogit.GetPullRequestIterationChangesArgs{
		RepositoryId:  &repositoryID,
		PullRequestId: pr.PullRequestId,
		IterationId:   &iterationID,
	})
	if err != nil {
		return nil, err
	}

	if changes.NextSkip != nil && *changes.NextSkip != 0 {
		return nil, fmt.Errorf("next skiptoken is not 0, requires pagination") // TODO: handle pagination
	}

	return getChangePaths(*changes.ChangeEntries)
}

func getChangePaths(changes []adogit.GitPullRequestChange) ([]string, error) {
	var paths []string

	for _, change := range changes {
		item, ok := change.Item.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("failed to cast Item")
		}

		path, ok := item["path"].(string)
		if !ok {
			return nil, fmt.Errorf("failed to cast path")
		}

		paths = append(paths, path)
	}

	return paths, nil
}

func ParseOwnerFile(content string) *ReviewerGroup {
	lines := strings.Split(content, "\n")

	ownersFile := ReviewerGroup{
		Owners: map[string]bool{},
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		switch {
		// Parse the reviewer group
		case strings.HasPrefix(line, PrefixGroup):
			ownersFile.Group = strings.TrimSpace(strings.TrimPrefix(line, PrefixGroup))
			continue

		// Parse owners with the no notify prefix
		case strings.HasPrefix(line, PrefixNoNotify):
			owner := strings.TrimSpace(strings.TrimPrefix(line, PrefixNoNotify))
			ownersFile.Owners[owner] = true
			continue

		// Ignore comments
		case strings.HasPrefix(line, PrefixComment):
			continue

		// Assume line without prefix is owner
		default:
			ownersFile.Owners[line] = true
		}
	}

	return &ownersFile
}

func ParseADOError(err error) error {
	adoErr, ok := err.(azuredevops.WrappedError)
	if !ok {
		return err
	}

	if adoErr.StatusCode != nil && *adoErr.StatusCode == 404 {
		return errNotFound
	}

	return err
}


func toBoolPtr(val bool) *bool {
	return &val
}
