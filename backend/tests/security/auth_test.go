package security_test

import (
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/harmanto-49/cankora/internal/identity/auth"
	"github.com/harmanto-49/cankora/internal/platform/config"
	"github.com/harmanto-49/cankora/tests/testkit"
)

func TestRevokedSessionCannotReachProtectedHandler(t *testing.T) {
	router, token, err := testkit.AuthenticatedRouter(nil, true)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest("GET", "/api/v1/users", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != 401 {
		t.Fatalf("want 401, got %d", rec.Code)
	}
}
func TestAccessAndRefreshTokensAreNotInterchangeable(t *testing.T) {
	svc := auth.NewTokenService(&config.AuthConfig{JWTSecret: "test-secret-at-least-32-characters", AccessExpMinutes: 15, RefreshExpDays: 1})
	pair, err := svc.GenerateTokenPair(&auth.User{ID: uuid.New(), OrganizationID: uuid.New()}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = svc.ValidateAccessToken(pair.AccessToken); err != nil {
		t.Fatal(err)
	}
	if _, err = svc.ValidateRefreshToken(pair.RefreshToken); err != nil {
		t.Fatal(err)
	}
	if _, err = svc.ValidateAccessToken(pair.RefreshToken); err == nil {
		t.Fatal("refresh token accepted as access token")
	}
	if _, err = svc.ValidateRefreshToken(pair.AccessToken); err == nil {
		t.Fatal("access token accepted as refresh token")
	}
}
