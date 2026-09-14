package authorization_test

import (
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/harmanto-49/cankora/internal/identity/auth"
	"github.com/harmanto-49/cankora/internal/identity/rbac"
	"github.com/harmanto-49/cankora/internal/platform/http/middleware"
	"github.com/harmanto-49/cankora/tests/testkit"
)

func TestPermissionChecksResourceAndAction(t *testing.T) {
	for _, tc := range []struct {
		name  string
		perms []rbac.Permission
		err   error
		want  int
	}{
		{"matching", []rbac.Permission{{Resource: "project", Action: "update"}}, nil, 204},
		{"read only", []rbac.Permission{{Resource: "project", Action: "view"}}, nil, 403},
		{"wrong resource", []rbac.Permission{{Resource: "user", Action: "update"}}, nil, 403},
		{"database failure", nil, errors.New("unavailable"), 500},
	} {
		t.Run(tc.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			r := gin.New()
			r.Use(func(c *gin.Context) { c.Set(string(auth.ContextKeyClaims), &auth.Claims{}); c.Next() })
			r.PUT("/project", middleware.RequirePermission(testkit.PermissionRepo{Permissions: tc.perms, Err: tc.err}, "project", "update"), func(c *gin.Context) { c.Status(204) })
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, httptest.NewRequest("PUT", "/project", nil))
			if rec.Code != tc.want {
				t.Fatalf("want %d, got %d", tc.want, rec.Code)
			}
		})
	}
}
