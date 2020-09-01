package utils

import (
	"context"
	"fmt"
	"strings"
	"time"

	adoIdentity "github.com/microsoft/azure-devops-go-api/azuredevops/identity"
	"github.com/samkreter/devopshelper/pkg/types"
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
		AdoID: identity.Id.String(),
		LastReviewTime: time.Time{},
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

	if len(*identities) == 0 {
		return nil, fmt.Errorf("no ado identities found for alias: %s", alias)
	}

	if len(*identities) == 1 {
		return &(*identities)[0], nil
	}

	for _, identity := range *identities {
		identityAlias, err := GetIdentityAlias(identity)
		if err != nil {
			return nil,  err
		}
		if strings.EqualFold(identityAlias, alias) {
			return &identity, nil
		}
	}

	return nil, fmt.Errorf("no ado identities found for alias: %s", alias)
}

func GetIdentityAlias(identity adoIdentity.Identity) (string, error) {
	properties, ok := identity.Properties.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("malformed identity")
	}
	directory, ok := properties["DirectoryAlias"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("malformed identity")
	}

	alias, ok := directory["$value"].(string)
	if !ok {
		return "", fmt.Errorf("malformed identity")
	}

	return alias, nil
}
