package vendorportal

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/harmanto-49/cankora/internal/identity/auth"
	"github.com/harmanto-49/cankora/internal/identity/rbac"
	"github.com/harmanto-49/cankora/internal/project"
	"github.com/harmanto-49/cankora/internal/shared/constants"
	"github.com/harmanto-49/cankora/internal/shared/utils"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

// Service implements the contractor portal. Every query is scoped by
// organization_id; contractor-side queries are also scoped by vendor_id.
type Service struct {
	db          *gorm.DB
	storageRoot string
	maxSize     int64
	now         func() time.Time
}

func NewService(db *gorm.DB, storageRoot string, maxSize int64) *Service {
	if maxSize <= 0 {
		maxSize = 20 << 20
	}
	return &Service{db: db, storageRoot: storageRoot, maxSize: maxSize, now: time.Now}
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

// VendorOf returns the vendor linked to a user, or ErrNotVendor.
func (s *Service) VendorOf(orgID, userID uuid.UUID) (uuid.UUID, error) {
	var row struct{ VendorID *uuid.UUID }
	err := s.db.Table("users").Select("vendor_id").
		Where("id = ? AND organization_id = ? AND deleted_at IS NULL AND is_active", userID, orgID).
		Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) || (err == nil && row.VendorID == nil) {
		return uuid.Nil, ErrNotVendor
	}
	if err != nil {
		return uuid.Nil, err
	}
	return *row.VendorID, nil
}

// contractScope is the minimal contract info needed to authorize access.
type contractScope struct {
	ID        uuid.UUID
	ProjectID uuid.UUID
	VendorID  uuid.UUID
}

// ownContract loads a contract that belongs to the vendor; anything else is ErrNotFound.
func (s *Service) ownContract(orgID, vendorID, contractID uuid.UUID) (*contractScope, error) {
	var c contractScope
	err := s.db.Table("contracts c").
		Select("c.id, c.project_id, c.vendor_id").
		Joins("JOIN projects p ON p.id = c.project_id AND p.deleted_at IS NULL AND p.organization_id = c.organization_id").
		Where("c.id = ? AND c.organization_id = ? AND c.vendor_id = ? AND c.deleted_at IS NULL", contractID, orgID, vendorID).
		Take(&c).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &c, err
}

// ── Contractor side ─────────────────────────────────────────────────────────

// MyContracts lists contracts held by the vendor.
func (s *Service) MyContracts(orgID, vendorID uuid.UUID) ([]ContractSummary, error) {
	var list []ContractSummary
	err := s.db.Table("contracts c").
		Select(`c.id, c.contract_number, c.title, c.status, c.contract_value, c.currency,
			c.start_date, c.end_date, p.id AS project_id, p.code AS project_code, p.name AS project_name`).
		Joins("JOIN projects p ON p.id = c.project_id AND p.deleted_at IS NULL AND p.organization_id = c.organization_id").
		Where("c.organization_id = ? AND c.vendor_id = ? AND c.deleted_at IS NULL", orgID, vendorID).
		Order("c.contract_number ASC").
		Scan(&list).Error
	if err != nil || len(list) == 0 {
		return list, err
	}
	for i := range list {
		if err := s.fillCounts(orgID, &list[i]); err != nil {
			return nil, err
		}
	}
	return list, nil
}

func (s *Service) fillCounts(orgID uuid.UUID, c *ContractSummary) error {
	base := func(table string) *gorm.DB {
		return s.db.Table(table).Where("organization_id = ? AND contract_id = ? AND deleted_at IS NULL", orgID, c.ID)
	}
	if err := base("vendor_progress_reports").Where("status = ?", StatusSubmitted).Count(&c.PendingReports).Error; err != nil {
		return err
	}
	if err := base("vendor_documents").Where("status = ?", StatusSubmitted).Count(&c.PendingDocuments).Error; err != nil {
		return err
	}
	var rr, rd int64
	if err := base("vendor_progress_reports").Where("status = ?", StatusRejected).Count(&rr).Error; err != nil {
		return err
	}
	if err := base("vendor_documents").Where("status = ?", StatusRejected).Count(&rd).Error; err != nil {
		return err
	}
	c.RejectedSubmissions = rr + rd
	var latest []float64
	if err := base("vendor_progress_reports").Where("status = ?", StatusVerified).
		Order("period_year DESC, period_month DESC").Limit(1).Pluck("physical_progress_pct", &latest).Error; err != nil {
		return err
	}
	if len(latest) == 1 {
		c.LatestVerifiedPct = &latest[0]
	}
	return nil
}

// MyContract returns one of the vendor's contracts.
func (s *Service) MyContract(orgID, vendorID, contractID uuid.UUID) (*ContractSummary, error) {
	list, err := s.MyContracts(orgID, vendorID)
	if err != nil {
		return nil, err
	}
	for i := range list {
		if list[i].ID == contractID {
			return &list[i], nil
		}
	}
	return nil, ErrNotFound
}

// ListReports lists progress reports of one of the vendor's contracts.
func (s *Service) ListReports(orgID, vendorID, contractID uuid.UUID) ([]ProgressReport, error) {
	if _, err := s.ownContract(orgID, vendorID, contractID); err != nil {
		return nil, err
	}
	var list []ProgressReport
	err := s.db.Where("organization_id = ? AND contract_id = ? AND vendor_id = ?", orgID, contractID, vendorID).
		Order("period_year DESC, period_month DESC, submitted_at DESC").Find(&list).Error
	return list, err
}

// CreateReport submits a monthly progress report for verification.
func (s *Service) CreateReport(orgID, vendorID, userID, contractID uuid.UUID, req CreateReportRequest) (*ProgressReport, error) {
	c, err := s.ownContract(orgID, vendorID, contractID)
	if err != nil {
		return nil, err
	}
	now := s.now()
	if req.PeriodYear > now.Year() || (req.PeriodYear == now.Year() && req.PeriodMonth > int(now.Month())) {
		return nil, invalid("period cannot be in the future")
	}
	r := ProgressReport{
		OrganizationID:      orgID,
		ProjectID:           c.ProjectID,
		ContractID:          c.ID,
		VendorID:            vendorID,
		PeriodYear:          req.PeriodYear,
		PeriodMonth:         req.PeriodMonth,
		PhysicalProgressPct: *req.PhysicalProgressPct,
		PlannedProgressPct:  req.PlannedProgressPct,
		Notes:               strings.TrimSpace(req.Notes),
		Status:              StatusSubmitted,
		SubmittedBy:         userID,
		SubmittedAt:         now,
	}
	if err := s.db.Create(&r).Error; err != nil {
		if isUniqueViolation(err) {
			return nil, ErrDuplicate
		}
		return nil, err
	}
	return &r, nil
}

// ListDocuments lists documents of one of the vendor's contracts.
func (s *Service) ListDocuments(orgID, vendorID, contractID uuid.UUID) ([]Document, error) {
	if _, err := s.ownContract(orgID, vendorID, contractID); err != nil {
		return nil, err
	}
	var list []Document
	err := s.db.Where("organization_id = ? AND contract_id = ? AND vendor_id = ?", orgID, contractID, vendorID).
		Order("submitted_at DESC").Find(&list).Error
	return list, err
}

// UploadInput is a contractor document upload.
type UploadInput struct {
	Filename    string
	Data        []byte
	Name        string
	Category    string
	PeriodYear  *int
	PeriodMonth *int
	Notes       string
}

// UploadDocument stores a contractor document for verification.
func (s *Service) UploadDocument(orgID, vendorID, userID, contractID uuid.UUID, in UploadInput) (*Document, error) {
	c, err := s.ownContract(orgID, vendorID, contractID)
	if err != nil {
		return nil, err
	}
	category := strings.ToUpper(strings.TrimSpace(in.Category))
	if !validDocCategories[category] {
		return nil, invalid("invalid document category")
	}
	if len(in.Data) == 0 {
		return nil, invalid("file is empty")
	}
	if int64(len(in.Data)) > s.maxSize {
		return nil, invalid(fmt.Sprintf("file exceeds maximum size of %d MB", s.maxSize>>20))
	}
	if (in.PeriodYear == nil) != (in.PeriodMonth == nil) {
		return nil, invalid("period_year and period_month must be provided together")
	}
	if in.PeriodMonth != nil && (*in.PeriodMonth < 1 || *in.PeriodMonth > 12) {
		return nil, invalid("period_month must be 1-12")
	}
	header := in.Data
	if len(header) > 512 {
		header = header[:512]
	}
	mime, err := project.ValidateUploadFile(header, in.Filename)
	if err != nil {
		return nil, invalid("file type is not allowed")
	}
	safe := project.SanitizeUploadFilename(in.Filename)
	name := strings.TrimSpace(in.Name)
	if name == "" {
		name = safe
	}
	if len(name) > 500 {
		return nil, invalid("name is too long")
	}
	id := uuid.New()
	key := fmt.Sprintf("vendor-documents/%s/%s/%s/%s", orgID, c.ID, id, safe)
	if err := project.SaveStoredFile(s.storageRoot, key, in.Data); err != nil {
		return nil, err
	}
	d := Document{
		ID: id, OrganizationID: orgID, ProjectID: c.ProjectID, ContractID: c.ID, VendorID: vendorID,
		Name: name, Category: category, PeriodYear: in.PeriodYear, PeriodMonth: in.PeriodMonth,
		FileURL: key, MimeType: mime, FileSize: int64(len(in.Data)), Notes: strings.TrimSpace(in.Notes),
		Status: StatusSubmitted, SubmittedBy: userID, SubmittedAt: s.now(),
	}
	if err := s.db.Create(&d).Error; err != nil {
		_ = project.DeleteStoredFile(s.storageRoot, key)
		return nil, err
	}
	return &d, nil
}

// DocumentFile resolves a document's file. vendorID restricts to the vendor's
// own documents; pass nil for reviewers (organization scope only).
func (s *Service) DocumentFile(orgID uuid.UUID, vendorID *uuid.UUID, docID uuid.UUID) (*Document, string, error) {
	q := s.db.Where("id = ? AND organization_id = ?", docID, orgID)
	if vendorID != nil {
		q = q.Where("vendor_id = ?", *vendorID)
	}
	var d Document
	if err := q.First(&d).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", ErrNotFound
		}
		return nil, "", err
	}
	path, _, err := project.OpenStoredFile(s.storageRoot, d.FileURL)
	if err != nil {
		return nil, "", ErrNotFound
	}
	return &d, path, nil
}

// ── Balai review side ───────────────────────────────────────────────────────

type contextRow struct {
	ContractID     uuid.UUID
	ContractNumber string
	ProjectCode    string
	ProjectName    string
	VendorName     string
	PPKUserID      *uuid.UUID
	PPKName        string
}

func (s *Service) contexts(orgID uuid.UUID, contractIDs []uuid.UUID) (map[uuid.UUID]Context, error) {
	out := map[uuid.UUID]Context{}
	if len(contractIDs) == 0 {
		return out, nil
	}
	var rows []contextRow
	err := s.db.Table("contracts c").
		Select("c.id AS contract_id, c.contract_number, p.code AS project_code, p.name AS project_name, v.name AS vendor_name, "+
			"c.ppk_user_id, TRIM(CONCAT(u.first_name, ' ', u.last_name)) AS ppk_name").
		Joins("JOIN projects p ON p.id = c.project_id").
		Joins("LEFT JOIN users u ON u.id = c.ppk_user_id").
		Joins("LEFT JOIN vendors v ON v.id = c.vendor_id").
		Where("c.organization_id = ? AND c.id IN ?", orgID, contractIDs).
		Scan(&rows).Error
	for _, r := range rows {
		out[r.ContractID] = Context{ContractNumber: r.ContractNumber, ProjectCode: r.ProjectCode, ProjectName: r.ProjectName, VendorName: r.VendorName, PPKUserID: r.PPKUserID, PPKName: r.PPKName}
	}
	return out, err
}

// ListSubmissions returns contractor submissions for Balai review.
func (s *Service) ListSubmissions(orgID uuid.UUID, f SubmissionFilter) (*Submissions, error) {
	status := strings.ToUpper(strings.TrimSpace(f.Status))
	scope := func(q *gorm.DB) *gorm.DB {
		q = q.Where("organization_id = ?", orgID)
		if status != "" {
			q = q.Where("status = ?", status)
		}
		if f.ProjectID != nil {
			q = q.Where("project_id = ?", *f.ProjectID)
		}
		return q
	}
	out := &Submissions{ProgressReports: []ProgressReport{}, Documents: []Document{}}
	if err := scope(s.db.Model(&ProgressReport{})).Order("submitted_at DESC").Limit(500).Find(&out.ProgressReports).Error; err != nil {
		return nil, err
	}
	if err := scope(s.db.Model(&Document{})).Order("submitted_at DESC").Limit(500).Find(&out.Documents).Error; err != nil {
		return nil, err
	}
	ids := map[uuid.UUID]bool{}
	for _, r := range out.ProgressReports {
		ids[r.ContractID] = true
	}
	for _, d := range out.Documents {
		ids[d.ContractID] = true
	}
	list := make([]uuid.UUID, 0, len(ids))
	for id := range ids {
		list = append(list, id)
	}
	ctx, err := s.contexts(orgID, list)
	if err != nil {
		return nil, err
	}
	for i := range out.ProgressReports {
		out.ProgressReports[i].Context = ctx[out.ProgressReports[i].ContractID]
	}
	for i := range out.Documents {
		out.Documents[i].Context = ctx[out.Documents[i].ContractID]
	}
	return out, nil
}

func normalizeDecision(req ReviewRequest) (string, string, error) {
	decision := strings.ToUpper(strings.TrimSpace(req.Decision))
	notes := strings.TrimSpace(req.Notes)
	if decision != StatusVerified && decision != StatusRejected {
		return "", "", invalid("decision must be VERIFIED or REJECTED")
	}
	if decision == StatusRejected && notes == "" {
		return "", "", invalid("a reason is required when rejecting")
	}
	return decision, notes, nil
}

// Reviewer identifies who reviews a submission.
type Reviewer struct {
	UserID uuid.UUID
}

// PrivilegedReviewRoles may review submissions regardless of the contract PPK.
var PrivilegedReviewRoles = []string{"SUPER_ADMIN", "ADMIN", "PMO"}

func (s *Service) isPrivileged(userID uuid.UUID) (bool, error) {
	var n int64
	err := s.db.Table("user_roles ur").
		Joins("JOIN roles r ON r.id = ur.role_id").
		Where("ur.user_id = ? AND r.code IN ?", userID, PrivilegedReviewRoles).
		Count(&n).Error
	return n > 0, err
}

// canReview enforces Perpres 16/2018 Ps.11: when a contract has a PPK, only
// that PPK (or a privileged reviewer) may verify its submissions. Contracts
// without a PPK keep the permission-only rule.
func (s *Service) canReview(orgID, contractID uuid.UUID, rv Reviewer) error {
	var ppk struct{ PPKUserID *uuid.UUID }
	if err := s.db.Table("contracts").Select("ppk_user_id").
		Where("id = ? AND organization_id = ?", contractID, orgID).Scan(&ppk).Error; err != nil {
		return err
	}
	if ppk.PPKUserID == nil || *ppk.PPKUserID == rv.UserID {
		return nil
	}
	priv, err := s.isPrivileged(rv.UserID)
	if err != nil {
		return err
	}
	if !priv {
		return ErrNotContractPPK
	}
	return nil
}

// ReviewReport verifies or rejects a SUBMITTED progress report.
func (s *Service) ReviewReport(orgID uuid.UUID, rv Reviewer, id uuid.UUID, req ReviewRequest) (*ProgressReport, error) {
	reviewerID := rv.UserID
	decision, notes, err := normalizeDecision(req)
	if err != nil {
		return nil, err
	}
	var r ProgressReport
	if err := s.db.Where("id = ? AND organization_id = ?", id, orgID).First(&r).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if r.Status != StatusSubmitted {
		return nil, ErrAlreadyDecided
	}
	if err := s.canReview(orgID, r.ContractID, rv); err != nil {
		return nil, err
	}
	now := s.now()
	r.Status, r.ReviewNotes, r.ReviewedBy, r.ReviewedAt = decision, notes, &reviewerID, &now
	// Guard against a concurrent review: only update while still SUBMITTED.
	res := s.db.Model(&ProgressReport{}).Where("id = ? AND status = ?", r.ID, StatusSubmitted).
		Updates(map[string]interface{}{"status": decision, "review_notes": notes, "reviewed_by": reviewerID, "reviewed_at": now, "updated_at": now})
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, ErrAlreadyDecided
	}
	return &r, nil
}

// ReviewDocument verifies or rejects a SUBMITTED document.
func (s *Service) ReviewDocument(orgID uuid.UUID, rv Reviewer, id uuid.UUID, req ReviewRequest) (*Document, error) {
	reviewerID := rv.UserID
	decision, notes, err := normalizeDecision(req)
	if err != nil {
		return nil, err
	}
	var d Document
	if err := s.db.Where("id = ? AND organization_id = ?", id, orgID).First(&d).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if d.Status != StatusSubmitted {
		return nil, ErrAlreadyDecided
	}
	if err := s.canReview(orgID, d.ContractID, rv); err != nil {
		return nil, err
	}
	now := s.now()
	d.Status, d.ReviewNotes, d.ReviewedBy, d.ReviewedAt = decision, notes, &reviewerID, &now
	res := s.db.Model(&Document{}).Where("id = ? AND status = ?", d.ID, StatusSubmitted).
		Updates(map[string]interface{}{"status": decision, "review_notes": notes, "reviewed_by": reviewerID, "reviewed_at": now, "updated_at": now})
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, ErrAlreadyDecided
	}
	return &d, nil
}

// ── Contractor accounts (admin) ─────────────────────────────────────────────

func (s *Service) vendorInOrg(orgID, vendorID uuid.UUID) error {
	var n int64
	if err := s.db.Table("vendors").Where("id = ? AND organization_id = ? AND deleted_at IS NULL", vendorID, orgID).Count(&n).Error; err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// ListAccounts lists contractor logins of a vendor.
func (s *Service) ListAccounts(orgID, vendorID uuid.UUID) ([]VendorAccount, error) {
	if err := s.vendorInOrg(orgID, vendorID); err != nil {
		return nil, err
	}
	var list []VendorAccount
	err := s.db.Table("users").
		Select("id, first_name, last_name, email, phone, is_active, must_change_pwd, last_login_at, created_at").
		Where("organization_id = ? AND vendor_id = ? AND deleted_at IS NULL", orgID, vendorID).
		Order("created_at ASC").Scan(&list).Error
	return list, err
}

// CreateAccount creates a contractor login with the VENDOR role and a one-time
// temporary password that must be changed at first login.
func (s *Service) CreateAccount(orgID, vendorID, actorID uuid.UUID, req CreateAccountRequest) (*CreatedAccount, error) {
	if err := s.vendorInOrg(orgID, vendorID); err != nil {
		return nil, err
	}
	var role rbac.Role
	if err := s.db.Where("organization_id = ? AND code = ?", orgID, constants.RoleVendor).First(&role).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("vendorportal: VENDOR role is not seeded")
		}
		return nil, err
	}
	password, err := temporaryPassword()
	if err != nil {
		return nil, err
	}
	hash, err := utils.HashPassword(password)
	if err != nil {
		return nil, err
	}
	actor := actorID
	vid := vendorID
	u := auth.User{
		ID:             uuid.New(),
		OrganizationID: orgID,
		VendorID:       &vid,
		FirstName:      strings.TrimSpace(req.FirstName),
		LastName:       strings.TrimSpace(req.LastName),
		Email:          strings.ToLower(strings.TrimSpace(req.Email)),
		Phone:          strings.TrimSpace(req.Phone),
		PasswordHash:   hash,
		IsActive:       true,
		MustChangePwd:  true,
		CreatedBy:      &actor,
	}
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&u).Error; err != nil {
			return err
		}
		return tx.Create(&rbac.UserRole{UserID: u.ID, RoleID: role.ID}).Error
	})
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrEmailTaken
		}
		return nil, err
	}
	return &CreatedAccount{
		Account: VendorAccount{
			ID: u.ID, FirstName: u.FirstName, LastName: u.LastName, Email: u.Email, Phone: u.Phone,
			IsActive: true, MustChangePwd: true, CreatedAt: u.CreatedAt,
		},
		TemporaryPassword: password,
	}, nil
}

// temporaryPassword returns a random 16-character password with upper, lower,
// digit, and symbol characters.
func temporaryPassword() (string, error) {
	sets := []string{"ABCDEFGHJKLMNPQRSTUVWXYZ", "abcdefghijkmnpqrstuvwxyz", "23456789", "!@#$%*?"}
	all := strings.Join(sets, "")
	out := make([]byte, 0, 16)
	pick := func(set string) error {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(set))))
		if err != nil {
			return err
		}
		out = append(out, set[n.Int64()])
		return nil
	}
	for _, set := range sets {
		if err := pick(set); err != nil {
			return "", err
		}
	}
	for len(out) < 16 {
		if err := pick(all); err != nil {
			return "", err
		}
	}
	// Shuffle so the guaranteed characters are not always first.
	for i := len(out) - 1; i > 0; i-- {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			return "", err
		}
		j := n.Int64()
		out[i], out[j] = out[j], out[i]
	}
	return string(out), nil
}
