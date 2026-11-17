package router

import (
	"github.com/harmanto-49/cankora/internal/analytics/benefit"
	"github.com/harmanto-49/cankora/internal/analytics/commandcenter"
	"github.com/harmanto-49/cankora/internal/analytics/dashboard"
	analyticsexecutive "github.com/harmanto-49/cankora/internal/analytics/executive"
	analyticsgis "github.com/harmanto-49/cankora/internal/analytics/gis"
	"github.com/harmanto-49/cankora/internal/analytics/priority"
	analyticsprogram "github.com/harmanto-49/cankora/internal/analytics/program"
	"github.com/harmanto-49/cankora/internal/analytics/projectcontrol"
	"github.com/harmanto-49/cankora/internal/analytics/report"
	analyticsreporting "github.com/harmanto-49/cankora/internal/analytics/reporting"
	"github.com/harmanto-49/cankora/internal/governance/audit"
	"github.com/harmanto-49/cankora/internal/governance/auditlog"
	"github.com/harmanto-49/cankora/internal/governance/dataquality"
	governance "github.com/harmanto-49/cankora/internal/governance/submissions"
	"github.com/harmanto-49/cankora/internal/governance/workflow"
	"github.com/harmanto-49/cankora/internal/identity/auth"
	"github.com/harmanto-49/cankora/internal/identity/organization"
	"github.com/harmanto-49/cankora/internal/identity/rbac"
	"github.com/harmanto-49/cankora/internal/identity/user"
	"github.com/harmanto-49/cankora/internal/integrations/bim"
	"github.com/harmanto-49/cankora/internal/integrations/government"
	"github.com/harmanto-49/cankora/internal/integrations/imports"
	"github.com/harmanto-49/cankora/internal/integrations/primavera"
	"github.com/harmanto-49/cankora/internal/kpbu"
	"github.com/harmanto-49/cankora/internal/notifications"
	notification "github.com/harmanto-49/cankora/internal/notifications/delivery"
	"github.com/harmanto-49/cankora/internal/platform/config"
	"github.com/harmanto-49/cankora/internal/project"
	projectcategory "github.com/harmanto-49/cankora/internal/project/category"
	"github.com/harmanto-49/cankora/internal/project/field"
	"github.com/harmanto-49/cankora/internal/project/health"
	"github.com/harmanto-49/cankora/internal/project/monitoring"
	"github.com/harmanto-49/cankora/internal/project/portfolio"
	"github.com/harmanto-49/cankora/internal/project/spatial"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Dependencies holds all wired-up services and repositories.
type Dependencies struct {
	DB           *gorm.DB
	Config       *config.Config
	Log          *zap.Logger
	FSM          *workflow.FSM
	Notification notification.Provider

	// Repositories
	AuthRepo  auth.Repository
	AuditRepo audit.Repository
	RBACRepo  rbac.Repository
	UserRepo  user.Repository

	// Services
	TokenSvc   *auth.TokenService
	AuthSvc    *auth.Service
	UserSvc    *user.Service
	ProjectSvc *project.Service

	// Writers
	AuditWriter *audit.Writer

	// Handlers
	AuthHandler               *auth.Handler
	ProjectHandler            *project.Handler
	UserHandler               *user.Handler
	DashboardHandler          *dashboard.Handler
	ReportHandler             *report.Handler
	OrgHandler                *organization.Handler
	OrgCRUDHandler            *organization.OrgCRUDHandler
	PortfolioHandler          *portfolio.Handler
	SpatialHandler            *spatial.Handler
	MonitoringHandler         *monitoring.Handler
	DataQualityHandler        *dataquality.Handler
	FieldHandler              *field.Handler
	HealthHandler             *health.Handler
	CommandCenterHandler      *commandcenter.Handler
	CommandDecisionHandler    *commandcenter.CommandHandler
	BenefitHandler            *benefit.Handler
	ProjectControlHandler     *projectcontrol.Handler
	PriorityHandler           *priority.Handler
	ProgramDashboardHandler   *analyticsprogram.Handler
	ExecutiveDashboardHandler *analyticsexecutive.Handler
	GISDashboardHandler       *analyticsgis.Handler
	ReportingAnalyticsHandler *analyticsreporting.Handler
	ImportHandler             *imports.Handler
	PrimaveraHandler          *primavera.Handler
	GovernmentHandler         *government.Handler
	BIMHandler                *bim.Handler
	GovernanceHandler         *governance.Handler
	AuditLogHandler           *auditlog.Handler
	NotificationHandler       *notifications.Handler
	ProjectCategoryHandler    *projectcategory.Handler
	KPBUHandler               *kpbu.Handler
}
