package autoreviewer


import (
	"bytes"
	"log"
	"testing"
	"text/template"
	"github.com/stretchr/testify/assert"
)


func TestParseOwnerFile(t *testing.T) {
	tests := []struct {
		Name string
		TestOwners []string
		ExpectedOwners []string
		ExpectedGroup string
	}{
		{Name: "Success Case",
			TestOwners: []string{PrefixGroup+"Test Group", PrefixNoNotify+"testOwenr", "testNoNotifyOwner", ";this is a comment", ""},
			ExpectedOwners: []string {"testOwenr", "testNoNotifyOwner"},
			ExpectedGroup: "Test Group",
		},
		{Name: "No Group",
			TestOwners: []string{"testOwenr", "testNoNotifyOwner", ";this is a comment", ""},
			ExpectedOwners: []string {"testOwenr", "testNoNotifyOwner"},
			ExpectedGroup: "",
		},
		{Name: "No Owners",
			TestOwners: []string{PrefixGroup+"Test Group",";this is a comment", ""},
			ExpectedOwners: []string {},
			ExpectedGroup: "Test Group",
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
			reviewerGroup := ParseOwnerFile(ownersFile)


			assert.Equal(t,tt.ExpectedGroup, reviewerGroup.Group, "Group should be equal")
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