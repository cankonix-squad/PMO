package integration_test

import (
	"net/http/httptest"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/harmanto-49/cankora/tests/testkit"
)

func TestRouteContract(t *testing.T) {
	router := testkit.Router()
	var routes []string
	for _, r := range router.Routes() {
		routes = append(routes, r.Method+" "+r.Path)
	}
	sort.Strings(routes)
	actual := strings.Join(routes, "\n") + "\n"
	if os.Getenv("UPDATE_ROUTE_CONTRACT") == "1" {
		if err := os.WriteFile("testdata/routes.txt", []byte(actual), 0644); err != nil {
			t.Fatal(err)
		}
	}
	expected, err := os.ReadFile("testdata/routes.txt")
	if err != nil {
		t.Fatal(err)
	}
	if actual != string(expected) {
		t.Fatalf("HTTP routes changed; review against testdata/routes.txt\n%s", actual)
	}
}

func TestProtectedRoutesRejectMissingAndInvalidTokens(t *testing.T) {
	router := testkit.Router()
	public := map[string]bool{"/api/v1/auth/login": true, "/api/v1/auth/refresh": true, "/api/v1/auth/forgot-password": true, "/api/v1/auth/reset-password": true}
	for _, route := range router.Routes() {
		if !strings.HasPrefix(route.Path, "/api/v1/") || public[route.Path] {
			continue
		}
		path := strings.Split(route.Path, "/")
		for i, p := range path {
			if strings.HasPrefix(p, ":") {
				path[i] = "11111111-1111-4111-8111-111111111111"
			}
		}
		for _, token := range []string{"", "Bearer invalid-token"} {
			t.Run(route.Method+" "+route.Path+"/"+token, func(t *testing.T) {
				req := httptest.NewRequest(route.Method, strings.Join(path, "/"), nil)
				req.Header.Set("Authorization", token)
				rec := httptest.NewRecorder()
				router.ServeHTTP(rec, req)
				if rec.Code != 401 {
					t.Fatalf("want 401, got %d: %s", rec.Code, rec.Body.String())
				}
			})
		}
	}
}
