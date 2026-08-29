package main

import (
	"context"
	"log"

	"github.com/google/uuid"
	"github.com/harmanto-49/cankora/internal/core/auth"
	"github.com/harmanto-49/cankora/internal/core/organization"
	"github.com/harmanto-49/cankora/internal/core/rbac"
	"github.com/harmanto-49/cankora/internal/platform/config"
	"github.com/harmanto-49/cankora/internal/platform/database"
	"github.com/harmanto-49/cankora/internal/shared/constants"
	"github.com/harmanto-49/cankora/internal/shared/utils"
	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync() //nolint:errcheck

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("seed: config: %v", err)
	}

	db, err := database.Connect(cfg.Database.DSN, true, logger)
	if err != nil {
		log.Fatalf("seed: database: %v", err)
	}
	defer database.Close(db)

	ctx := context.Background()

	// ------------------------------------------------------------------
	// 1. Default organization
	// ------------------------------------------------------------------
	org := &organization.Organization{}
	if err := db.WithContext(ctx).
		Where("code = ?", "CANKORA").
		Attrs(organization.Organization{
			ID:        uuid.New(),
			Code:      "CANKORA",
			Name:      "CANKORA Default Organization",
			ShortName: "CANKORA",
			IsActive:  true,
		}).
		FirstOrCreate(org).Error; err != nil {
		log.Fatalf("seed: create organization: %v", err)
	}
	logger.Info("organization seeded", zap.String("id", org.ID.String()))

	// ------------------------------------------------------------------
	// 2. Default permissions (resource × action matrix)
	// ------------------------------------------------------------------
	resources := []string{
		constants.ResourceProjects, constants.ResourceTasks, constants.ResourceMilestones,
		constants.ResourceIssues, constants.ResourceRisks, constants.ResourceBudgets,
		constants.ResourceVendors, constants.ResourceContracts, constants.ResourceDocuments,
		constants.ResourceTeam, constants.ResourceUsers, constants.ResourceRoles,
		constants.ResourceOrganizations, constants.ResourceAuditLogs, constants.ResourceReports,
		constants.ResourceDataSubmission, constants.ResourceValidationQueue,
		constants.ResourceFieldInspection, constants.ResourceFieldEvidence,
		constants.ResourceHealthFormula, constants.ResourceHealthSnapshot,
		constants.ResourceCommandCenter,
		constants.ResourceBenefit,
		constants.ResourcePriority,
		constants.ResourceProgramDashboard,
		constants.ResourceExecutiveDashboard,
		constants.ResourceGISMap,
		constants.ResourceImport,
		constants.ResourcePrimaveraSync,
		constants.ResourceGovernmentConnector,
		constants.ResourceBIMIntegration,
		constants.ResourceDataGovernance,
		constants.ResourceNotification,
		constants.ResourceSector,
		constants.ResourceRegion,
		constants.ResourceRiverBasin,
		constants.ResourceCorrectiveAction,
		constants.ResourcePeriodicReport,
		constants.ResourceOrgUnit,
	}
	actions := []string{
		constants.ActionView, constants.ActionCreate, constants.ActionUpdate,
		constants.ActionDelete, constants.ActionApprove, constants.ActionExport,
	}

	permMap := make(map[string]*rbac.Permission)
	for _, res := range resources {
		for _, act := range actions {
			key := res + ":" + act
			perm := &rbac.Permission{}
			if err := db.WithContext(ctx).
				Where("resource = ? AND action = ?", res, act).
				Attrs(rbac.Permission{
					ID:       uuid.New(),
					Resource: res,
					Action:   act,
				}).
				FirstOrCreate(perm).Error; err != nil {
				log.Fatalf("seed: create permission %s: %v", key, err)
			}
			permMap[key] = perm
		}
	}
	logger.Info("permissions seeded", zap.Int("count", len(permMap)))

	// ------------------------------------------------------------------
	// 3. Default roles
	// ------------------------------------------------------------------
	type roleDef struct {
		Code        string
		Name        string
		Description string
		AllPerms    bool     // true = all permissions
		Resources   []string // if not AllPerms, these resources get all actions
	}

	roleDefs := []roleDef{
		{
			Code:     constants.RoleSuperAdmin,
			Name:     "Super Administrator",
			AllPerms: true,
		},
		{
			Code:     constants.RoleAdmin,
			Name:     "Administrator",
			AllPerms: true,
		},
		{
			Code:      constants.RolePMO,
			Name:      "PMO",
			Resources: resources, // all resources, all actions
		},
		{
			Code: constants.RoleProjectManager,
			Name: "Project Manager",
			Resources: []string{
				constants.ResourceProjects, constants.ResourceTasks, constants.ResourceMilestones,
				constants.ResourceIssues, constants.ResourceRisks, constants.ResourceBudgets,
				constants.ResourceVendors, constants.ResourceContracts, constants.ResourceDocuments,
				constants.ResourceTeam, constants.ResourceNotification,
			},
		},
		{
			Code: constants.RoleProjectOfficer,
			Name: "Project Officer",
			Resources: []string{
				constants.ResourceProjects, constants.ResourceTasks, constants.ResourceMilestones,
				constants.ResourceIssues, constants.ResourceDocuments, constants.ResourceNotification,
			},
		},
		{
			Code: constants.RoleExecutiveViewer,
			Name: "Executive Viewer",
			Resources: []string{
				constants.ResourceProjects, constants.ResourceReports, constants.ResourcePriority,
				constants.ResourceProgramDashboard, constants.ResourceExecutiveDashboard,
				constants.ResourceGISMap, constants.ResourceNotification,
			},
		},
		{
			Code: constants.RoleAuditor,
			Name: "Auditor",
			Resources: []string{
				constants.ResourceAuditLogs, constants.ResourceProjects, constants.ResourceReports,
				constants.ResourceNotification,
			},
		},
	}

	roleMap := make(map[string]*rbac.Role)
	for _, rd := range roleDefs {
		role := &rbac.Role{}
		if err := db.WithContext(ctx).
			Where("organization_id = ? AND code = ?", org.ID, rd.Code).
			Attrs(rbac.Role{
				ID:             uuid.New(),
				OrganizationID: org.ID,
				Code:           rd.Code,
				Name:           rd.Name,
				IsSystem:       true,
			}).
			FirstOrCreate(role).Error; err != nil {
			log.Fatalf("seed: create role %s: %v", rd.Code, err)
		}
		roleMap[rd.Code] = role

		// Assign permissions
		var permIDs []uuid.UUID
		for _, res := range resources {
			if rd.AllPerms || contains(rd.Resources, res) {
				for _, act := range actions {
					key := res + ":" + act
					if p, ok := permMap[key]; ok {
						permIDs = append(permIDs, p.ID)
					}
				}
			}
		}
		// Auditor & ExecutiveViewer: view + export only
		if rd.Code == constants.RoleAuditor || rd.Code == constants.RoleExecutiveViewer {
			permIDs = nil
			for _, res := range rd.Resources {
				for _, act := range []string{constants.ActionView, constants.ActionExport} {
					key := res + ":" + act
					if p, ok := permMap[key]; ok {
						permIDs = append(permIDs, p.ID)
					}
				}
			}
		}

		// Sync role permissions
		if err := db.WithContext(ctx).Where("role_id = ?", role.ID).Delete(&rbac.RolePermission{}).Error; err != nil {
			log.Fatalf("seed: clear role permissions %s: %v", rd.Code, err)
		}
		for _, pid := range permIDs {
			rp := rbac.RolePermission{RoleID: role.ID, PermissionID: pid}
			if err := db.WithContext(ctx).
				Where("role_id = ? AND permission_id = ?", role.ID, pid).
				FirstOrCreate(&rp).Error; err != nil {
				log.Fatalf("seed: assign permission to role %s: %v", rd.Code, err)
			}
		}
	}
	logger.Info("roles seeded", zap.Int("count", len(roleMap)))

	// ------------------------------------------------------------------
	// 4. Super Admin user
	// ------------------------------------------------------------------
	hash, err := utils.HashPassword("Admin@Cankora2024!")
	if err != nil {
		log.Fatalf("seed: hash password: %v", err)
	}

	adminUser := &auth.User{}
	if err := db.WithContext(ctx).
		Where("email = ?", "admin@cankora.local").
		Attrs(auth.User{
			ID:             uuid.New(),
			OrganizationID: org.ID,
			FirstName:      "Super",
			LastName:       "Admin",
			Email:          "admin@cankora.local",
			PasswordHash:   hash,
			IsActive:       true,
			MustChangePwd:  false,
		}).
		FirstOrCreate(adminUser).Error; err != nil {
		log.Fatalf("seed: create super admin: %v", err)
	}

	if adminUser.OrganizationID != org.ID || !adminUser.IsActive || adminUser.MustChangePwd {
		adminUser.OrganizationID = org.ID
		adminUser.IsActive = true
		adminUser.MustChangePwd = false
	}
	if err := utils.CheckPassword("Admin@Cankora2024!", adminUser.PasswordHash); err != nil {
		adminUser.PasswordHash = hash
	}
	if adminUser.LoginFailed != 0 || adminUser.LockedUntil != nil {
		adminUser.LoginFailed = 0
		adminUser.LockedUntil = nil
	}
	if err := db.WithContext(ctx).Save(adminUser).Error; err != nil {
		log.Fatalf("seed: update super admin: %v", err)
	}

	// Assign Super Admin role
	superAdminRole := roleMap[constants.RoleSuperAdmin]
	adminRole := rbac.UserRole{
		UserID: adminUser.ID,
		RoleID: superAdminRole.ID,
	}
	if err := db.WithContext(ctx).
		Where(adminRole).
		FirstOrCreate(&adminRole).Error; err != nil {
		log.Fatalf("seed: assign super admin role: %v", err)
	}

	logger.Info("super admin seeded",
		zap.String("email", "admin@cankora.local"),
		zap.String("password", "Admin@Cankora2024!"),
	)

	// ------------------------------------------------------------------
	// 5. Demo programs & sectors (P2-006)
	// ------------------------------------------------------------------
	type ProgramRecord struct {
		ID          uuid.UUID
		Code        string
		Name        string
		Description string
		FiscalYear  int
		IsActive    bool
		SortOrder   int
	}
	programDefs := []struct {
		Code string
		Name string
		Desc string
	}{
		{"PRG-BND", "Bendungan", "Program pembangunan dan rehabilitasi bendungan"},
		{"PRG-IRG", "Irigasi", "Program pengembangan dan rehabilitasi jaringan irigasi"},
		{"PRG-BNJ", "Pengendalian Banjir", "Program pengendalian banjir dan pengamanan pantai"},
		{"PRG-ABK", "Air Baku", "Program penyediaan air baku untuk kebutuhan domestik dan industri"},
		{"PRG-PTN", "Pertanian", "Program dukungan infrastruktur untuk sektor pertanian"},
	}
	programIDs := make([]uuid.UUID, len(programDefs))
	for i, pd := range programDefs {
		var existing struct{ ID uuid.UUID }
		err := db.WithContext(ctx).Table("programs").
			Select("id").
			Where("organization_id = ? AND code = ? AND deleted_at IS NULL", org.ID, pd.Code).
			First(&existing).Error
		if err == nil {
			programIDs[i] = existing.ID
			continue
		}
		newID := uuid.New()
		programIDs[i] = newID
		if err := db.WithContext(ctx).Exec(`
			INSERT INTO programs (id, organization_id, code, name, description, fiscal_year, is_active, sort_order)
			VALUES (?, ?, ?, ?, ?, 2025, true, ?)`,
			newID, org.ID, pd.Code, pd.Name, pd.Desc, i+1,
		).Error; err != nil {
			log.Fatalf("seed: create program %s: %v", pd.Code, err)
		}
	}
	logger.Info("programs seeded", zap.Int("count", len(programDefs)))

	sectorDefs := []struct {
		Code string
		Name string
		Desc string
	}{
		{"SEC-SDA", "Sumber Daya Air", "Sektor pengelolaan sumber daya air"},
		{"SEC-IRG", "Irigasi & Rawa", "Sektor irigasi dan rawa"},
		{"SEC-BNJ", "Banjir & Drainase", "Sektor pengendalian banjir dan drainase"},
		{"SEC-ABK", "Air Baku & Sanitasi", "Sektor air baku dan sanitasi"},
		{"SEC-PTN", "Pertanian & Perdesaan", "Sektor infrastruktur pertanian dan perdesaan"},
	}
	sectorIDs := make([]uuid.UUID, len(sectorDefs))
	for i, sd := range sectorDefs {
		var existing struct{ ID uuid.UUID }
		err := db.WithContext(ctx).Table("sectors").
			Select("id").
			Where("organization_id = ? AND code = ? AND deleted_at IS NULL", org.ID, sd.Code).
			First(&existing).Error
		if err == nil {
			sectorIDs[i] = existing.ID
			continue
		}
		newID := uuid.New()
		sectorIDs[i] = newID
		if err := db.WithContext(ctx).Exec(`
			INSERT INTO sectors (id, organization_id, code, name, description, is_active, sort_order)
			VALUES (?, ?, ?, ?, ?, true, ?)`,
			newID, org.ID, sd.Code, sd.Name, sd.Desc, i+1,
		).Error; err != nil {
			log.Fatalf("seed: create sector %s: %v", sd.Code, err)
		}
	}
	logger.Info("sectors seeded", zap.Int("count", len(sectorDefs)))

	// Link existing projects to programs & sectors (round-robin, only if not already set)
	var projectIDs []uuid.UUID
	db.WithContext(ctx).Table("projects").
		Select("id").
		Where("organization_id = ? AND deleted_at IS NULL AND program_id IS NULL", org.ID).
		Pluck("id", &projectIDs)
	for i, pid := range projectIDs {
		pIdx := i % len(programIDs)
		sIdx := i % len(sectorIDs)
		if err := db.WithContext(ctx).Exec(
			"UPDATE projects SET program_id = ?, sector_id = ? WHERE id = ?",
			programIDs[pIdx], sectorIDs[sIdx], pid,
		).Error; err != nil {
			logger.Warn("seed: update project program/sector", zap.Error(err))
		}
	}
	logger.Info("projects linked to programs/sectors", zap.Int("count", len(projectIDs)))

	// ------------------------------------------------------------------
	// 6. Demo users per role (UAT-006 — idempotent)
	// ------------------------------------------------------------------
	type demoUserDef struct {
		Email     string
		FirstName string
		LastName  string
		RoleCode  string
		Password  string // same simple UAT password for all demo accounts
	}

	demoUsers := []demoUserDef{
		{Email: "pmo@cankora.local", FirstName: "Demo", LastName: "PMO Admin", RoleCode: constants.RolePMO, Password: "Demo@Cankora2024!"},
		{Email: "pm@cankora.local", FirstName: "Demo", LastName: "Project Manager", RoleCode: constants.RoleProjectManager, Password: "Demo@Cankora2024!"},
		{Email: "officer@cankora.local", FirstName: "Demo", LastName: "Project Officer", RoleCode: constants.RoleProjectOfficer, Password: "Demo@Cankora2024!"},
		{Email: "executive@cankora.local", FirstName: "Demo", LastName: "Executive Viewer", RoleCode: constants.RoleExecutiveViewer, Password: "Demo@Cankora2024!"},
		{Email: "auditor@cankora.local", FirstName: "Demo", LastName: "Auditor", RoleCode: constants.RoleAuditor, Password: "Demo@Cankora2024!"},
	}

	for _, du := range demoUsers {
		hashed, err := utils.HashPassword(du.Password)
		if err != nil {
			log.Fatalf("seed: hash demo password for %s: %v", du.Email, err)
		}
		demoUser := &auth.User{}
		if err := db.WithContext(ctx).
			Where("organization_id = ? AND email = ?", org.ID, du.Email).
			Attrs(auth.User{
				ID:             uuid.New(),
				OrganizationID: org.ID,
				Email:          du.Email,
				FirstName:      du.FirstName,
				LastName:       du.LastName,
				PasswordHash:   hashed,
				IsActive:       true,
				MustChangePwd:  false,
			}).
			FirstOrCreate(demoUser).Error; err != nil {
			log.Fatalf("seed: create demo user %s: %v", du.Email, err)
		}

		// Assign role
		demoRole, ok := roleMap[du.RoleCode]
		if !ok {
			log.Fatalf("seed: role not found for demo user %s: %s", du.Email, du.RoleCode)
		}
		userRoleRecord := rbac.UserRole{
			UserID: demoUser.ID,
			RoleID: demoRole.ID,
		}
		if err := db.WithContext(ctx).
			Where(userRoleRecord).
			FirstOrCreate(&userRoleRecord).Error; err != nil {
			log.Fatalf("seed: assign role for demo user %s: %v", du.Email, err)
		}
		logger.Info("demo user seeded", zap.String("email", du.Email), zap.String("role", du.RoleCode))
	}

	logger.Info("seed completed successfully")
}

func contains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
