// Package testkit constructs HTTP wiring without opening a database or network connection.
package testkit

import (
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/harmanto-49/cankora/internal/identity/auth"
	"github.com/harmanto-49/cankora/internal/platform/config"
	"github.com/harmanto-49/cankora/internal/platform/http/router"
	"go.uber.org/zap"
)

func Router(configure ...func(*router.Dependencies)) *gin.Engine {
	gin.SetMode(gin.TestMode)
	deps := &router.Dependencies{}
	// Handlers can register bound methods without their backing services. Tests
	// below exercise guards only; entering a business handler must fail the test.
	value := reflect.ValueOf(deps).Elem()
	for i := 0; i < value.NumField(); i++ {
		f := value.Field(i)
		if f.Kind() == reflect.Pointer {
			f.Set(reflect.New(f.Type().Elem()))
		}
	}
	deps.Config = &config.Config{App: config.AppConfig{Env: "test", Version: "test", URL: "http://localhost"}}
	deps.Log = zap.NewNop()
	deps.TokenSvc = auth.NewTokenService(&config.AuthConfig{JWTSecret: "test-only-secret-with-at-least-32-characters", AccessExpMinutes: 15, RefreshExpDays: 1})
	for _, fn := range configure {
		fn(deps)
	}
	return router.New(deps)
}
