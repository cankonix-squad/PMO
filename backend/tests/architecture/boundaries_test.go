package architecture_test

import (
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestMigratedPackagesDoNotDependOnLegacyBoundaries(t *testing.T) {
	err := filepath.WalkDir("../../internal", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		f, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, im := range f.Imports {
			name, _ := strconv.Unquote(im.Path.Value)
			for _, old := range []string{"/internal/core/", "/internal/platform/middleware", "/internal/platform/server", "/internal/modules/project", "/internal/modules/analytics", "/internal/modules/integration"} {
				if strings.Contains(name, old) {
					t.Errorf("%s imports retired boundary %s", path, name)
				}
			}
			// The services extracted in this refactor must stay independent of HTTP
			// frameworks and database clients. Adapters belong in handler/repository files.
			pureService := (filepath.Dir(path) == "../../internal/project" && strings.HasSuffix(path, "service.go")) || path == "../../internal/identity/user/service.go"
			if pureService && (name == "gorm.io/gorm" || name == "github.com/gin-gonic/gin" || strings.Contains(name, "/pgx/")) {
				t.Errorf("%s imports infrastructure %s", path, name)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
