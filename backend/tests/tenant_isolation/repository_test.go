package tenant_isolation_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/harmanto-49/cankora/internal/identity/auth"
	"github.com/harmanto-49/cankora/internal/identity/user"
	"github.com/harmanto-49/cankora/internal/project"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// The recording driver runs the real repository/GORM query construction and
// returns no records. It never opens a network connection or touches user data.
type queryRecord struct {
	query string
	args  []driver.NamedValue
}
type recorder struct{ records []queryRecord }
type recordingConnector struct{ r *recorder }

func (c recordingConnector) Connect(context.Context) (driver.Conn, error) {
	return recordingConn{c.r}, nil
}
func (c recordingConnector) Driver() driver.Driver { return recordingDriver{c.r} }

type recordingDriver struct{ r *recorder }

func (d recordingDriver) Open(string) (driver.Conn, error) { return recordingConn{d.r}, nil }

type recordingConn struct{ r *recorder }

func (c recordingConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("unexpected prepare")
}
func (c recordingConn) Close() error              { return nil }
func (c recordingConn) Begin() (driver.Tx, error) { return nil, errors.New("unexpected transaction") }
func (c recordingConn) QueryContext(_ context.Context, q string, args []driver.NamedValue) (driver.Rows, error) {
	c.r.records = append(c.r.records, queryRecord{q, args})
	return emptyRows{}, nil
}

type emptyRows struct{}

func (emptyRows) Columns() []string         { return []string{"id"} }
func (emptyRows) Close() error              { return nil }
func (emptyRows) Next([]driver.Value) error { return io.EOF }

func TestRepositoriesScopeReadsToOrganization(t *testing.T) {
	record := &recorder{}
	sqlDB := sql.OpenDB(recordingConnector{record})
	t.Cleanup(func() { sqlDB.Close() })
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{DisableAutomaticPing: true, SkipDefaultTransaction: true, Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	projects := project.NewRepository(db)
	users := user.NewRepository(db, auth.NewRepository(db))
	for _, org := range []uuid.UUID{uuid.New(), uuid.New()} {
		for _, tc := range []struct {
			name string
			run  func()
		}{
			{"project", func() { projects.FindByID(context.Background(), uuid.New(), org) }},
			{"tasks", func() { projects.FindTaskByID(context.Background(), uuid.New(), uuid.New(), org) }},
			{"users", func() { users.ListUsers(context.Background(), user.UserListFilter{OrganizationID: org}) }},
		} {
			t.Run(tc.name+"/"+org.String(), func(t *testing.T) {
				record.records = nil
				tc.run()
				if len(record.records) == 0 {
					t.Fatal("no SQL executed")
				}
				for _, q := range record.records {
					if !strings.Contains(q.query, "organization_id") || !strings.Contains(q.query, "deleted_at") {
						t.Fatalf("missing tenant/soft-delete filter: %s", q.query)
					}
					found := false
					for _, arg := range q.args {
						if arg.Value == org.String() {
							found = true
						}
					}
					if !found {
						t.Fatalf("organization not bound to query: %s", q.query)
					}
				}
			})
		}
	}
}
