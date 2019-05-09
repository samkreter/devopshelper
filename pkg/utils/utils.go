package utils

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/samkreter/devopshelper/pkg/types"
)

const (
	identityURL = "https://vssps.dev.azure.com/mseng/_apis/identities?searchFilter=DirectoryAlias&filterValue=%s&api-version=5.0"
)

type IdentityResponse struct {
	Count int64      `json:"count"`
	Value []Identity `json:"value"`
}

type Identity struct {
	ID                  string              `json:"id"`
	Descriptor          string              `json:"descriptor"`
	SubjectDescriptor   string              `json:"subjectDescriptor"`
	ProviderDisplayName string              `json:"providerDisplayName"`
	IsActive            bool                `json:"isActive"`
	Members             []interface{}       `json:"members"`
	MemberOf            []interface{}       `json:"memberOf"`
	MemberIDS           []interface{}       `json:"memberIds"`
	Properties          map[string]Property `json:"properties"`
	ResourceVersion     int64               `json:"resourceVersion"`
	MetaTypeID          int64               `json:"metaTypeId"`
}

type Property struct {
	Type  Type   `json:"$type"`
	Value string `json:"$value"`
}

type Type string

const (
	SystemDateTime Type = "System.DateTime"
	SystemString   Type = "System.String"
)

type devOpsClient interface {
	Do(method, url string, body io.Reader) ([]byte, error)
}

// GetReviwerFromAlias gets a reviewer from an alias
func GetReviwerFromAlias(alias string, client devOpsClient) (*types.Reviewer, error) {
	identity, err := GetDevOpsIdentity(alias, client)
	if err != nil {
		return nil, err
	}

	return &types.Reviewer{
		Alias: alias,
		ID:    identity.ID,
	}, nil
}

// GetDevOpsIdentity returns the azure devops identity for a given alais
func GetDevOpsIdentity(alias string, client devOpsClient) (*Identity, error) {
	b, err := client.Do("GET", fmt.Sprintf(identityURL, alias), nil)
	if err != nil {
		return nil, err
	}

	var identityResp IdentityResponse
	if err := json.Unmarshal(b, &identityResp); err != nil {
		return nil, err
	}

	if identityResp.Count != 1 {
		return nil, fmt.Errorf("Found multiple identities for alias '%s'", alias)
	}

	return &identityResp.Value[0], nil
}
