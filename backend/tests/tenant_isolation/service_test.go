package tenant_isolation_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/harmanto-49/cankora/internal/governance/workflow"
	"github.com/harmanto-49/cankora/internal/identity/auth"
	"github.com/harmanto-49/cankora/internal/identity/user"
	"github.com/harmanto-49/cankora/internal/project"
	"github.com/harmanto-49/cankora/tests/testkit"
	"go.uber.org/zap"
)

type projectRepo struct {
	project.Repository
	owner uuid.UUID
	task  *project.Task
}

func (r projectRepo) FindByID(_ context.Context, id, org uuid.UUID) (*project.Project, error) {
	if org != r.owner {
		return nil, project.ErrProjectNotFound
	}
	return &project.Project{ID: id, OrganizationID: org}, nil
}
func (r projectRepo) FindTaskByID(_ context.Context, p, id, org uuid.UUID) (*project.Task, error) {
	if org != r.owner || r.task.ProjectID != p || r.task.ID != id {
		return nil, project.ErrTaskNotFound
	}
	return r.task, nil
}
func TestProjectChildAccessChecksTenantAndParent(t *testing.T) {
	owner, other, parent, id := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	repo := projectRepo{owner: owner, task: &project.Task{ID: id, ProjectID: parent, OrganizationID: owner}}
	svc := project.NewService(repo, workflow.New(), zap.NewNop())
	for _, tc := range []struct {
		name        string
		parent, org uuid.UUID
		allowed     bool
	}{
		{"owner", parent, owner, true}, {"other tenant", parent, other, false}, {"wrong parent", uuid.New(), owner, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := svc.GetTask(context.Background(), tc.parent, id, tc.org)
			if tc.allowed {
				if err != nil || got.ID != id {
					t.Fatalf("owner access: %v", err)
				}
			} else if err == nil || got != nil {
				t.Fatal("foreign child was accessible")
			}
		})
	}
}

type userRepo struct {
	user.Repository
	record *auth.User
}

func (r userRepo) FindUserByID(context.Context, uuid.UUID) (*auth.User, error) { return r.record, nil }
func TestUserProfileIsNotReturnedAcrossOrganizations(t *testing.T) {
	owner := uuid.New()
	record := &auth.User{ID: uuid.New(), OrganizationID: owner}
	svc := user.NewService(userRepo{record: record}, testkit.PermissionRepo{}, zap.NewNop())
	if _, err := svc.GetByIDForOrg(context.Background(), record.ID, owner); err != nil {
		t.Fatal(err)
	}
	got, err := svc.GetByIDForOrg(context.Background(), record.ID, uuid.New())
	if !errors.Is(err, auth.ErrUserNotFound) || got != nil {
		t.Fatalf("cross-tenant access: %+v, %v", got, err)
	}
}
