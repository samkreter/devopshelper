package autoreviewer

import (
	"fmt"
	"github.com/microsoft/azure-devops-go-api/azuredevops"
	adocore "github.com/microsoft/azure-devops-go-api/azuredevops/core"
	"path/filepath"
	"context"
	"errors"
	"strings"

	adogit "github.com/microsoft/azure-devops-go-api/azuredevops/git"
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
	Team   string
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
			ownersFile, err := getEffectingOwnersFile(ctx, client, pr.Repository.Id.String(), path)
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
	item, err := client.GetItem(ctx, adogit.GetItemArgs{
		RepositoryId:   &repoID,
		Path:           &path,
		IncludeContent: toBoolPtr(true),
	})
	if err != nil {
		return nil, ParseADOError(err)
	}

	return item, nil
}

func getEffectingOwnersFile(ctx context.Context, client adogit.Client, repoID, path string) (*adogit.GitItem, error) {
	dirPath := filepath.Dir(path)
	ownersPath := filepath.Join(dirPath, "owners.txt")

	ownersFile, err := getOwnersFile(ctx, client, repoID, ownersPath)
	if err != nil {
		switch ParseADOError(err) {
		case errNotFound:
			if dirPath == "/" {
				return nil, errNotFound
			}
			return getEffectingOwnersFile(ctx, client, repoID, dirPath)
		default:
			return nil, err
		}
	}

	return ownersFile, nil
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

func getTeamMembers(ctx context.Context, client adocore.Client, projectName, teamName string) ([]string, error){
	members, err := client.GetTeamMembersWithExtendedProperties(ctx, adocore.GetTeamMembersWithExtendedPropertiesArgs{
		ProjectId: &projectName,
		TeamId:    &teamName,
	})
	if err != nil {
		return nil, err
	}

	reviewers := []string{}
	for _, member := range *members {
		reviewers = append(reviewers, *member.Identity.DirectoryAlias)
	}

	return reviewers, nil
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
			ownersFile.Team = strings.TrimSpace(strings.TrimPrefix(line, PrefixGroup))
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
