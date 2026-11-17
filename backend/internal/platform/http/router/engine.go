package router

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/harmanto-49/cankora/internal/identity/auth"
	"github.com/harmanto-49/cankora/internal/identity/rbac"
	"github.com/harmanto-49/cankora/internal/platform/http/middleware"
	"github.com/harmanto-49/cankora/internal/shared/constants"
)

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

	// Health checks are infrastructure routes, independent from domain wiring.
	RegisterHealth(r, deps.Config.App.Version, deps.Config.App.Env)

	// API v1
	v1 := r.Group("/api/v1")

	// Authentication routes and controls are owned by the HTTP router layer.
	RegisterAuth(v1, deps.AuthHandler, deps.TokenSvc, deps.AuthSvc)

	// Project aggregate routes are registered by the HTTP router boundary.
	RegisterProjects(v1, deps.ProjectHandler, deps.TokenSvc, deps.AuthSvc, deps.RBACRepo)
	projectsGroup := v1.Group("/projects")
	projectsGroup.Use(middleware.AuthRequired(deps.TokenSvc, deps.AuthSvc))
	RegisterProjectChildren(v1, deps.ProjectHandler, deps.TokenSvc, deps.AuthSvc, deps.RBACRepo)
	RegisterProjectControls(v1, deps.ProjectHandler, deps.TokenSvc, deps.AuthSvc, deps.RBACRepo)
	RegisterProjectFinance(v1, deps.ProjectHandler, deps.TokenSvc, deps.AuthSvc, deps.RBACRepo)
	RegisterProjectDocuments(v1, deps.ProjectHandler, deps.TokenSvc, deps.AuthSvc, deps.RBACRepo)
	RegisterProjectGovernance(v1, deps.ProjectHandler, deps.TokenSvc, deps.AuthSvc, deps.RBACRepo)

	RegisterAdmin(v1, deps.UserHandler, deps.GovernanceHandler, deps.AuditLogHandler, deps.NotificationHandler, deps.TokenSvc, deps.AuthSvc, deps.RBACRepo)

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

	// Organizations (protected) — top-level org CRUD
	orgsGroup := v1.Group("/organizations")
	orgsGroup.Use(middleware.AuthRequired(deps.TokenSvc, deps.AuthSvc))
	orgsGroup.GET("", middleware.RequirePermission(deps.RBACRepo, constants.ResourceOrganization, constants.ActionView), deps.OrgCRUDHandler.List)
	orgsGroup.POST("", middleware.RequirePermission(deps.RBACRepo, constants.ResourceOrganization, constants.ActionCreate), deps.OrgCRUDHandler.Create)
	orgsGroup.GET("/:orgID", middleware.RequirePermission(deps.RBACRepo, constants.ResourceOrganization, constants.ActionView), deps.OrgCRUDHandler.Get)
	orgsGroup.PUT("/:orgID", middleware.RequirePermission(deps.RBACRepo, constants.ResourceOrganization, constants.ActionUpdate), deps.OrgCRUDHandler.Update)

	// Roles (protected) — list roles for current org (closure avoids rbac→auth import cycle)
	rolesGroup := v1.Group("/roles")
	rolesGroup.Use(middleware.AuthRequired(deps.TokenSvc, deps.AuthSvc))
	rolesGroup.GET("", middleware.RequirePermission(deps.RBACRepo, constants.ResourceRole, constants.ActionView), func(c *gin.Context) {
		val, exists := c.Get(string(auth.ContextKeyClaims))
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "unauthorized"})
			return
		}
		claims, ok := val.(*auth.Claims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "unauthorized"})
			return
		}
		roles, err := deps.RBACRepo.ListRoles(c.Request.Context(), claims.OrganizationID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "internal error"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true, "message": "roles retrieved", "data": roles})
	})
	rolesGroup.POST("", middleware.RequirePermission(deps.RBACRepo, constants.ResourceRole, constants.ActionCreate), func(c *gin.Context) {
		val, exists := c.Get(string(auth.ContextKeyClaims))
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "unauthorized"})
			return
		}
		claims, ok := val.(*auth.Claims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "unauthorized"})
			return
		}
		var req struct {
			Code        string `json:"code" binding:"required"`
			Name        string `json:"name" binding:"required"`
			Description string `json:"description"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
			return
		}
		// Check duplicate code within org
		if _, err := deps.RBACRepo.FindRoleByCode(c.Request.Context(), claims.OrganizationID, req.Code); err == nil {
			c.JSON(http.StatusConflict, gin.H{"success": false, "message": "role code already exists"})
			return
		}
		role := &rbac.Role{
			OrganizationID: claims.OrganizationID,
			Code:           req.Code,
			Name:           req.Name,
			Description:    req.Description,
		}
		if err := deps.RBACRepo.CreateRole(c.Request.Context(), role); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "internal error"})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"success": true, "message": "role created", "data": role})
	})
	rolesGroup.PUT("/:roleID", middleware.RequirePermission(deps.RBACRepo, constants.ResourceRole, constants.ActionUpdate), func(c *gin.Context) {
		val, exists := c.Get(string(auth.ContextKeyClaims))
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "unauthorized"})
			return
		}
		claims, ok := val.(*auth.Claims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "unauthorized"})
			return
		}
		roleID, err := uuid.Parse(c.Param("roleID"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid role id"})
			return
		}
		role, err := deps.RBACRepo.FindRoleByID(c.Request.Context(), roleID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "role not found"})
			return
		}
		// Tenant guard
		if role.OrganizationID != claims.OrganizationID {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "role not found"})
			return
		}
		if role.IsSystem {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "message": "system roles cannot be modified"})
			return
		}
		var req struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
			return
		}
		if req.Name != "" {
			role.Name = req.Name
		}
		role.Description = req.Description
		if err := deps.RBACRepo.UpdateRole(c.Request.Context(), role); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "internal error"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true, "message": "role updated", "data": role})
	})
	rolesGroup.DELETE("/:roleID", middleware.RequirePermission(deps.RBACRepo, constants.ResourceRole, constants.ActionDelete), func(c *gin.Context) {
		val, exists := c.Get(string(auth.ContextKeyClaims))
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "unauthorized"})
			return
		}
		claims, ok := val.(*auth.Claims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "unauthorized"})
			return
		}
		roleID, err := uuid.Parse(c.Param("roleID"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "invalid role id"})
			return
		}
		role, err := deps.RBACRepo.FindRoleByID(c.Request.Context(), roleID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "role not found"})
			return
		}
		if role.OrganizationID != claims.OrganizationID {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "role not found"})
			return
		}
		if role.IsSystem {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "message": "system roles cannot be deleted"})
			return
		}
		if err := deps.RBACRepo.DeleteRole(c.Request.Context(), roleID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "internal error"})
			return
		}
		c.JSON(http.StatusNoContent, nil)
	})

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

	// Project Categories — PROJECT-FORM-002
	projectCategoryGroup := v1.Group("/project-categories")
	projectCategoryGroup.Use(middleware.AuthRequired(deps.TokenSvc, deps.AuthSvc))
	projectCategoryGroup.GET("", middleware.RequirePermission(deps.RBACRepo, constants.ResourceProjectCategory, constants.ActionView), deps.ProjectCategoryHandler.List)
	projectCategoryGroup.POST("", middleware.RequirePermission(deps.RBACRepo, constants.ResourceProjectCategory, constants.ActionCreate), deps.ProjectCategoryHandler.Create)
	projectCategoryGroup.GET("/:id", middleware.RequirePermission(deps.RBACRepo, constants.ResourceProjectCategory, constants.ActionView), deps.ProjectCategoryHandler.GetByID)
	projectCategoryGroup.PUT("/:id", middleware.RequirePermission(deps.RBACRepo, constants.ResourceProjectCategory, constants.ActionUpdate), deps.ProjectCategoryHandler.Update)
	projectCategoryGroup.DELETE("/:id", middleware.RequirePermission(deps.RBACRepo, constants.ResourceProjectCategory, constants.ActionDelete), deps.ProjectCategoryHandler.Delete)

	// KPBU monitoring — KPBU-K1 (see docs/planning/KPBU-GRAND-DESIGN.md)
	kpbuGroup := v1.Group("/kpbu/projects")
	kpbuGroup.Use(middleware.AuthRequired(deps.TokenSvc, deps.AuthSvc))
	kpbuPerm := func(action string) gin.HandlerFunc {
		return middleware.RequirePermission(deps.RBACRepo, constants.ResourceKPBUProject, action)
	}
	kpbuGroup.GET("", kpbuPerm(constants.ActionView), deps.KPBUHandler.ListProjects)
	kpbuGroup.POST("", kpbuPerm(constants.ActionCreate), deps.KPBUHandler.CreateProject)
	kpbuGroup.GET("/:id", kpbuPerm(constants.ActionView), deps.KPBUHandler.GetProject)
	kpbuGroup.PUT("/:id", kpbuPerm(constants.ActionUpdate), deps.KPBUHandler.UpdateProject)
	kpbuGroup.DELETE("/:id", kpbuPerm(constants.ActionDelete), deps.KPBUHandler.DeleteProject)
	kpbuGroup.POST("/:id/transition", kpbuPerm(constants.ActionUpdate), deps.KPBUHandler.TransitionProject)
	kpbuGroup.GET("/:id/stage-milestones", kpbuPerm(constants.ActionView), deps.KPBUHandler.ListStageMilestones)
	kpbuGroup.PUT("/:id/stage-milestones", kpbuPerm(constants.ActionUpdate), deps.KPBUHandler.UpsertStageMilestone)
	kpbuGroup.GET("/:id/parties", kpbuPerm(constants.ActionView), deps.KPBUHandler.ListParties)
	kpbuGroup.POST("/:id/parties", kpbuPerm(constants.ActionCreate), deps.KPBUHandler.CreateParty)
	kpbuGroup.PUT("/:id/parties/:partyID", kpbuPerm(constants.ActionUpdate), deps.KPBUHandler.UpdateParty)
	kpbuGroup.DELETE("/:id/parties/:partyID", kpbuPerm(constants.ActionDelete), deps.KPBUHandler.DeleteParty)
	// KPBU-K1.5: document checklist, documents with verification, guarantees
	kpbuGroup.GET("/:id/checklist", kpbuPerm(constants.ActionView), deps.KPBUHandler.GetChecklist)
	kpbuGroup.POST("/:id/documents", kpbuPerm(constants.ActionCreate), deps.KPBUHandler.UploadDocument)
	kpbuGroup.GET("/:id/documents/:docID/download", kpbuPerm(constants.ActionView), deps.KPBUHandler.DownloadDocument)
	kpbuGroup.POST("/:id/documents/:docID/review", kpbuPerm(constants.ActionApprove), deps.KPBUHandler.ReviewDocument)
	kpbuGroup.DELETE("/:id/documents/:docID", kpbuPerm(constants.ActionDelete), deps.KPBUHandler.DeleteDocument)
	kpbuGroup.GET("/:id/guarantees", kpbuPerm(constants.ActionView), deps.KPBUHandler.ListGuarantees)
	kpbuGroup.POST("/:id/guarantees", kpbuPerm(constants.ActionCreate), deps.KPBUHandler.CreateGuarantee)
	kpbuGroup.PUT("/:id/guarantees/:guaranteeID", kpbuPerm(constants.ActionUpdate), deps.KPBUHandler.UpdateGuarantee)
	kpbuGroup.DELETE("/:id/guarantees/:guaranteeID", kpbuPerm(constants.ActionDelete), deps.KPBUHandler.DeleteGuarantee)

	// KPBU-K2: financing structure, government support, availability payments
	fin := deps.KPBUHandler.Finance()
	kpbuGroup.GET("/:id/finance-summary", kpbuPerm(constants.ActionView), deps.KPBUHandler.FinanceSummary)
	kpbuGroup.GET("/:id/financings", kpbuPerm(constants.ActionView), fin.Financing.List)
	kpbuGroup.POST("/:id/financings", kpbuPerm(constants.ActionCreate), fin.Financing.Create)
	kpbuGroup.PUT("/:id/financings/:financingID", kpbuPerm(constants.ActionUpdate), fin.Financing.Update)
	kpbuGroup.DELETE("/:id/financings/:financingID", kpbuPerm(constants.ActionDelete), fin.Financing.Delete)
	kpbuGroup.GET("/:id/government-supports", kpbuPerm(constants.ActionView), fin.Support.List)
	kpbuGroup.POST("/:id/government-supports", kpbuPerm(constants.ActionCreate), fin.Support.Create)
	kpbuGroup.PUT("/:id/government-supports/:supportID", kpbuPerm(constants.ActionUpdate), fin.Support.Update)
	kpbuGroup.DELETE("/:id/government-supports/:supportID", kpbuPerm(constants.ActionDelete), fin.Support.Delete)
	kpbuGroup.GET("/:id/availability-payments", kpbuPerm(constants.ActionView), fin.Payment.List)
	kpbuGroup.POST("/:id/availability-payments", kpbuPerm(constants.ActionCreate), fin.Payment.Create)
	kpbuGroup.PUT("/:id/availability-payments/:paymentID", kpbuPerm(constants.ActionUpdate), fin.Payment.Update)
	kpbuGroup.DELETE("/:id/availability-payments/:paymentID", kpbuPerm(constants.ActionDelete), fin.Payment.Delete)

	// KPBU-K3: service KPIs, measurements, scorecard
	kpi := deps.KPBUHandler.KPIRoutes()
	kpbuGroup.GET("/:id/kpis", kpbuPerm(constants.ActionView), kpi.List)
	kpbuGroup.POST("/:id/kpis", kpbuPerm(constants.ActionCreate), kpi.Create)
	kpbuGroup.PUT("/:id/kpis/:kpiID", kpbuPerm(constants.ActionUpdate), kpi.Update)
	kpbuGroup.DELETE("/:id/kpis/:kpiID", kpbuPerm(constants.ActionDelete), kpi.Delete)
	kpbuGroup.GET("/:id/kpi-measurements", kpbuPerm(constants.ActionView), kpi.ListMeasurements)
	kpbuGroup.POST("/:id/kpi-measurements", kpbuPerm(constants.ActionCreate), kpi.CreateMeasurement)
	kpbuGroup.POST("/:id/kpi-measurements/:measurementID/review", kpbuPerm(constants.ActionApprove), kpi.ReviewMeasurement)
	kpbuGroup.GET("/:id/kpi-scorecard", kpbuPerm(constants.ActionView), kpi.Scorecard)

	kpbuReq := v1.Group("/kpbu/document-requirements")
	kpbuReq.Use(middleware.AuthRequired(deps.TokenSvc, deps.AuthSvc))
	kpbuReq.GET("", kpbuPerm(constants.ActionView), deps.KPBUHandler.ListRequirements)
	kpbuReq.POST("", kpbuPerm(constants.ActionCreate), deps.KPBUHandler.CreateRequirement)
	kpbuReq.PUT("/:reqID", kpbuPerm(constants.ActionUpdate), deps.KPBUHandler.UpdateRequirement)
	kpbuReq.DELETE("/:reqID", kpbuPerm(constants.ActionDelete), deps.KPBUHandler.DeleteRequirement)

	// Contractor portal — VENDOR-PORTAL-001. Vendor-side routes additionally
	// require the caller to be linked to a vendor (users.vendor_id) and only
	// expose that vendor's own contracts.
	vp := deps.VendorPortalHandler
	vendorPortal := v1.Group("/vendor-portal")
	vendorPortal.Use(middleware.AuthRequired(deps.TokenSvc, deps.AuthSvc))
	vpPerm := func(action string) gin.HandlerFunc {
		return middleware.RequirePermission(deps.RBACRepo, constants.ResourceVendorPortal, action)
	}
	vendorPortal.GET("/contracts", vpPerm(constants.ActionView), vp.MyContracts)
	vendorPortal.GET("/contracts/:contractID", vpPerm(constants.ActionView), vp.MyContract)
	vendorPortal.GET("/contracts/:contractID/progress-reports", vpPerm(constants.ActionView), vp.ListReports)
	vendorPortal.POST("/contracts/:contractID/progress-reports", vpPerm(constants.ActionCreate), vp.CreateReport)
	vendorPortal.GET("/contracts/:contractID/documents", vpPerm(constants.ActionView), vp.ListDocuments)
	vendorPortal.POST("/contracts/:contractID/documents", vpPerm(constants.ActionCreate), vp.UploadDocument)
	vendorPortal.GET("/documents/:documentID/download", vpPerm(constants.ActionView), vp.DownloadOwnDocument)

	// BUP portal: KPBU projects where the account's vendor is the BUP party.
	kh := deps.KPBUHandler
	vendorPortal.GET("/kpbu-projects", vpPerm(constants.ActionView), kh.PortalListProjects)
	bup := vendorPortal.Group("/kpbu-projects/:id")
	bup.Use(kh.PortalScope)
	bupKPI := kh.KPIRoutes()
	bup.GET("", vpPerm(constants.ActionView), kh.PortalGetProject)
	bup.GET("/checklist", vpPerm(constants.ActionView), kh.GetChecklist)
	bup.POST("/documents", vpPerm(constants.ActionCreate), kh.UploadDocument)
	bup.GET("/documents/:docID/download", vpPerm(constants.ActionView), kh.DownloadDocument)
	bup.GET("/kpis", vpPerm(constants.ActionView), bupKPI.List)
	bup.GET("/kpi-measurements", vpPerm(constants.ActionView), bupKPI.ListMeasurements)
	bup.POST("/kpi-measurements", vpPerm(constants.ActionCreate), bupKPI.CreateMeasurement)
	bup.GET("/availability-payments", vpPerm(constants.ActionView), kh.Finance().Payment.List)

	// Balai review queue for contractor submissions.
	vendorSubmissions := v1.Group("/vendor-submissions")
	vendorSubmissions.Use(middleware.AuthRequired(deps.TokenSvc, deps.AuthSvc))
	vsPerm := func(action string) gin.HandlerFunc {
		return middleware.RequirePermission(deps.RBACRepo, constants.ResourceVendorSubmission, action)
	}
	vendorSubmissions.GET("", vsPerm(constants.ActionView), vp.ListSubmissions)
	vendorSubmissions.POST("/progress-reports/:id/review", vsPerm(constants.ActionApprove), vp.ReviewReport)
	vendorSubmissions.POST("/documents/:id/review", vsPerm(constants.ActionApprove), vp.ReviewDocument)
	vendorSubmissions.GET("/documents/:id/download", vsPerm(constants.ActionView), vp.DownloadSubmittedDocument)

	// Contractor accounts (admin).
	vendorAccounts := v1.Group("/vendors/:vendorID/accounts")
	vendorAccounts.Use(middleware.AuthRequired(deps.TokenSvc, deps.AuthSvc))
	vendorAccounts.GET("", middleware.RequirePermission(deps.RBACRepo, constants.ResourceUsers, constants.ActionView), vp.ListAccounts)
	vendorAccounts.POST("", middleware.RequirePermission(deps.RBACRepo, constants.ResourceUsers, constants.ActionCreate), vp.CreateAccount)

	return r
}

func getEnvOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
