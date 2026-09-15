// Command reset-local clears demo/transactional data from a LOCAL PMO database
// while preserving the schema (the golang-migrate bookkeeping table).
//
// It is intended for local development and UAT resets only. Baseline data
// (default organization, permissions, roles, programs/sectors/regions/river
// basins, project categories and the default Super Admin user) is re-created
// afterwards by `go run ./cmd/seed` with SEED_SKIP_DEMO_USERS=true.
//
// Usage (run from the backend/ directory so .env is found):
//
//	go run ./cmd/reset-local --yes      # truncate application data
//	go run ./cmd/reset-local --verify   # assert the reset result (read-only)
//
// Safety guards (see guardLocalOnly):
//   - refuses APP_ENV=production
//   - refuses managed/cloud database hosts
//   - refuses non-local DB hosts
//   - requires DB_NAME to match RESET_LOCAL_EXPECTED_DB (default cankora_db)
//   - requires --yes for the destructive path
//
// Override with ALLOW_PRODUCTION_RESET=true only if you are absolutely sure.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/harmanto-49/cankora/internal/platform/config"
	"github.com/harmanto-49/cankora/internal/platform/database"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// -----------------------------------------------------------------------------
// Baseline tables — wiped then re-created by cmd/seed
// -----------------------------------------------------------------------------

// baselineTables are cleared and then re-created by `cmd/seed`. They are listed
// children-first so the delete intent stays FK-safe.
var baselineTables = []string{
	// children of users/roles/groups
	"user_roles",            // → users, roles
	"group_roles",           // → groups, roles
	"user_groups",           // → users, groups
	"user_scopes",           // → users
	"password_reset_tokens", // → users
	"user_sessions",         // → users

	// children of organizations / users
	"project_categories", // → organizations, users(created_by)
	"org_units",          // → organizations
	"river_basins",       // → organizations
	"regions",            // → organizations
	"sectors",            // → organizations
	"programs",           // → organizations

	// role/permission graph
	"role_permissions", // → roles, permissions
	"roles",            // → organizations
	"groups",           // → organizations

	// standalone lookup tables
	"health_formulas",
	"priority_formulas",
	"report_definitions",
	"permissions",

	// identity parents
	"users",
	"organizations",
}

// -----------------------------------------------------------------------------
// Transactional tables — cleared and NOT re-created
// -----------------------------------------------------------------------------

// transactionalTables are wiped and left empty. Ordered children-first so the
// delete sequence is FK-safe; the actual TRUNCATE uses CASCADE, which makes the
// order defensive rather than strictly required.
var transactionalTables = []string{
	// ── deepest children (nothing depends on them) ───────────────────────────
	"project_priority_score_components", // → project_priority_scores
	"priority_formula_components",       // → priority_formulas
	"bim_model_versions",                // → bim_models
	"bim_project_mappings",              // → bim_models, projects
	"data_submission_items",             // → data_submissions
	"import_rows",                       // → import_jobs
	"primavera_activity_mappings",       // → primavera_sync_runs
	"government_sync_records",           // → government_sync_runs
	"field_evidence",                    // → field_inspections
	"benefit_measurements",              // → benefit_indicators
	"task_assignments",                  // → tasks
	"project_teams",                     // → projects, users

	// ── project-scoped transactional ─────────────────────────────────────────
	"tasks",
	"issues",
	"risks",
	"milestones",
	"project_budgets",
	"contracts",
	"corrective_actions",
	"field_inspections",
	"project_documents",
	"project_periodic_reports",
	"project_progress_history",
	"project_baselines",
	"project_snapshots",
	"project_priority_scores",
	"benefit_indicators",
	"data_submissions",
	"data_lock_periods",
	"health_snapshots",
	"report_snapshots",
	"report_export_requests",
	"approval_requests",
	"command_escalations",
	"executive_decisions",
	"audit_logs",
	"notifications",
	"import_jobs",
	"government_sync_runs",
	"primavera_sync_runs",
	"bim_models",
	"government_external_mappings",
	"vendors",

	// ── projects last ────────────────────────────────────────────────────────
	"projects",
}

// preservedTables are never touched by this command.
var preservedTables = []string{
	"schema_migrations", // golang-migrate bookkeeping — keeps the schema intact
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("reset-local: %v", err)
	}
}

func run() error {
	confirm := flag.Bool("yes", false, "confirm the destructive reset")
	verifyOnly := flag.Bool("verify", false, "only run verification checks (read-only)")
	flag.Parse()

	if *verifyOnly {
		*confirm = true
	}
	if !*confirm {
		return errors.New("refusing to run without --yes")
	}

	logger, _ := zap.NewDevelopment()
	defer logger.Sync() //nolint:errcheck

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}
	if err := guardLocalOnly(cfg); err != nil {
		return err
	}

	db, err := database.Connect(cfg.Database.DSN, false, logger)
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	defer database.Close(db)

	ctx := context.Background()

	fmt.Println("── Reset target ──")
	fmt.Printf("   host=%s port=%s db=%s user=%s app_env=%s\n",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.Name,
		cfg.Database.User, cfg.App.Env)
	fmt.Println()

	if *verifyOnly {
		return verify(ctx, db)
	}

	return truncate(ctx, db)
}

// -----------------------------------------------------------------------------
// Truncate
// -----------------------------------------------------------------------------

func truncate(ctx context.Context, db *gorm.DB) error {
	existing, err := existingTables(ctx, db)
	if err != nil {
		return err
	}

	preserved := map[string]bool{}
	for _, t := range preservedTables {
		preserved[t] = true
	}

	// Ordered, de-duplicated plan. Missing tables are skipped so the command
	// keeps working across schema versions.
	var plan []string
	seen := map[string]bool{}
	for _, group := range [][]string{transactionalTables, baselineTables} {
		for _, t := range group {
			if seen[t] || preserved[t] || !existing[t] {
				continue
			}
			seen[t] = true
			plan = append(plan, t)
		}
	}
	if len(plan) == 0 {
		return errors.New("no known PMO tables found — is DB_NAME pointing at the PMO database?")
	}

	// Drift check: a public table outside the plan means this command is out of
	// date with the schema. Warn before doing anything destructive.
	var unknown []string
	for t := range existing {
		if preserved[t] || seen[t] {
			continue
		}
		unknown = append(unknown, t)
	}
	if len(unknown) > 0 {
		unknown = sortStrings(unknown)
		fmt.Printf("~ WARN: %d table(s) not covered by the reset plan: %s\n",
			len(unknown), strings.Join(unknown, ", "))
		fmt.Println("        these are emptied only if they depend on a truncated table (CASCADE).")
		fmt.Println()
	}

	fmt.Printf("→ TRUNCATE %d table(s) RESTART IDENTITY CASCADE...\n", len(plan))
	stmt := fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", quotedList(plan))
	if err := db.WithContext(ctx).Exec(stmt).Error; err != nil {
		return fmt.Errorf("truncate: %w", err)
	}

	// Confirm the tables we targeted are actually empty.
	var nonEmpty []string
	for _, t := range plan {
		n, err := countRows(ctx, db, t)
		if err != nil {
			continue
		}
		if n != 0 {
			nonEmpty = append(nonEmpty, fmt.Sprintf("%s(%d)", t, n))
		}
	}
	if len(nonEmpty) > 0 {
		return fmt.Errorf("tables not fully emptied: %s", strings.Join(nonEmpty, ", "))
	}

	// Drift check: a public table outside the plan may still hold application
	// data (e.g. a table added by a newer migration). Report it so the plan can
	// be updated instead of silently leaving stale rows behind.
	uncovered, err := uncoveredTables(ctx, db, seen)
	if err != nil {
		return err
	}
	var stale []string
	for _, t := range uncovered {
		n, err := countRows(ctx, db, t)
		if err != nil || n == 0 {
			continue
		}
		stale = append(stale, fmt.Sprintf("%s(%d)", t, n))
	}

	fmt.Println()
	fmt.Println("✓ Truncate complete")
	fmt.Printf("   cleared   : %d table(s)\n", len(plan))
	fmt.Printf("   preserved : %s\n", strings.Join(preservedTables, ", "))
	if len(uncovered) > 0 {
		fmt.Printf("   note      : %d table(s) outside the reset plan: %s\n",
			len(uncovered), strings.Join(uncovered, ", "))
	}
	if len(stale) > 0 {
		return fmt.Errorf(
			"table(s) outside the reset plan still hold data: %s — add them to transactionalTables/baselineTables",
			strings.Join(stale, ", "))
	}
	return nil
}

// -----------------------------------------------------------------------------
// Verify
// -----------------------------------------------------------------------------

type checkResult struct {
	name   string
	ok     bool
	detail string
}

func verify(ctx context.Context, db *gorm.DB) error {
	var checks []checkResult
	add := func(name string, ok bool, detail string) {
		checks = append(checks, checkResult{name: name, ok: ok, detail: detail})
	}

	// ── Super Admin user ─────────────────────────────────────────────────────
	users, err := scalarStrings(ctx, db, `SELECT email FROM users ORDER BY email`)
	if err != nil {
		return err
	}
	add("exactly one user remains", len(users) == 1,
		fmt.Sprintf("found %d: %s", len(users), strings.Join(users, ", ")))
	add("remaining user is admin@cankora.local",
		len(users) == 1 && users[0] == "admin@cankora.local",
		strings.Join(users, ", "))

	demoEmails := []string{
		"pmo@cankora.local", "pm@cankora.local", "officer@cankora.local",
		"executive@cankora.local", "auditor@cankora.local",
	}
	var foundDemo []string
	for _, e := range demoEmails {
		n, err := countWhere(ctx, db, "users", "email = ?", e)
		if err != nil {
			return err
		}
		if n > 0 {
			foundDemo = append(foundDemo, e)
		}
	}
	add("no demo users remain", len(foundDemo) == 0, strings.Join(foundDemo, ", "))

	// ── Super Admin role wiring ──────────────────────────────────────────────
	adminRoleCount, err := countRaw(ctx, db, `
		SELECT count(*) FROM user_roles ur
		JOIN users u ON u.id = ur.user_id
		JOIN roles r ON r.id = ur.role_id
		WHERE u.email = 'admin@cankora.local' AND r.code = 'SUPER_ADMIN'`)
	if err != nil {
		return err
	}
	add("admin has SUPER_ADMIN role", adminRoleCount >= 1, fmt.Sprintf("matches=%d", adminRoleCount))

	// ── Baseline config ──────────────────────────────────────────────────────
	orgCount, err := countRaw(ctx, db, `SELECT count(*) FROM organizations WHERE deleted_at IS NULL`)
	if err != nil {
		return err
	}
	add("default organization present", orgCount >= 1, fmt.Sprintf("organizations=%d", orgCount))

	roleCount, err := countRaw(ctx, db, `SELECT count(*) FROM roles WHERE deleted_at IS NULL`)
	if err != nil {
		return err
	}
	add("RBAC roles present (7)", roleCount == 7, fmt.Sprintf("roles=%d", roleCount))

	permCount, err := countRaw(ctx, db, `SELECT count(*) FROM permissions`)
	if err != nil {
		return err
	}
	add("permissions present", permCount > 0, fmt.Sprintf("permissions=%d", permCount))

	rpCount, err := countRaw(ctx, db, `SELECT count(*) FROM role_permissions`)
	if err != nil {
		return err
	}
	add("role permissions mapped", rpCount > 0, fmt.Sprintf("role_permissions=%d", rpCount))

	// ── Demo project data ────────────────────────────────────────────────────
	projTotal, err := countRaw(ctx, db, `SELECT count(*) FROM projects`)
	if err != nil {
		return err
	}
	add("no projects remain", projTotal == 0, fmt.Sprintf("projects=%d", projTotal))

	demoProj, err := countRaw(ctx, db, `SELECT count(*) FROM projects WHERE code LIKE 'DEMO-SDA-%'`)
	if err != nil {
		return err
	}
	add("no DEMO-SDA-* projects", demoProj == 0, fmt.Sprintf("demo projects=%d", demoProj))

	// ── Project child data ───────────────────────────────────────────────────
	childTables := []string{
		"tasks", "milestones", "issues", "risks", "project_budgets", "contracts",
		"project_documents", "project_teams", "project_progress_history",
		"project_periodic_reports", "corrective_actions", "field_inspections",
		"field_evidence", "benefit_indicators", "benefit_measurements",
		"project_baselines", "project_snapshots", "health_snapshots",
		"project_priority_scores", "report_export_requests", "data_submissions",
		"import_jobs", "government_sync_runs", "primavera_sync_runs", "bim_models",
		"vendors", "audit_logs", "notifications", "command_escalations",
		"executive_decisions", "approval_requests",
	}
	var dirty []string
	for _, t := range childTables {
		n, err := countRows(ctx, db, t)
		if err != nil {
			continue // table may not exist in older schemas
		}
		if n != 0 {
			dirty = append(dirty, fmt.Sprintf("%s=%d", t, n))
		}
	}
	add("project child data empty", len(dirty) == 0, strings.Join(dirty, ", "))

	// ── Report ───────────────────────────────────────────────────────────────
	fmt.Println("── Verification ──")
	failed := 0
	for _, c := range checks {
		mark := "✓ PASS"
		if !c.ok {
			mark = "✗ FAIL"
			failed++
		}
		line := fmt.Sprintf(" %s  %s", mark, c.name)
		if c.detail != "" {
			line += fmt.Sprintf("  (%s)", c.detail)
		}
		fmt.Println(line)
	}
	fmt.Println()
	fmt.Printf(" %d/%d checks passed\n", len(checks)-failed, len(checks))

	if failed > 0 {
		return fmt.Errorf("%d verification check(s) failed", failed)
	}
	fmt.Println("✓ Local database contains only the default Super Admin baseline")
	return nil
}

// -----------------------------------------------------------------------------
// Guards
// -----------------------------------------------------------------------------

var cloudHostMarkers = []string{
	"rds.amazonaws", "supabase", "neon.tech", "railway.app",
	"azure.com", "cockroachlabs", "render.com", "digitalocean",
}

func guardLocalOnly(cfg *config.Config) error {
	host := strings.ToLower(strings.TrimSpace(cfg.Database.Host))
	env := strings.ToLower(strings.TrimSpace(cfg.App.Env))
	allowOverride := strings.EqualFold(os.Getenv("ALLOW_PRODUCTION_RESET"), "true")

	if !allowOverride {
		if env == "production" || env == "prod" {
			return errors.New(
				"APP_ENV=production — refusing to reset " +
					"(set ALLOW_PRODUCTION_RESET=true only if you are absolutely sure)")
		}
		for _, m := range cloudHostMarkers {
			if strings.Contains(host, m) {
				return fmt.Errorf("DB_HOST %q looks like a managed/cloud database — refusing to reset", host)
			}
		}
		if !isLocalHost(host) {
			return fmt.Errorf(
				"DB_HOST %q is not a local host — refusing to reset "+
					"(set ALLOW_PRODUCTION_RESET=true to override)", host)
		}
	}

	expected := os.Getenv("RESET_LOCAL_EXPECTED_DB")
	if expected == "" {
		expected = "cankora_db"
	}
	if cfg.Database.Name != expected {
		return fmt.Errorf(
			"DB_NAME %q does not match RESET_LOCAL_EXPECTED_DB %q — refusing to reset",
			cfg.Database.Name, expected)
	}

	return nil
}

func isLocalHost(host string) bool {
	switch host {
	case "localhost", "127.0.0.1", "::1", "0.0.0.0", "host.docker.internal", "db", "postgres":
		return true
	}
	return strings.HasSuffix(host, ".local")
}

// -----------------------------------------------------------------------------
// Query helpers
// -----------------------------------------------------------------------------

func existingTables(ctx context.Context, db *gorm.DB) (map[string]bool, error) {
	var names []string
	err := db.WithContext(ctx).Raw(
		`SELECT table_name FROM information_schema.tables
		  WHERE table_schema = 'public' AND table_type = 'BASE TABLE'`,
	).Scan(&names).Error
	if err != nil {
		return nil, fmt.Errorf("list tables: %w", err)
	}
	out := make(map[string]bool, len(names))
	for _, n := range names {
		out[n] = true
	}
	return out, nil
}

// uncoveredTables returns every public base table that is neither in the reset
// plan nor in preservedTables, sorted for stable output.
func uncoveredTables(ctx context.Context, db *gorm.DB, planned map[string]bool) ([]string, error) {
	existing, err := existingTables(ctx, db)
	if err != nil {
		return nil, err
	}
	preserved := map[string]bool{}
	for _, t := range preservedTables {
		preserved[t] = true
	}
	var out []string
	for t := range existing {
		if planned[t] || preserved[t] {
			continue
		}
		out = append(out, t)
	}
	return sortStrings(out), nil
}

func countRows(ctx context.Context, db *gorm.DB, table string) (int64, error) {
	if !validIdent(table) {
		return 0, fmt.Errorf("invalid table name %q", table)
	}
	var n int64
	err := db.WithContext(ctx).
		Raw(fmt.Sprintf("SELECT count(*) FROM %s", quoteIdent(table))).
		Scan(&n).Error
	return n, err
}

func countWhere(ctx context.Context, db *gorm.DB, table, where string, args ...any) (int64, error) {
	if !validIdent(table) {
		return 0, fmt.Errorf("invalid table name %q", table)
	}
	var n int64
	err := db.WithContext(ctx).
		Raw(fmt.Sprintf("SELECT count(*) FROM %s WHERE %s", quoteIdent(table), where), args...).
		Scan(&n).Error
	return n, err
}

func countRaw(ctx context.Context, db *gorm.DB, query string) (int64, error) {
	var n int64
	err := db.WithContext(ctx).Raw(query).Scan(&n).Error
	return n, err
}

func scalarStrings(ctx context.Context, db *gorm.DB, query string) ([]string, error) {
	var out []string
	err := db.WithContext(ctx).Raw(query).Scan(&out).Error
	return out, err
}

// validIdent guards against SQL injection through table names. Every table name
// is a compile-time constant, but the check keeps the helper safe.
func validIdent(name string) bool {
	if name == "" {
		return false
	}
	for _, r := range name {
		if (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != '_' {
			return false
		}
	}
	return true
}

func quoteIdent(ident string) string {
	return `"` + strings.ReplaceAll(ident, `"`, `""`) + `"`
}

func quotedList(names []string) string {
	quoted := make([]string, 0, len(names))
	for _, n := range names {
		quoted = append(quoted, quoteIdent(n))
	}
	return strings.Join(quoted, ", ")
}

func sortStrings(in []string) []string {
	out := make([]string, len(in))
	copy(out, in)
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j] < out[j-1]; j-- {
			out[j], out[j-1] = out[j-1], out[j]
		}
	}
	return out
}
