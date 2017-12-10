package vsts

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func GetTestServer(t *testing.T, method string, route string, jsonData interface{}, target interface{}) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		w.WriteHeader(http.StatusOK)

		assert.Equal(t, method, req.Method)
		assert.Equal(t, route, req.URL.EscapedPath())

		if method == "POST" {
			err := json.NewDecoder(req.Body).Decode(target)
			if err != nil {
				t.Errorf("Error while reading request JSON: %s", err)
			}

			if jsonData != target {
				t.Errorf("The Post data does not match")
			}
		}

	}))

	return ts
}

// func TestGetInprogressReviews(t *testing.T) {
// 	ts := GetTestServer(t,"GET", "/test", nil, nil)

// 	defer ts.Close()

// 	nsqdUrl := ts.URL
// 	pullRequests := new(VstsPullRequests)
// 	err := GetJsonResponse(nsqdUrl, pullRequests)
// 	fmt.Println("got through the test")

// 	if err != nil {
// 		t.Errorf("Publish() returned an error: %s", err)
// 	}
// }
