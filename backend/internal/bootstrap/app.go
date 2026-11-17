package bootstrap

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	"github.com/harmanto-49/cankora/internal/platform/database"
	httprouter "github.com/harmanto-49/cankora/internal/platform/http/router"
	"github.com/harmanto-49/cankora/internal/project"
	projectcategory "github.com/harmanto-49/cankora/internal/project/category"
	"github.com/harmanto-49/cankora/internal/project/field"
	"github.com/harmanto-49/cankora/internal/project/health"
	"github.com/harmanto-49/cankora/internal/project/monitoring"
	"github.com/harmanto-49/cankora/internal/project/portfolio"
	"github.com/harmanto-49/cankora/internal/project/spatial"
	"github.com/harmanto-49/cankora/internal/project/vendorportal"
	"go.uber.org/zap"
)

// Wire initialises all dependencies and returns a populated Dependencies struct.
func Wire(cfg *config.Config, log *zap.Logger) (*httprouter.Dependencies, error) {
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
	authHandler := auth.NewHandler(authSvc, tokenSvc).WithAudit(auditWriter)
	projectHandler := project.NewHandler(projectSvc, auditWriter).WithDB(db)
	userHandler := user.NewHandler(userSvc).WithAudit(auditWriter)
	dashboardHandler := dashboard.NewHandler(db, log)
	reportHandler := report.NewHandler(db, log)
	orgHandler := organization.NewHandler(db, log).WithAudit(auditWriter)
	orgCRUDHandler := organization.NewOrgCRUDHandler(db, log)
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
	projectCategoryHandler := projectcategory.NewHandler(db, log)
	kpbuHandler := kpbu.NewHandler(db, cfg.Storage.LocalPath, cfg.Storage.MaxSizeBytes, auditWriter, log)
	vendorPortalHandler := vendorportal.NewHandler(db, cfg.Storage.LocalPath, cfg.Storage.MaxSizeBytes, auditWriter, log)

	_ = auditWriter

	return &httprouter.Dependencies{
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
		OrgCRUDHandler:            orgCRUDHandler,
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
		ProjectCategoryHandler:    projectCategoryHandler,
		KPBUHandler:               kpbuHandler,
		VendorPortalHandler:       vendorPortalHandler,
	}, nil
}

// Start runs the HTTP server with graceful shutdown.
func Start(cfg *config.Config, log *zap.Logger) error {
	deps, err := Wire(cfg, log)
	if err != nil {
		return err
	}

	engine := httprouter.New(deps)

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
