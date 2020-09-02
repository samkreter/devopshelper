package autoreviewer

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/microsoft/azure-devops-go-api/azuredevops"
	adocore "github.com/microsoft/azure-devops-go-api/azuredevops/core"
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
	repositoryID := pr.Repository.Id.String()
	its, err := client.GetPullRequestIterations(ctx, adogit.GetPullRequestIterationsArgs{
		RepositoryId:  &repositoryID,
		PullRequestId: pr.PullRequestId,
	})
	if err != nil {
		return nil, ParseADOError(err)
	}

	iterationID := len(*its)
	changes, err := client.GetPullRequestIterationChanges(ctx, adogit.GetPullRequestIterationChangesArgs{
		RepositoryId:  &repositoryID,
		PullRequestId: pr.PullRequestId,
		IterationId:   &iterationID,
	})
	if err != nil {
		return nil, ParseADOError(err)
	}

	if changes.NextSkip != nil && *changes.NextSkip != 0 {
		return nil, errors.New("next skiptoken is not 0, requires pagination") // TODO: handle pagination
	}

	return getChangePaths(*changes.ChangeEntries)
}

// getChangePaths returns a slice of paths for each change
func getChangePaths(changes []adogit.GitPullRequestChange) ([]string, error) {
	var paths []string

	for _, change := range changes {
		item, ok := change.Item.(map[string]interface{})
		if !ok {
			return nil, errors.New("failed to cast Item")
		}

		path, ok := item["path"].(string)
		if !ok {
			return nil, errors.New("failed to cast path")
		}

		paths = append(paths, path)
	}

	return paths, nil
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


// getFileFromADO retrieves a file and its contents from ADO
func getFileFromADO(ctx context.Context, client adogit.Client, repoID, path string) (*adogit.GitItem, error) {
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

func parseEmailToAlias(email string ) string {
	emailPtrs := strings.Split(email, "@")
	if len(emailPtrs) != 2 {
		return ""
	}

	return strings.TrimSpace(emailPtrs[0])
}


// getTeamMemberADOIDs returns the ADO IDs for team members
func (a *AutoReviewer) getTeamMembers(ctx context.Context, teamName string) ([]string, error){
	members, err := a.adoCoreClient.GetTeamMembersWithExtendedProperties(ctx, adocore.GetTeamMembersWithExtendedPropertiesArgs{
		ProjectId: &a.Repo.ProjectName,
		TeamId:    &teamName,
	})
	if err != nil {
		return nil, ParseADOError(err)
	}

	if members == nil {
		return []string{}, nil
	}

	var aliases []string
	for _, member := range *members {
		if member.Identity != nil && member.Identity.UniqueName != nil {
			if alias := parseEmailToAlias(*member.Identity.UniqueName); alias != "" {
				aliases = append(aliases, alias)
			}
		}
	}

	return aliases, nil
}

// ParseADOError converts ADO errors to internal errors
func ParseADOError(err error) error {
	adoErr, ok := err.(azuredevops.WrappedError)
	if !ok {
		return errors.WithStack(err)
	}

	if adoErr.StatusCode != nil && *adoErr.StatusCode == 404 {
		return  errors.WithStack(errNotFound)
	}

	return  errors.WithStack(err)
}

func toBoolPtr(val bool) *bool {
	return &val
}
