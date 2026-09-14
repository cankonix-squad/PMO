package authorization_test

import (
	"net/http/httptest"
	"testing"

	"github.com/harmanto-49/cankora/tests/testkit"
)

func TestAdminAndProjectRoutesDenyUnprivilegedUser(t *testing.T) {
	router, token, err := testkit.AuthenticatedRouter(nil, false)
	if err != nil {
		t.Fatal(err)
	}
	id := "11111111-1111-4111-8111-111111111111"
	routes := [][2]string{
		{"GET", "/users"}, {"POST", "/users"}, {"GET", "/users/" + id}, {"PUT", "/users/" + id}, {"POST", "/users/" + id + "/deactivate"},
		{"GET", "/governance/submissions"}, {"POST", "/governance/submissions"}, {"POST", "/governance/submissions/" + id + "/approve"},
		{"GET", "/audit-logs"}, {"GET", "/notifications"},
		{"GET", "/projects"}, {"POST", "/projects"}, {"PUT", "/projects/" + id}, {"DELETE", "/projects/" + id},
		{"GET", "/projects/" + id + "/tasks"}, {"POST", "/projects/" + id + "/tasks"}, {"GET", "/projects/" + id + "/budgets"},
		{"GET", "/projects/" + id + "/documents"}, {"GET", "/projects/" + id + "/periodic-reports"},
	}
	for _, r := range routes {
		t.Run(r[0]+r[1], func(t *testing.T) {
			req := httptest.NewRequest(r[0], "/api/v1"+r[1], nil)
			req.Header.Set("Authorization", "Bearer "+token)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			if rec.Code != 403 {
				t.Fatalf("want 403, got %d: %s", rec.Code, rec.Body.String())
			}
		})
	}
}
