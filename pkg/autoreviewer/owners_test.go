package autoreviewer


import (
	"bytes"
	"log"
	"testing"
	"text/template"
	"github.com/stretchr/testify/assert"
)


func TestParseEmailToAlias(t *testing.T){
	tests := []struct {
		Name string
		Email string
		ExpectedAlias string
	}{
		{
			Name: "Success Case",
			Email:         "test@gmail.com",
			ExpectedAlias: "test",
		},
		{
			Name: "Invalid Email: No Separator",
			Email: "test",
			ExpectedAlias: "",
		},
		{
			Name: "Invalid Email: Extra Separators",
			Email: "test@tester@gmail.com",
			ExpectedAlias: "",
		},
		{
			Name: "Empty Email",
			Email: "",
			ExpectedAlias: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			alias := parseEmailToAlias(tt.Email)
			assert.Equal(t, tt.ExpectedAlias, alias, "Should get expected alias.")
		})
	}
}

func TestParseOwnerFile(t *testing.T) {
	tests := []struct {
		Name string
		TestOwners []string
		ExpectedOwners []string
		ExpectedGroup string
	}{
		{Name: "Success Case",
			TestOwners: []string{PrefixGroup+"Test Team", PrefixNoNotify+"testOwenr", "testNoNotifyOwner", ";this is a comment", ""},
			ExpectedOwners: []string {"testOwenr", "testNoNotifyOwner"},
			ExpectedGroup: "Test Team",
		},
		{Name: "No Team",
			TestOwners: []string{"testOwenr", "testNoNotifyOwner", ";this is a comment", ""},
			ExpectedOwners: []string {"testOwenr", "testNoNotifyOwner"},
			ExpectedGroup: "",
		},
		{Name: "No Owners",
			TestOwners: []string{PrefixGroup+"Test Team",";this is a comment", ""},
			ExpectedOwners: []string {},
			ExpectedGroup: "Test Team",
		},
		{Name: "No Owners or Groups",
			TestOwners: []string{";this is a comment", ""},
			ExpectedOwners: []string {},
			ExpectedGroup: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			ownersFile := generateTestOwnersFile(tt.TestOwners)
			reviewerGroup := newReviewerGroupFromOwnersFile(ownersFile)


			assert.Equal(t,tt.ExpectedGroup, reviewerGroup.Team, "Team should be equal")
			assert.Equal(t, len(tt.ExpectedOwners), len(reviewerGroup.Owners), "Should have expected number owners")

			for _, expectOwner := range tt.ExpectedOwners {
				assert.True(t, reviewerGroup.Owners[expectOwner], "Should have correct owners")
			}
		})
	}
}

func generateTestOwnersFile(owners []string) string{
	ownersFileTmpl := `
    {{range $owner := .}}
    {{$owner}}
    {{end}}
	`

	tpl, err := template.New("ownersfile").Parse(ownersFileTmpl)
	if err != nil {
		log.Fatalln(err)
	}

	buf := new(bytes.Buffer)
	err = tpl.Execute(buf, owners)
	if err != nil {
		log.Fatalln(err)
	}

	return buf.String()
}