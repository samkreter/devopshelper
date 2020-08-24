package utils

import (
	"fmt"
	"context"

	"github.com/samkreter/devopshelper/pkg/types"
	adoIdentity "github.com/microsoft/azure-devops-go-api/azuredevops/identity"
)


var (
	DirectoryAliasIdentitySearchFilter = "DirectoryAlias"
)


// GetReviewerFromAlias gets a reviewer from an alias
func GetReviewerFromAlias(ctx context.Context, alias string, adoIdentityClient adoIdentity.Client) (*types.Reviewer, error) {
	identity, err := GetDevOpsIdentity(ctx, alias, adoIdentityClient)
	if err != nil {
		return nil, err
	}

	return &types.Reviewer{
		Alias: alias,
		ID:    identity.Id.String(),
	}, nil
}

// GetDevOpsIdentity returns the azure devops identity for a given alais
func GetDevOpsIdentity(ctx context.Context, alias string, adoIdentityClient adoIdentity.Client) (*adoIdentity.Identity, error) {
	identities, err := adoIdentityClient.ReadIdentities(ctx, adoIdentity.ReadIdentitiesArgs{
		SearchFilter: &DirectoryAliasIdentitySearchFilter,
		FilterValue:  &alias,
	})
	if err != nil {
		return nil, err
	}

	if len(*identities) != 1 {
		return nil, fmt.Errorf("Found multiple identities for alias '%s'", alias)
	}

	return &(*identities)[0], nil
}
