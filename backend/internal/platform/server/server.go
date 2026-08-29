package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/harmanto-49/cankora/internal/core/audit"
	"github.com/harmanto-49/cankora/internal/core/auth"
	"github.com/harmanto-49/cankora/internal/core/notification"
	"github.com/harmanto-49/cankora/internal/core/organization"
	"github.com/harmanto-49/cankora/internal/core/rbac"
	"github.com/harmanto-49/cankora/internal/core/user"
	"github.com/harmanto-49/cankora/internal/core/workflow"
	analyticsexecutive "github.com/harmanto-49/cankora/internal/modules/analytics/executive"
	analyticsgis "github.com/harmanto-49/cankora/internal/modules/analytics/gis"
	analyticsprogram "github.com/harmanto-49/cankora/internal/modules/analytics/program"
	"github.com/harmanto-49/cankora/internal/modules/analytics/projectcontrol"
	analyticsreporting "github.com/harmanto-49/cankora/internal/modules/analytics/reporting"
	"github.com/harmanto-49/cankora/internal/modules/auditlog"
	"github.com/harmanto-49/cankora/internal/modules/benefit"
	"github.com/harmanto-49/cankora/internal/modules/commandcenter"
	"github.com/harmanto-49/cankora/internal/modules/dataquality"
	"github.com/harmanto-49/cankora/internal/modules/field"
	"github.com/harmanto-49/cankora/internal/modules/governance"
	"github.com/harmanto-49/cankora/internal/modules/health"
	"github.com/harmanto-49/cankora/internal/modules/imports"
	"github.com/harmanto-49/cankora/internal/modules/integration/bim"
	"github.com/harmanto-49/cankora/internal/modules/integration/government"
	"github.com/harmanto-49/cankora/internal/modules/integration/primavera"
	"github.com/harmanto-49/cankora/internal/modules/monitoring"
	"github.com/harmanto-49/cankora/internal/modules/notifications"
	"github.com/harmanto-49/cankora/internal/modules/portfolio"
	"github.com/harmanto-49/cankora/internal/modules/priority"
	"github.com/harmanto-49/cankora/internal/modules/project"
	"github.com/harmanto-49/cankora/internal/modules/report"
	"github.com/harmanto-49/cankora/internal/modules/spatial"
	"github.com/harmanto-49/cankora/internal/platform/config"
	"github.com/harmanto-49/cankora/internal/platform/dashboard"
	"github.com/harmanto-49/cankora/internal/platform/database"
	"github.com/harmanto-49/cankora/internal/platform/middleware"
	"github.com/harmanto-49/cankora/internal/shared/constants"
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
}

// Wire initialises all dependencies and returns a populated Dependencies struct.
func Wire(cfg *config.Config, log *zap.Logger) (*Dependencies, error) {
	db, err := database.Connect(cfg.Database.DSN, cfg.IsDevelopment(), log)
	if err != nil {
		return nil, fmt.Errorf("server: database: %w", err)
	}

	fsm := workflow.New()

	notifProvider := notification.NewSMTPProvider(notification.SMTPConfig{
		Host:     cfg.SMTP.Host,
		Port:     cfg.SMTP.Port,
		User:     cfg.SMTP.User,
		Password: cfg.SMTP.Password,
		From:     cfg.SMTP.From,
	})

	// Repositories
	authRepo := auth.NewRepository(db)
	auditRepo := audit.NewRepository(db)
	rbacRepo := rbac.NewRepository(db)
	userRepo := user.NewRepository(db, authRepo)

	// Services
	tokenSvc := auth.NewTokenService(&cfg.Auth)
	authSvc := auth.NewService(authRepo, rbacRepo, tokenSvc, cfg, log)
	userSvc := user.NewService(userRepo, rbacRepo, log)
	auditWriter := audit.NewWriter(auditRepo, log)
	projectRepo := project.NewRepository(db)
	projectSvc := project.NewService(projectRepo, fsm, log)
	projectSvc.ConfigureDocumentStorage(cfg.Storage.LocalPath, cfg.Storage.MaxSizeBytes)

	// Handlers
	authHandler := auth.NewHandler(authSvc, tokenSvc)
	projectHandler := project.NewHandler(projectSvc, auditWriter).WithDB(db)
	userHandler := user.NewHandler(userSvc)
	dashboardHandler := dashboard.NewHandler(db, log)
	reportHandler := report.NewHandler(db, log)
	orgHandler := organization.NewHandler(db, log)
	portfolioHandler := portfolio.NewHandler(db, log)
	spatialHandler := spatial.NewHandler(db, log)
	monitoringHandler := monitoring.NewHandler(db, log)
	dataQualityHandler := dataquality.NewHandler(dataquality.NewService(db, auditWriter), log)
	fieldHandler := field.NewHandler(field.NewService(db, auditWriter, cfg.Storage.LocalPath, cfg.Storage.MaxSizeBytes), log)
	healthHandler := health.NewHandler(health.NewService(db, auditWriter))
	commandCenterHandler := commandcenter.NewHandler(commandcenter.NewService(db))
	commandDecisionHandler := commandcenter.NewCommandHandler(db, auditWriter)
	benefitHandler := benefit.NewHandler(benefit.NewService(db, auditWriter))
	projectControlHandler := projectcontrol.NewHandler(projectcontrol.NewService(db))
	priorityHandler := priority.NewHandler(priority.NewService(db, auditWriter))
	programDashboardHandler := analyticsprogram.NewHandler(analyticsprogram.NewService(db))
	reportingAnalyticsHandler := analyticsreporting.NewServiceAndHandler(db)
	executiveDashboardHandler := analyticsexecutive.NewHandler(analyticsexecutive.NewService(db))
	gisDashboardHandler := analyticsgis.NewHandler(analyticsgis.NewService(db))
	importHandler := imports.NewHandler(imports.NewService(db, auditWriter, log))
	primaveraHandler := primavera.NewHandler(primavera.NewService(db, auditWriter, log))
	governmentHandler := government.NewHandler(government.NewService(db, auditWriter, log))
	bimHandler := bim.NewHandler(bim.NewService(db, auditWriter))
	governanceHandler := governance.NewHandler(governance.NewService(db, auditWriter, log), log)
	auditLogHandler := auditlog.NewHandler(auditRepo)
	notifRepo := notification.NewRepository(db)
	notifSvc := notification.NewService(notifRepo, notifProvider, log, cfg.SMTP.From)
	notificationHandler := notifications.NewHandler(notifSvc)

	_ = auditWriter

	return &Dependencies{
		DB:                        db,
		Config:                    cfg,
		Log:                       log,
		FSM:                       fsm,
		Notification:              notifProvider,
		AuthRepo:                  authRepo,
		AuditRepo:                 auditRepo,
		RBACRepo:                  rbacRepo,
		UserRepo:                  userRepo,
		TokenSvc:                  tokenSvc,
		AuthSvc:                   authSvc,
		UserSvc:                   userSvc,
		ProjectSvc:                projectSvc,
		AuditWriter:               auditWriter,
		AuthHandler:               authHandler,
		ProjectHandler:            projectHandler,
		OrgHandler:                orgHandler,
		UserHandler:               userHandler,
		DashboardHandler:          dashboardHandler,
		ReportHandler:             reportHandler,
		PortfolioHandler:          portfolioHandler,
		SpatialHandler:            spatialHandler,
		MonitoringHandler:         monitoringHandler,
		DataQualityHandler:        dataQualityHandler,
		FieldHandler:              fieldHandler,
		HealthHandler:             healthHandler,
		CommandCenterHandler:      commandCenterHandler,
		CommandDecisionHandler:    commandDecisionHandler,
		BenefitHandler:            benefitHandler,
		ProjectControlHandler:     projectControlHandler,
		PriorityHandler:           priorityHandler,
		ProgramDashboardHandler:   programDashboardHandler,
		ExecutiveDashboardHandler: executiveDashboardHandler,
		GISDashboardHandler:       gisDashboardHandler,
		ReportingAnalyticsHandler: reportingAnalyticsHandler,
		ImportHandler:             importHandler,
		PrimaveraHandler:          primaveraHandler,
		GovernmentHandler:         governmentHandler,
		BIMHandler:                bimHandler,
		GovernanceHandler:         governanceHandler,
		AuditLogHandler:           auditLogHandler,
		NotificationHandler:       notificationHandler,
	}, nil
}

// New creates a configured *gin.Engine with all routes registered.
func New(deps *Dependencies) *gin.Engine {
	if deps.Config.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.UseH2C = false

	// Global middleware
	allowedOrigins := strings.Split(
		getEnvOr("CORS_ALLOWED_ORIGINS", deps.Config.App.URL),
		",",
	)
	r.Use(middleware.RequestID())
	r.Use(middleware.CORS(allowedOrigins))
	r.Use(middleware.Logger(deps.Log))
	r.Use(gin.Recovery())

	// Health check — no auth required
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"version": deps.Config.App.Version,
			"env":     deps.Config.App.Env,
		})
	})

	// API v1
	v1 := r.Group("/api/v1")

	// Auth routes (public)
	authPublic := v1.Group("/auth")
	deps.AuthHandler.RegisterRoutes(authPublic)

	// Auth routes (protected)
	authProtected := v1.Group("/auth")
	authProtected.Use(middleware.AuthRequired(deps.TokenSvc, deps.AuthSvc))
	deps.AuthHandler.RegisterProtectedRoutes(authProtected)

	// Projects (protected)
	projectsGroup := v1.Group("/projects")
	projectsGroup.Use(middleware.AuthRequired(deps.TokenSvc, deps.AuthSvc))
	projectsGroup.GET("", middleware.RequirePermission(deps.RBACRepo, constants.ResourceProjects, constants.ActionView), deps.ProjectHandler.List)
	projectsGroup.POST("", middleware.RequirePermission(deps.RBACRepo, constants.ResourceProjects, constants.ActionCreate), deps.ProjectHandler.Create)
	projectsGroup.GET("/:id", middleware.RequirePermission(deps.RBACRepo, constants.ResourceProjects, constants.ActionView), deps.ProjectHandler.GetByID)
	projectsGroup.PUT("/:id", middleware.RequirePermission(deps.RBACRepo, constants.ResourceProjects, constants.ActionUpdate), deps.ProjectHandler.Update)
	projectsGroup.DELETE("/:id", middleware.RequirePermission(deps.RBACRepo, constants.ResourceProjects, constants.ActionDelete), deps.ProjectHandler.Delete)
	projectsGroup.POST("/:id/transition", middleware.RequirePermission(deps.RBACRepo, constants.ResourceProjects, constants.ActionUpdate), deps.ProjectHandler.Transition)
	projectsGroup.GET("/:id/progress-history", middleware.RequirePermission(deps.RBACRepo, constants.ResourceProjects, constants.ActionView), deps.ProjectHandler.GetProgressHistory)

	// Tasks — nested under projects
	tasksGroup := projectsGroup.Group("/:id/tasks")
	tasksGroup.GET("", middleware.RequirePermission(deps.RBACRepo, constants.ResourceTasks, constants.ActionView), deps.ProjectHandler.ListTasks)
	tasksGroup.POST("", middleware.RequirePermission(deps.RBACRepo, constants.ResourceTasks, constants.ActionCreate), deps.ProjectHandler.CreateTask)
	tasksGroup.GET("/:taskID", middleware.RequirePermission(deps.RBACRepo, constants.ResourceTasks, constants.ActionView), deps.ProjectHandler.GetTask)
	tasksGroup.PUT("/:taskID", middleware.RequirePermission(deps.RBACRepo, constants.ResourceTasks, constants.ActionUpdate), deps.ProjectHandler.UpdateTask)
	tasksGroup.DELETE("/:taskID", middleware.RequirePermission(deps.RBACRepo, constants.ResourceTasks, constants.ActionDelete), deps.ProjectHandler.DeleteTask)

	// Milestones — nested under projects
	milestonesGroup := projectsGroup.Group("/:id/milestones")
	milestonesGroup.GET("", middleware.RequirePermission(deps.RBACRepo, constants.ResourceMilestones, constants.ActionView), deps.ProjectHandler.ListMilestones)
	milestonesGroup.POST("", middleware.RequirePermission(deps.RBACRepo, constants.ResourceMilestones, constants.ActionCreate), deps.ProjectHandler.CreateMilestone)
	milestonesGroup.GET("/:milestoneID", middleware.RequirePermission(deps.RBACRepo, constants.ResourceMilestones, constants.ActionView), deps.ProjectHandler.GetMilestone)
	milestonesGroup.PUT("/:milestoneID", middleware.RequirePermission(deps.RBACRepo, constants.ResourceMilestones, constants.ActionUpdate), deps.ProjectHandler.UpdateMilestone)
	milestonesGroup.DELETE("/:milestoneID", middleware.RequirePermission(deps.RBACRepo, constants.ResourceMilestones, constants.ActionDelete), deps.ProjectHandler.DeleteMilestone)

	// Issues — nested under projects
	issuesGroup := projectsGroup.Group("/:id/issues")
	issuesGroup.GET("", middleware.RequirePermission(deps.RBACRepo, constants.ResourceIssues, constants.ActionView), deps.ProjectHandler.ListIssues)
	issuesGroup.POST("", middleware.RequirePermission(deps.RBACRepo, constants.ResourceIssues, constants.ActionCreate), deps.ProjectHandler.CreateIssue)
	issuesGroup.GET("/:issueID", middleware.RequirePermission(deps.RBACRepo, constants.ResourceIssues, constants.ActionView), deps.ProjectHandler.GetIssue)
	issuesGroup.PUT("/:issueID", middleware.RequirePermission(deps.RBACRepo, constants.ResourceIssues, constants.ActionUpdate), deps.ProjectHandler.UpdateIssue)
	issuesGroup.DELETE("/:issueID", middleware.RequirePermission(deps.RBACRepo, constants.ResourceIssues, constants.ActionDelete), deps.ProjectHandler.DeleteIssue)

	// Risks — nested under projects
	risksGroup := projectsGroup.Group("/:id/risks")
	risksGroup.GET("", middleware.RequirePermission(deps.RBACRepo, constants.ResourceRisks, constants.ActionView), deps.ProjectHandler.ListRisks)
	risksGroup.POST("", middleware.RequirePermission(deps.RBACRepo, constants.ResourceRisks, constants.ActionCreate), deps.ProjectHandler.CreateRisk)
	risksGroup.GET("/:riskID", middleware.RequirePermission(deps.RBACRepo, constants.ResourceRisks, constants.ActionView), deps.ProjectHandler.GetRisk)
	risksGroup.PUT("/:riskID", middleware.RequirePermission(deps.RBACRepo, constants.ResourceRisks, constants.ActionUpdate), deps.ProjectHandler.UpdateRisk)
	risksGroup.DELETE("/:riskID", middleware.RequirePermission(deps.RBACRepo, constants.ResourceRisks, constants.ActionDelete), deps.ProjectHandler.DeleteRisk)

	// Budgets — nested under projects
	budgetsGroup := projectsGroup.Group("/:id/budgets")
	budgetsGroup.GET("", middleware.RequirePermission(deps.RBACRepo, constants.ResourceBudgets, constants.ActionView), deps.ProjectHandler.ListBudgets)
	budgetsGroup.POST("", middleware.RequirePermission(deps.RBACRepo, constants.ResourceBudgets, constants.ActionCreate), deps.ProjectHandler.CreateBudget)
	budgetsGroup.GET("/:budgetID", middleware.RequirePermission(deps.RBACRepo, constants.ResourceBudgets, constants.ActionView), deps.ProjectHandler.GetBudget)
	budgetsGroup.PUT("/:budgetID", middleware.RequirePermission(deps.RBACRepo, constants.ResourceBudgets, constants.ActionUpdate), deps.ProjectHandler.UpdateBudget)
	budgetsGroup.DELETE("/:budgetID", middleware.RequirePermission(deps.RBACRepo, constants.ResourceBudgets, constants.ActionDelete), deps.ProjectHandler.DeleteBudget)

	// Contracts — nested under projects
	contractsGroup := projectsGroup.Group("/:id/contracts")
	contractsGroup.GET("", middleware.RequirePermission(deps.RBACRepo, constants.ResourceContracts, constants.ActionView), deps.ProjectHandler.ListContracts)
	contractsGroup.POST("", middleware.RequirePermission(deps.RBACRepo, constants.ResourceContracts, constants.ActionCreate), deps.ProjectHandler.CreateContract)
	contractsGroup.GET("/:contractID", middleware.RequirePermission(deps.RBACRepo, constants.ResourceContracts, constants.ActionView), deps.ProjectHandler.GetContract)
	contractsGroup.PUT("/:contractID", middleware.RequirePermission(deps.RBACRepo, constants.ResourceContracts, constants.ActionUpdate), deps.ProjectHandler.UpdateContract)
	contractsGroup.DELETE("/:contractID", middleware.RequirePermission(deps.RBACRepo, constants.ResourceContracts, constants.ActionDelete), deps.ProjectHandler.DeleteContract)

	// Documents — nested under projects (P1-005)
	documentsGroup := projectsGroup.Group("/:id/documents")
	documentsGroup.GET("", middleware.RequirePermission(deps.RBACRepo, constants.ResourceDocuments, constants.ActionView), deps.ProjectHandler.ListDocuments)
	documentsGroup.POST("", middleware.RequirePermission(deps.RBACRepo, constants.ResourceDocuments, constants.ActionCreate), deps.ProjectHandler.UploadDocument)
	documentsGroup.GET("/:documentID", middleware.RequirePermission(deps.RBACRepo, constants.ResourceDocuments, constants.ActionView), deps.ProjectHandler.GetDocument)
	documentsGroup.GET("/:documentID/download", middleware.RequirePermission(deps.RBACRepo, constants.ResourceDocuments, constants.ActionView), deps.ProjectHandler.DownloadDocument)
	documentsGroup.PUT("/:documentID", middleware.RequirePermission(deps.RBACRepo, constants.ResourceDocuments, constants.ActionUpdate), deps.ProjectHandler.UpdateDocument)
	documentsGroup.DELETE("/:documentID", middleware.RequirePermission(deps.RBACRepo, constants.ResourceDocuments, constants.ActionDelete), deps.ProjectHandler.DeleteDocument)

	// Periodic Reports — nested under projects (CANKORA-DASH-002)
	periodicGroup := projectsGroup.Group("/:id/periodic-reports")
	periodicGroup.GET("", middleware.RequirePermission(deps.RBACRepo, constants.ResourcePeriodicReport, constants.ActionView), deps.ProjectHandler.ListPeriodicReports)
	periodicGroup.POST("", middleware.RequirePermission(deps.RBACRepo, constants.ResourcePeriodicReport, constants.ActionCreate), deps.ProjectHandler.CreatePeriodicReport)
	periodicGroup.GET("/:reportID", middleware.RequirePermission(deps.RBACRepo, constants.ResourcePeriodicReport, constants.ActionView), deps.ProjectHandler.GetPeriodicReport)
	periodicGroup.PUT("/:reportID", middleware.RequirePermission(deps.RBACRepo, constants.ResourcePeriodicReport, constants.ActionUpdate), deps.ProjectHandler.UpdatePeriodicReport)
	periodicGroup.DELETE("/:reportID", middleware.RequirePermission(deps.RBACRepo, constants.ResourcePeriodicReport, constants.ActionDelete), deps.ProjectHandler.DeletePeriodicReport)

	// Corrective Actions — nested under projects (P1-006)
	caGroup := projectsGroup.Group("/:id/corrective-actions")
	caGroup.GET("", middleware.RequirePermission(deps.RBACRepo, constants.ResourceCorrectiveActions, constants.ActionView), deps.ProjectHandler.ListCorrectiveActions)
	caGroup.POST("", middleware.RequirePermission(deps.RBACRepo, constants.ResourceCorrectiveActions, constants.ActionCreate), deps.ProjectHandler.CreateCorrectiveAction)
	caGroup.GET("/:caID", middleware.RequirePermission(deps.RBACRepo, constants.ResourceCorrectiveActions, constants.ActionView), deps.ProjectHandler.GetCorrectiveAction)
	caGroup.PUT("/:caID", middleware.RequirePermission(deps.RBACRepo, constants.ResourceCorrectiveActions, constants.ActionUpdate), deps.ProjectHandler.UpdateCorrectiveAction)
	caGroup.DELETE("/:caID", middleware.RequirePermission(deps.RBACRepo, constants.ResourceCorrectiveActions, constants.ActionDelete), deps.ProjectHandler.DeleteCorrectiveAction)
	caGroup.POST("/:caID/transition", middleware.RequirePermission(deps.RBACRepo, constants.ResourceCorrectiveActions, constants.ActionUpdate), deps.ProjectHandler.TransitionCorrectiveAction)

	// Vendors (protected, top-level master)
	vendorsGroup := v1.Group("/vendors")
	vendorsGroup.Use(middleware.AuthRequired(deps.TokenSvc, deps.AuthSvc))
	vendorsGroup.GET("", middleware.RequirePermission(deps.RBACRepo, constants.ResourceVendors, constants.ActionView), deps.ProjectHandler.ListVendors)
	vendorsGroup.POST("", middleware.RequirePermission(deps.RBACRepo, constants.ResourceVendors, constants.ActionCreate), deps.ProjectHandler.CreateVendor)
	vendorsGroup.GET("/:vendorID", middleware.RequirePermission(deps.RBACRepo, constants.ResourceVendors, constants.ActionView), deps.ProjectHandler.GetVendor)
	vendorsGroup.PUT("/:vendorID", middleware.RequirePermission(deps.RBACRepo, constants.ResourceVendors, constants.ActionUpdate), deps.ProjectHandler.UpdateVendor)
	vendorsGroup.DELETE("/:vendorID", middleware.RequirePermission(deps.RBACRepo, constants.ResourceVendors, constants.ActionDelete), deps.ProjectHandler.DeleteVendor)

	// Users (protected)
	usersGroup := v1.Group("/users")
	usersGroup.Use(middleware.AuthRequired(deps.TokenSvc, deps.AuthSvc))
	usersGroup.GET("", middleware.RequirePermission(deps.RBACRepo, constants.ResourceUsers, constants.ActionView), deps.UserHandler.List)
	usersGroup.POST("", middleware.RequirePermission(deps.RBACRepo, constants.ResourceUsers, constants.ActionCreate), deps.UserHandler.Create)
	usersGroup.GET("/:id", middleware.RequirePermission(deps.RBACRepo, constants.ResourceUsers, constants.ActionView), deps.UserHandler.GetByID)
	usersGroup.PUT("/:id", middleware.RequirePermission(deps.RBACRepo, constants.ResourceUsers, constants.ActionUpdate), deps.UserHandler.Update)
	usersGroup.POST("/:id/deactivate", middleware.RequirePermission(deps.RBACRepo, constants.ResourceUsers, constants.ActionUpdate), deps.UserHandler.Deactivate)

	// Dashboard (protected)
	v1.GET("/dashboard",
		middleware.AuthRequired(deps.TokenSvc, deps.AuthSvc),
		middleware.RequirePermission(deps.RBACRepo, constants.ResourceReports, constants.ActionView),
		deps.DashboardHandler.Get,
	)
	v1.GET("/dashboard/trend",
		middleware.AuthRequired(deps.TokenSvc, deps.AuthSvc),
		middleware.RequirePermission(deps.RBACRepo, constants.ResourceReports, constants.ActionView),
		deps.DashboardHandler.GetTrend,
	)

	// Reports (protected)
	reportsGroup := v1.Group("/reports")
	reportsGroup.Use(
		middleware.AuthRequired(deps.TokenSvc, deps.AuthSvc),
		middleware.RequirePermission(deps.RBACRepo, constants.ResourceReports, constants.ActionView),
	)
	deps.ReportHandler.RegisterRoutes(reportsGroup)

	// Org Units (protected) — P1-008
	orgUnitsGroup := v1.Group("/org-units")
	orgUnitsGroup.Use(middleware.AuthRequired(deps.TokenSvc, deps.AuthSvc))
	orgUnitsGroup.GET("", middleware.RequirePermission(deps.RBACRepo, constants.ResourceOrgUnit, constants.ActionView), deps.OrgHandler.List)
	orgUnitsGroup.POST("", middleware.RequirePermission(deps.RBACRepo, constants.ResourceOrgUnit, constants.ActionCreate), deps.OrgHandler.Create)
	orgUnitsGroup.GET("/:unitID", middleware.RequirePermission(deps.RBACRepo, constants.ResourceOrgUnit, constants.ActionView), deps.OrgHandler.Get)
	orgUnitsGroup.PUT("/:unitID", middleware.RequirePermission(deps.RBACRepo, constants.ResourceOrgUnit, constants.ActionUpdate), deps.OrgHandler.Update)
	orgUnitsGroup.DELETE("/:unitID", middleware.RequirePermission(deps.RBACRepo, constants.ResourceOrgUnit, constants.ActionDelete), deps.OrgHandler.Delete)

	// Programs (protected) — P1-010
	programsGroup := v1.Group("/programs")
	programsGroup.Use(middleware.AuthRequired(deps.TokenSvc, deps.AuthSvc))
	programsGroup.GET("", middleware.RequirePermission(deps.RBACRepo, constants.ResourceProgram, constants.ActionView), deps.PortfolioHandler.List)
	programsGroup.POST("", middleware.RequirePermission(deps.RBACRepo, constants.ResourceProgram, constants.ActionCreate), deps.PortfolioHandler.Create)
	programsGroup.GET("/:programID", middleware.RequirePermission(deps.RBACRepo, constants.ResourceProgram, constants.ActionView), deps.PortfolioHandler.Get)
	programsGroup.PUT("/:programID", middleware.RequirePermission(deps.RBACRepo, constants.ResourceProgram, constants.ActionUpdate), deps.PortfolioHandler.Update)
	programsGroup.DELETE("/:programID", middleware.RequirePermission(deps.RBACRepo, constants.ResourceProgram, constants.ActionDelete), deps.PortfolioHandler.Delete)

	// Sectors (protected) — P1-010; RequirePermission added UAT-006
	sectorsGroup := v1.Group("/sectors")
	sectorsGroup.Use(middleware.AuthRequired(deps.TokenSvc, deps.AuthSvc))
	sectorsGroup.GET("", middleware.RequirePermission(deps.RBACRepo, constants.ResourceSector, constants.ActionView), deps.SpatialHandler.ListSectors)
	sectorsGroup.POST("", middleware.RequirePermission(deps.RBACRepo, constants.ResourceSector, constants.ActionCreate), deps.SpatialHandler.CreateSector)
	sectorsGroup.GET("/:sectorID", middleware.RequirePermission(deps.RBACRepo, constants.ResourceSector, constants.ActionView), deps.SpatialHandler.GetSector)
	sectorsGroup.PUT("/:sectorID", middleware.RequirePermission(deps.RBACRepo, constants.ResourceSector, constants.ActionUpdate), deps.SpatialHandler.UpdateSector)
	sectorsGroup.DELETE("/:sectorID", middleware.RequirePermission(deps.RBACRepo, constants.ResourceSector, constants.ActionDelete), deps.SpatialHandler.DeleteSector)

	// Regions (protected) — P1-010; RequirePermission added UAT-006
	regionsGroup := v1.Group("/regions")
	regionsGroup.Use(middleware.AuthRequired(deps.TokenSvc, deps.AuthSvc))
	regionsGroup.GET("", middleware.RequirePermission(deps.RBACRepo, constants.ResourceRegion, constants.ActionView), deps.SpatialHandler.ListRegions)
	regionsGroup.POST("", middleware.RequirePermission(deps.RBACRepo, constants.ResourceRegion, constants.ActionCreate), deps.SpatialHandler.CreateRegion)
	regionsGroup.GET("/:regionID", middleware.RequirePermission(deps.RBACRepo, constants.ResourceRegion, constants.ActionView), deps.SpatialHandler.GetRegion)
	regionsGroup.PUT("/:regionID", middleware.RequirePermission(deps.RBACRepo, constants.ResourceRegion, constants.ActionUpdate), deps.SpatialHandler.UpdateRegion)
	regionsGroup.DELETE("/:regionID", middleware.RequirePermission(deps.RBACRepo, constants.ResourceRegion, constants.ActionDelete), deps.SpatialHandler.DeleteRegion)

	// River Basins (protected) — P1-010; RequirePermission added UAT-006
	riverBasinsGroup := v1.Group("/river-basins")
	riverBasinsGroup.Use(middleware.AuthRequired(deps.TokenSvc, deps.AuthSvc))
	riverBasinsGroup.GET("", middleware.RequirePermission(deps.RBACRepo, constants.ResourceRiverBasin, constants.ActionView), deps.SpatialHandler.ListRiverBasins)
	riverBasinsGroup.POST("", middleware.RequirePermission(deps.RBACRepo, constants.ResourceRiverBasin, constants.ActionCreate), deps.SpatialHandler.CreateRiverBasin)
	riverBasinsGroup.GET("/:riverBasinID", middleware.RequirePermission(deps.RBACRepo, constants.ResourceRiverBasin, constants.ActionView), deps.SpatialHandler.GetRiverBasin)
	riverBasinsGroup.PUT("/:riverBasinID", middleware.RequirePermission(deps.RBACRepo, constants.ResourceRiverBasin, constants.ActionUpdate), deps.SpatialHandler.UpdateRiverBasin)
	riverBasinsGroup.DELETE("/:riverBasinID", middleware.RequirePermission(deps.RBACRepo, constants.ResourceRiverBasin, constants.ActionDelete), deps.SpatialHandler.DeleteRiverBasin)

	// Baselines (protected) — P1-011
	baselinesGroup := v1.Group("/projects/:id/baselines")
	baselinesGroup.Use(middleware.AuthRequired(deps.TokenSvc, deps.AuthSvc))
	baselinesGroup.GET("", middleware.RequirePermission(deps.RBACRepo, constants.ResourceBaseline, constants.ActionView), deps.MonitoringHandler.ListBaselines)
	baselinesGroup.POST("", middleware.RequirePermission(deps.RBACRepo, constants.ResourceBaseline, constants.ActionCreate), deps.MonitoringHandler.CreateBaseline)
	baselinesGroup.GET("/:baselineID", middleware.RequirePermission(deps.RBACRepo, constants.ResourceBaseline, constants.ActionView), deps.MonitoringHandler.GetBaseline)
	baselinesGroup.PUT("/:baselineID", middleware.RequirePermission(deps.RBACRepo, constants.ResourceBaseline, constants.ActionUpdate), deps.MonitoringHandler.UpdateBaseline)
	baselinesGroup.DELETE("/:baselineID", middleware.RequirePermission(deps.RBACRepo, constants.ResourceBaseline, constants.ActionDelete), deps.MonitoringHandler.DeleteBaseline)

	// Snapshots (protected) — P1-011
	snapshotsGroup := v1.Group("/projects/:id/snapshots")
	snapshotsGroup.Use(middleware.AuthRequired(deps.TokenSvc, deps.AuthSvc))
	snapshotsGroup.GET("", middleware.RequirePermission(deps.RBACRepo, constants.ResourceSnapshot, constants.ActionView), deps.MonitoringHandler.ListSnapshots)
	snapshotsGroup.POST("", middleware.RequirePermission(deps.RBACRepo, constants.ResourceSnapshot, constants.ActionCreate), deps.MonitoringHandler.CreateSnapshot)
	snapshotsGroup.GET("/:snapshotID", middleware.RequirePermission(deps.RBACRepo, constants.ResourceSnapshot, constants.ActionView), deps.MonitoringHandler.GetSnapshot)
	snapshotsGroup.PUT("/:snapshotID", middleware.RequirePermission(deps.RBACRepo, constants.ResourceSnapshot, constants.ActionUpdate), deps.MonitoringHandler.UpdateSnapshot)
	snapshotsGroup.PATCH("/:snapshotID/status", middleware.RequirePermission(deps.RBACRepo, constants.ResourceSnapshot, constants.ActionUpdate), deps.MonitoringHandler.TransitionSnapshot)
	snapshotsGroup.DELETE("/:snapshotID", middleware.RequirePermission(deps.RBACRepo, constants.ResourceSnapshot, constants.ActionDelete), deps.MonitoringHandler.DeleteSnapshot)

	// Data validation queue (protected) - P1-012
	validationGroup := v1.Group("/validation-queue")
	validationGroup.Use(middleware.AuthRequired(deps.TokenSvc, deps.AuthSvc))
	validationGroup.GET("", middleware.RequirePermission(deps.RBACRepo, constants.ResourceValidationQueue, constants.ActionView), deps.DataQualityHandler.List)
	validationGroup.POST("", middleware.RequirePermission(deps.RBACRepo, constants.ResourceDataSubmission, constants.ActionCreate), deps.DataQualityHandler.Create)
	validationGroup.GET("/:submissionID", middleware.RequirePermission(deps.RBACRepo, constants.ResourceValidationQueue, constants.ActionView), deps.DataQualityHandler.Get)
	validationGroup.PATCH("/:submissionID/status", middleware.RequirePermission(deps.RBACRepo, constants.ResourceValidationQueue, constants.ActionApprove), deps.DataQualityHandler.Transition)

	// Field inspections and evidence (protected) - P1-013
	inspectionsGroup := v1.Group("/projects/:id/inspections")
	inspectionsGroup.Use(middleware.AuthRequired(deps.TokenSvc, deps.AuthSvc))
	inspectionsGroup.GET("", middleware.RequirePermission(deps.RBACRepo, constants.ResourceFieldInspection, constants.ActionView), deps.FieldHandler.List)
	inspectionsGroup.POST("", middleware.RequirePermission(deps.RBACRepo, constants.ResourceFieldInspection, constants.ActionCreate), deps.FieldHandler.Create)
	inspectionsGroup.GET("/:inspectionID/evidence/:evidenceID/download", middleware.RequirePermission(deps.RBACRepo, constants.ResourceFieldEvidence, constants.ActionView), deps.FieldHandler.Download)
	inspectionsGroup.PATCH("/:inspectionID/verification", middleware.RequirePermission(deps.RBACRepo, constants.ResourceFieldInspection, constants.ActionApprove), deps.FieldHandler.Verify)
	inspectionsGroup.DELETE("/:inspectionID", middleware.RequirePermission(deps.RBACRepo, constants.ResourceFieldInspection, constants.ActionDelete), deps.FieldHandler.Delete)
	inspectionsGroup.POST("/:inspectionID/evidence", middleware.RequirePermission(deps.RBACRepo, constants.ResourceFieldEvidence, constants.ActionCreate), deps.FieldHandler.AddEvidence)

	// Configurable project health score (protected) - P1-014
	healthFormulasGroup := v1.Group("/health-formulas")
	healthFormulasGroup.Use(middleware.AuthRequired(deps.TokenSvc, deps.AuthSvc))
	healthFormulasGroup.GET("", middleware.RequirePermission(deps.RBACRepo, constants.ResourceHealthFormula, constants.ActionView), deps.HealthHandler.ListFormulas)
	healthFormulasGroup.POST("", middleware.RequirePermission(deps.RBACRepo, constants.ResourceHealthFormula, constants.ActionCreate), deps.HealthHandler.CreateFormula)
	healthFormulasGroup.PATCH("/:formulaID/status", middleware.RequirePermission(deps.RBACRepo, constants.ResourceHealthFormula, constants.ActionApprove), deps.HealthHandler.TransitionFormula)
	healthGroup := v1.Group("/projects/:id/health")
	healthGroup.Use(middleware.AuthRequired(deps.TokenSvc, deps.AuthSvc))
	healthGroup.GET("", middleware.RequirePermission(deps.RBACRepo, constants.ResourceHealthSnapshot, constants.ActionView), deps.HealthHandler.ListSnapshots)
	healthGroup.POST("/calculate", middleware.RequirePermission(deps.RBACRepo, constants.ResourceHealthSnapshot, constants.ActionCreate), deps.HealthHandler.Calculate)
	commandCenterGroup := v1.Group("/command-center")
	commandCenterGroup.Use(middleware.AuthRequired(deps.TokenSvc, deps.AuthSvc))
	commandCenterGroup.GET("", middleware.RequirePermission(deps.RBACRepo, constants.ResourceCommandCenter, constants.ActionView), deps.CommandCenterHandler.Get)
	commandCenterGroup.GET("/escalations", middleware.RequirePermission(deps.RBACRepo, constants.ResourceCommandCenter, constants.ActionView), deps.CommandDecisionHandler.ListEscalations)
	commandCenterGroup.POST("/escalations", middleware.RequirePermission(deps.RBACRepo, constants.ResourceCommandCenter, constants.ActionUpdate), deps.CommandDecisionHandler.CreateEscalation)
	commandCenterGroup.PATCH("/escalations/:escalationID/status", middleware.RequirePermission(deps.RBACRepo, constants.ResourceCommandCenter, constants.ActionUpdate), deps.CommandDecisionHandler.UpdateEscalationStatus)
	commandCenterGroup.GET("/decisions", middleware.RequirePermission(deps.RBACRepo, constants.ResourceCommandCenter, constants.ActionView), deps.CommandDecisionHandler.ListDecisions)
	commandCenterGroup.POST("/decisions", middleware.RequirePermission(deps.RBACRepo, constants.ResourceCommandCenter, constants.ActionCreate), deps.CommandDecisionHandler.CreateDecision)
	commandCenterGroup.PATCH("/decisions/:decisionID/status", middleware.RequirePermission(deps.RBACRepo, constants.ResourceCommandCenter, constants.ActionUpdate), deps.CommandDecisionHandler.UpdateDecisionStatus)

	projectControlGroup := v1.Group("/projects/:id/control")
	projectControlGroup.Use(middleware.AuthRequired(deps.TokenSvc, deps.AuthSvc))
	projectControlGroup.GET("", middleware.RequirePermission(deps.RBACRepo, constants.ResourceCommandCenter, constants.ActionView), deps.ProjectControlHandler.Get)
	benefitGroup := v1.Group("/benefits")
	benefitGroup.Use(middleware.AuthRequired(deps.TokenSvc, deps.AuthSvc))
	benefitGroup.GET("", middleware.RequirePermission(deps.RBACRepo, constants.ResourceBenefit, constants.ActionView), deps.BenefitHandler.List)
	benefitGroup.POST("", middleware.RequirePermission(deps.RBACRepo, constants.ResourceBenefit, constants.ActionCreate), deps.BenefitHandler.Create)
	benefitGroup.GET("/summary", middleware.RequirePermission(deps.RBACRepo, constants.ResourceBenefit, constants.ActionView), deps.BenefitHandler.Summary)
	benefitGroup.GET("/:indicatorID", middleware.RequirePermission(deps.RBACRepo, constants.ResourceBenefit, constants.ActionView), deps.BenefitHandler.Get)
	benefitGroup.PUT("/:indicatorID", middleware.RequirePermission(deps.RBACRepo, constants.ResourceBenefit, constants.ActionUpdate), deps.BenefitHandler.Update)
	benefitGroup.DELETE("/:indicatorID", middleware.RequirePermission(deps.RBACRepo, constants.ResourceBenefit, constants.ActionDelete), deps.BenefitHandler.Delete)
	benefitGroup.GET("/:indicatorID/measurements", middleware.RequirePermission(deps.RBACRepo, constants.ResourceBenefit, constants.ActionView), deps.BenefitHandler.ListMeasurements)
	benefitGroup.POST("/:indicatorID/measurements", middleware.RequirePermission(deps.RBACRepo, constants.ResourceBenefit, constants.ActionCreate), deps.BenefitHandler.CreateMeasurement)
	benefitGroup.PUT("/:indicatorID/measurements/:measurementID", middleware.RequirePermission(deps.RBACRepo, constants.ResourceBenefit, constants.ActionUpdate), deps.BenefitHandler.UpdateMeasurement)
	benefitGroup.DELETE("/:indicatorID/measurements/:measurementID", middleware.RequirePermission(deps.RBACRepo, constants.ResourceBenefit, constants.ActionDelete), deps.BenefitHandler.DeleteMeasurement)
	benefitGroup.GET("/:indicatorID/aggregate", middleware.RequirePermission(deps.RBACRepo, constants.ResourceBenefit, constants.ActionView), deps.BenefitHandler.Aggregate)

	// Priority Scoring & Decision Support (protected) — P2-004
	priorityGroup := v1.Group("/priority")
	priorityGroup.Use(middleware.AuthRequired(deps.TokenSvc, deps.AuthSvc))
	priorityGroup.GET("/formulas", middleware.RequirePermission(deps.RBACRepo, constants.ResourcePriority, constants.ActionView), deps.PriorityHandler.ListFormulas)
	priorityGroup.POST("/formulas", middleware.RequirePermission(deps.RBACRepo, constants.ResourcePriority, constants.ActionCreate), deps.PriorityHandler.CreateFormula)
	priorityGroup.GET("/formulas/:id", middleware.RequirePermission(deps.RBACRepo, constants.ResourcePriority, constants.ActionView), deps.PriorityHandler.GetFormula)
	priorityGroup.PUT("/formulas/:id", middleware.RequirePermission(deps.RBACRepo, constants.ResourcePriority, constants.ActionUpdate), deps.PriorityHandler.UpdateFormula)
	priorityGroup.POST("/formulas/:id/activate", middleware.RequirePermission(deps.RBACRepo, constants.ResourcePriority, constants.ActionApprove), deps.PriorityHandler.ActivateFormula)
	priorityGroup.POST("/calculate", middleware.RequirePermission(deps.RBACRepo, constants.ResourcePriority, constants.ActionCreate), deps.PriorityHandler.Calculate)
	priorityGroup.POST("/batch-calculate", middleware.RequirePermission(deps.RBACRepo, constants.ResourcePriority, constants.ActionCreate), deps.PriorityHandler.BatchCalculate)
	priorityGroup.GET("/projects", middleware.RequirePermission(deps.RBACRepo, constants.ResourcePriority, constants.ActionView), deps.PriorityHandler.ListRanking)
	priorityGroup.GET("/projects/:projectID", middleware.RequirePermission(deps.RBACRepo, constants.ResourcePriority, constants.ActionView), deps.PriorityHandler.GetProjectScore)
	priorityGroup.GET("/projects/:projectID/explain", middleware.RequirePermission(deps.RBACRepo, constants.ResourcePriority, constants.ActionView), deps.PriorityHandler.ExplainProjectScore)

	// Program & Sector Dashboard Analytics (protected) — P2-006
	analyticsGroup := v1.Group("/analytics")
	analyticsGroup.Use(middleware.AuthRequired(deps.TokenSvc, deps.AuthSvc))
	analyticsGroup.GET("/programs", middleware.RequirePermission(deps.RBACRepo, constants.ResourceProgramDashboard, constants.ActionView), deps.ProgramDashboardHandler.ListPrograms)
	analyticsGroup.GET("/programs/:id", middleware.RequirePermission(deps.RBACRepo, constants.ResourceProgramDashboard, constants.ActionView), deps.ProgramDashboardHandler.GetProgram)
	analyticsGroup.GET("/sectors", middleware.RequirePermission(deps.RBACRepo, constants.ResourceProgramDashboard, constants.ActionView), deps.ProgramDashboardHandler.ListSectors)
	analyticsGroup.GET("/sectors/:id", middleware.RequirePermission(deps.RBACRepo, constants.ResourceProgramDashboard, constants.ActionView), deps.ProgramDashboardHandler.GetSector)

	// Level 1 Executive Dashboard (protected) — P2-007
	analyticsGroup.GET("/executive", middleware.RequirePermission(deps.RBACRepo, constants.ResourceExecutiveDashboard, constants.ActionView), deps.ExecutiveDashboardHandler.GetDashboard)

	// GIS Map (protected) — P2-008
	analyticsGroup.GET("/gis/projects", middleware.RequirePermission(deps.RBACRepo, constants.ResourceGISMap, constants.ActionView), deps.GISDashboardHandler.GetProjects)
	analyticsGroup.GET("/gis/summary", middleware.RequirePermission(deps.RBACRepo, constants.ResourceGISMap, constants.ActionView), deps.GISDashboardHandler.GetSummary)
	analyticsGroup.GET("/gis/projects/:id", middleware.RequirePermission(deps.RBACRepo, constants.ResourceGISMap, constants.ActionView), deps.GISDashboardHandler.GetProjectDetail)

	// Reporting Analytics — catalog, datasets, power BI config, export requests (P2-009)
	reportingAnalyticsGroup := analyticsGroup.Group("/reports")
	reportingAnalyticsGroup.Use(middleware.RequirePermission(deps.RBACRepo, constants.ResourceReport, constants.ActionView))
	deps.ReportingAnalyticsHandler.RegisterRoutes(reportingAnalyticsGroup)

	// CSV/Excel Import — P2-001
	importGroup := v1.Group("/imports")
	importGroup.Use(middleware.AuthRequired(deps.TokenSvc, deps.AuthSvc))
	importGroup.GET("/templates", middleware.RequirePermission(deps.RBACRepo, constants.ResourceImport, constants.ActionView), deps.ImportHandler.ListTemplates)
	importGroup.GET("/jobs", middleware.RequirePermission(deps.RBACRepo, constants.ResourceImport, constants.ActionView), deps.ImportHandler.ListJobs)
	importGroup.POST("/jobs", middleware.RequirePermission(deps.RBACRepo, constants.ResourceImport, constants.ActionCreate), deps.ImportHandler.CreateJob)
	importGroup.GET("/jobs/:id", middleware.RequirePermission(deps.RBACRepo, constants.ResourceImport, constants.ActionView), deps.ImportHandler.GetJob)
	importGroup.POST("/jobs/:id/validate", middleware.RequirePermission(deps.RBACRepo, constants.ResourceImport, constants.ActionCreate), deps.ImportHandler.ValidateJob)
	importGroup.POST("/jobs/:id/commit", middleware.RequirePermission(deps.RBACRepo, constants.ResourceImport, constants.ActionCreate), deps.ImportHandler.CommitJob)
	importGroup.POST("/jobs/:id/cancel", middleware.RequirePermission(deps.RBACRepo, constants.ResourceImport, constants.ActionCreate), deps.ImportHandler.CancelJob)
	importGroup.GET("/jobs/:id/rows", middleware.RequirePermission(deps.RBACRepo, constants.ResourceImport, constants.ActionView), deps.ImportHandler.ListRows)

	// Primavera P6 Integration — P2-011
	primaveraGroup := v1.Group("/integrations/primavera")
	primaveraGroup.Use(middleware.AuthRequired(deps.TokenSvc, deps.AuthSvc))
	primaveraGroup.GET("/runs", middleware.RequirePermission(deps.RBACRepo, constants.ResourcePrimaveraSync, constants.ActionView), deps.PrimaveraHandler.ListRuns)
	primaveraGroup.POST("/runs", middleware.RequirePermission(deps.RBACRepo, constants.ResourcePrimaveraSync, constants.ActionCreate), deps.PrimaveraHandler.CreateRun)
	primaveraGroup.GET("/runs/:id", middleware.RequirePermission(deps.RBACRepo, constants.ResourcePrimaveraSync, constants.ActionView), deps.PrimaveraHandler.GetRun)
	primaveraGroup.POST("/runs/:id/process", middleware.RequirePermission(deps.RBACRepo, constants.ResourcePrimaveraSync, constants.ActionCreate), deps.PrimaveraHandler.ProcessRun)
	primaveraGroup.POST("/runs/:id/cancel", middleware.RequirePermission(deps.RBACRepo, constants.ResourcePrimaveraSync, constants.ActionCreate), deps.PrimaveraHandler.CancelRun)
	primaveraGroup.GET("/runs/:id/mappings", middleware.RequirePermission(deps.RBACRepo, constants.ResourcePrimaveraSync, constants.ActionView), deps.PrimaveraHandler.ListMappings)

	// Government Connector Foundation — P2-002
	govGroup := v1.Group("/integrations/government")
	govGroup.Use(middleware.AuthRequired(deps.TokenSvc, deps.AuthSvc))
	govGroup.GET("/connectors", middleware.RequirePermission(deps.RBACRepo, constants.ResourceGovernmentConnector, constants.ActionView), deps.GovernmentHandler.ListConnectors)
	govGroup.GET("/connectors/:key", middleware.RequirePermission(deps.RBACRepo, constants.ResourceGovernmentConnector, constants.ActionView), deps.GovernmentHandler.GetConnector)
	govGroup.GET("/config", middleware.RequirePermission(deps.RBACRepo, constants.ResourceGovernmentConnector, constants.ActionView), deps.GovernmentHandler.GetConfig)
	govGroup.GET("/runs", middleware.RequirePermission(deps.RBACRepo, constants.ResourceGovernmentConnector, constants.ActionView), deps.GovernmentHandler.ListRuns)
	govGroup.POST("/runs", middleware.RequirePermission(deps.RBACRepo, constants.ResourceGovernmentConnector, constants.ActionCreate), deps.GovernmentHandler.CreateRun)
	govGroup.GET("/runs/:id", middleware.RequirePermission(deps.RBACRepo, constants.ResourceGovernmentConnector, constants.ActionView), deps.GovernmentHandler.GetRun)
	govGroup.POST("/runs/:id/cancel", middleware.RequirePermission(deps.RBACRepo, constants.ResourceGovernmentConnector, constants.ActionCreate), deps.GovernmentHandler.CancelRun)
	govGroup.GET("/runs/:id/records", middleware.RequirePermission(deps.RBACRepo, constants.ResourceGovernmentConnector, constants.ActionView), deps.GovernmentHandler.ListRecords)
	govGroup.GET("/mappings", middleware.RequirePermission(deps.RBACRepo, constants.ResourceGovernmentConnector, constants.ActionView), deps.GovernmentHandler.ListMappings)
	// Government Entity Resolution — P3-002
	govGroup.GET("/mappings/pending", middleware.RequirePermission(deps.RBACRepo, constants.ResourceGovernmentConnector, constants.ActionView), deps.GovernmentHandler.ListPendingMappings)
	govGroup.GET("/mappings/:id", middleware.RequirePermission(deps.RBACRepo, constants.ResourceGovernmentConnector, constants.ActionView), deps.GovernmentHandler.GetMapping)
	govGroup.GET("/mappings/:id/candidates", middleware.RequirePermission(deps.RBACRepo, constants.ResourceGovernmentConnector, constants.ActionView), deps.GovernmentHandler.GetMappingCandidates)
	govGroup.POST("/mappings/:id/match", middleware.RequirePermission(deps.RBACRepo, constants.ResourceGovernmentConnector, constants.ActionUpdate), deps.GovernmentHandler.MatchMapping)
	govGroup.POST("/mappings/:id/unmatch", middleware.RequirePermission(deps.RBACRepo, constants.ResourceGovernmentConnector, constants.ActionUpdate), deps.GovernmentHandler.UnmatchMapping)
	govGroup.POST("/mappings/:id/reject", middleware.RequirePermission(deps.RBACRepo, constants.ResourceGovernmentConnector, constants.ActionUpdate), deps.GovernmentHandler.RejectMapping)

	// BIM/Digital Twin Integration — P3-001
	bimGroup := v1.Group("/integrations/bim")
	bimGroup.Use(middleware.AuthRequired(deps.TokenSvc, deps.AuthSvc))
	bimGroup.GET("/models", middleware.RequirePermission(deps.RBACRepo, constants.ResourceBIMIntegration, constants.ActionView), deps.BIMHandler.ListModels)
	bimGroup.POST("/models", middleware.RequirePermission(deps.RBACRepo, constants.ResourceBIMIntegration, constants.ActionCreate), deps.BIMHandler.CreateModel)
	bimGroup.GET("/models/:id", middleware.RequirePermission(deps.RBACRepo, constants.ResourceBIMIntegration, constants.ActionView), deps.BIMHandler.GetModel)
	bimGroup.PATCH("/models/:id", middleware.RequirePermission(deps.RBACRepo, constants.ResourceBIMIntegration, constants.ActionUpdate), deps.BIMHandler.UpdateModel)
	bimGroup.DELETE("/models/:id", middleware.RequirePermission(deps.RBACRepo, constants.ResourceBIMIntegration, constants.ActionDelete), deps.BIMHandler.DeleteModel)
	bimGroup.GET("/models/:id/versions", middleware.RequirePermission(deps.RBACRepo, constants.ResourceBIMIntegration, constants.ActionView), deps.BIMHandler.ListVersions)
	bimGroup.POST("/models/:id/versions", middleware.RequirePermission(deps.RBACRepo, constants.ResourceBIMIntegration, constants.ActionCreate), deps.BIMHandler.AddVersion)
	bimGroup.GET("/models/:id/mappings", middleware.RequirePermission(deps.RBACRepo, constants.ResourceBIMIntegration, constants.ActionView), deps.BIMHandler.ListMappings)
	bimGroup.POST("/models/:id/mappings", middleware.RequirePermission(deps.RBACRepo, constants.ResourceBIMIntegration, constants.ActionCreate), deps.BIMHandler.LinkProject)
	bimGroup.DELETE("/models/:id/mappings/:project_id", middleware.RequirePermission(deps.RBACRepo, constants.ResourceBIMIntegration, constants.ActionDelete), deps.BIMHandler.UnlinkProject)

	// Data Governance — Official Validation & Approval Workflow (P3-003)
	governanceGroup := v1.Group("/governance")
	governanceGroup.Use(middleware.AuthRequired(deps.TokenSvc, deps.AuthSvc))
	governanceGroup.GET("/submissions", middleware.RequirePermission(deps.RBACRepo, constants.ResourceDataGovernance, constants.ActionView), deps.GovernanceHandler.ListSubmissions)
	governanceGroup.POST("/submissions", middleware.RequirePermission(deps.RBACRepo, constants.ResourceDataGovernance, constants.ActionCreate), deps.GovernanceHandler.CreateSubmission)
	governanceGroup.GET("/submissions/:id", middleware.RequirePermission(deps.RBACRepo, constants.ResourceDataGovernance, constants.ActionView), deps.GovernanceHandler.GetSubmission)
	governanceGroup.POST("/submissions/:id/submit", middleware.RequirePermission(deps.RBACRepo, constants.ResourceDataGovernance, constants.ActionUpdate), deps.GovernanceHandler.Submit)
	governanceGroup.POST("/submissions/:id/review", middleware.RequirePermission(deps.RBACRepo, constants.ResourceDataGovernance, constants.ActionApprove), deps.GovernanceHandler.StartReview)
	governanceGroup.POST("/submissions/:id/approve", middleware.RequirePermission(deps.RBACRepo, constants.ResourceDataGovernance, constants.ActionApprove), deps.GovernanceHandler.Approve)
	governanceGroup.POST("/submissions/:id/reject", middleware.RequirePermission(deps.RBACRepo, constants.ResourceDataGovernance, constants.ActionApprove), deps.GovernanceHandler.Reject)
	governanceGroup.POST("/submissions/:id/lock", middleware.RequirePermission(deps.RBACRepo, constants.ResourceDataGovernance, constants.ActionUpdate), deps.GovernanceHandler.Lock)
	governanceGroup.POST("/submissions/:id/cancel", middleware.RequirePermission(deps.RBACRepo, constants.ResourceDataGovernance, constants.ActionUpdate), deps.GovernanceHandler.Cancel)
	governanceGroup.GET("/lock-periods", middleware.RequirePermission(deps.RBACRepo, constants.ResourceDataGovernance, constants.ActionView), deps.GovernanceHandler.ListLockPeriods)
	governanceGroup.POST("/lock-periods", middleware.RequirePermission(deps.RBACRepo, constants.ResourceDataGovernance, constants.ActionCreate), deps.GovernanceHandler.CreateLockPeriod)
	governanceGroup.POST("/lock-periods/:id/lock", middleware.RequirePermission(deps.RBACRepo, constants.ResourceDataGovernance, constants.ActionUpdate), deps.GovernanceHandler.LockPeriod)

	// Audit Logs (protected, read-only) — UAT-003
	auditLogGroup := v1.Group("/audit-logs")
	auditLogGroup.Use(middleware.AuthRequired(deps.TokenSvc, deps.AuthSvc))
	auditLogGroup.Use(middleware.RequirePermission(deps.RBACRepo, constants.ResourceAuditLogs, constants.ActionView))
	deps.AuditLogHandler.RegisterRoutes(auditLogGroup)

	// Notifications — UAT-004
	notifGroup := v1.Group("/notifications")
	notifGroup.Use(middleware.AuthRequired(deps.TokenSvc, deps.AuthSvc))
	notifGroup.Use(middleware.RequirePermission(deps.RBACRepo, constants.ResourceNotification, constants.ActionView))
	deps.NotificationHandler.RegisterRoutes(notifGroup)

	return r
}

// Start runs the HTTP server with graceful shutdown.
func Start(cfg *config.Config, log *zap.Logger) error {
	deps, err := Wire(cfg, log)
	if err != nil {
		return err
	}

	engine := New(deps)

	srv := &http.Server{
		Addr:         ":" + cfg.App.Port,
		Handler:      engine,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in background
	go func() {
		log.Info("server started", zap.String("addr", srv.Addr), zap.String("env", cfg.App.Env))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("server listen error", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("server shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("server forced shutdown: %w", err)
	}

	database.Close(deps.DB)
	log.Info("server stopped gracefully")
	return nil
}

func getEnvOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
