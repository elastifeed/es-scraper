package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gotest.tools/assert"

	"github.com/gorilla/mux"
)

func Test_decodeRequest(t *testing.T) {
	tests := []struct {
		name    string
		args    *http.Request
		want    string
		wantErr bool
	}{
		{name: "test_empty_get",
			args:    httptest.NewRequest("GET", "/", nil), //&http.Request{Method: "GET", Body: nil},
			want:    "",
			wantErr: true},
		{name: "test_empty_post",
			args:    httptest.NewRequest("POST", "/", nil), // &http.Request{Method: "POST", Body: nil},
			want:    "",
			wantErr: true},
		{name: "test_not_json",
			args:    httptest.NewRequest("POST", "/", strings.NewReader("http://golem.de")),
			want:    "",
			wantErr: true},
		{name: "test_additional_json",
			args: httptest.NewRequest("POST", "/", strings.NewReader(
				`{\"somethingelse\" : \"elastifeed\",\"url\" : \"http://golem.de\"}`,
			)),
			want:    "",
			wantErr: true},
		{name: "test_json_no_url",
			args:    httptest.NewRequest("POST", "/", strings.NewReader("{ \"target\" : \"http://golem.de\" }")),
			want:    "",
			wantErr: true},
		{name: "test_json_upper",
			args:    httptest.NewRequest("POST", "/", strings.NewReader("{ \"Url\" : \"http://golem.de\" }")),
			want:    "http://golem.de",
			wantErr: false},
		{name: "test_json_correct",
			args:    httptest.NewRequest("POST", "/", strings.NewReader("{ \"url\" : \"http://golem.de\" }")),
			want:    "http://golem.de",
			wantErr: false},
	}
	// Test runs
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := decodeRequest(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("decodeRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("decodeRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInitRouter(t *testing.T) {
	tests := []struct {
		name string
		have *mux.Router
	}{
		{name: "test_router_init",
			have: InitRouter()},
	}

	validRoute := func(route string) bool {
		valid := []string{"", "/scrape/", "/scrape/thumbnail", "/scrape/pdf", "/scrape/content"}
		for _, elem := range valid {
			if route == elem {
				return true
			}
		}
		return false
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Walk all the routes and (sub)routers and assert that they are as expected
			tt.have.Walk(
				func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
					methods, _ := route.GetMethods()
					assert.DeepEqual(t, methods, []string{"POST"})                                   // Only have POST routes
					assert.Assert(t, validRoute(route.GetName()), "Router had an unexpected route!") // No unexpected routes or endpoints
					return nil
				})
		})
	}

}
