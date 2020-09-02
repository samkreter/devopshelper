package autoreviewer

import (
	"context"
	"strings"

	"github.com/microsoft/azure-devops-go-api/azuredevops"
	adocore "github.com/microsoft/azure-devops-go-api/azuredevops/core"
	adogit "github.com/microsoft/azure-devops-go-api/azuredevops/git"
	"github.com/pkg/errors"
)


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

func parseEmailToAlias(email string ) string {
	emailPtrs := strings.Split(email, "@")
	if len(emailPtrs) != 2 {
		return ""
	}

	return strings.TrimSpace(emailPtrs[0])
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
