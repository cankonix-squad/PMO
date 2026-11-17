package kpbu

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/harmanto-49/cankora/internal/project"
	"github.com/harmanto-49/cankora/internal/shared/types"
	"gorm.io/gorm"
)

// financialCloseStatus fills the computed financial-close deadline fields.
func financialCloseStatus(p *Project, now time.Time) {
	p.FinancialCloseStatus = "NONE"
	p.FinancialCloseDeadline = nil
	p.FinancialCloseDaysLeft = nil
	if p.FinancialCloseActual != nil && !p.FinancialCloseActual.IsZero() {
		p.FinancialCloseStatus = "ACHIEVED"
	}
	if p.AgreementSignedDate == nil || p.AgreementSignedDate.IsZero() {
		return
	}
	dl := types.FlexTime{Time: today(p.AgreementSignedDate.Time).AddDate(0, FinancialCloseMonths, 0)}
	p.FinancialCloseDeadline = &dl
	if p.FinancialCloseStatus == "ACHIEVED" {
		return
	}
	days := int(math.Round(dl.Sub(today(now)).Hours() / 24))
	p.FinancialCloseDaysLeft = &days
	switch {
	case days < 0:
		p.FinancialCloseStatus = "OVERDUE"
	case days <= FinancialCloseWarnDays:
		p.FinancialCloseStatus = "WARNING"
	default:
		p.FinancialCloseStatus = "ON_TRACK"
	}
}

func appliesTo(r DocumentRequirement, initiative string) bool {
	return r.AppliesTo == AppliesAll || r.AppliesTo == initiative
}

// ── Requirement master ──────────────────────────────────────────────────────

func (s *Service) ListRequirements(orgID uuid.UUID, includeInactive bool) ([]DocumentRequirement, error) {
	q := s.db.Where("organization_id = ?", orgID)
	if !includeInactive {
		q = q.Where("is_active = TRUE")
	}
	var list []DocumentRequirement
	err := q.Order("sort_order ASC, code ASC").Find(&list).Error
	return list, err
}

func normalizeRequirement(req *RequirementRequest) error {
	req.Code = strings.ToUpper(strings.TrimSpace(req.Code))
	req.Category = strings.ToUpper(strings.TrimSpace(req.Category))
	req.Name = strings.TrimSpace(req.Name)
	req.AppliesTo = strings.ToUpper(strings.TrimSpace(req.AppliesTo))
	req.SourceStatus = strings.ToUpper(strings.TrimSpace(req.SourceStatus))
	if req.AppliesTo == "" {
		req.AppliesTo = AppliesAll
	}
	if req.SourceStatus == "" {
		req.SourceStatus = SourceUnverified
	}
	if req.Code == "" || req.Name == "" {
		return invalid("code and name are required")
	}
	if !validCategories[req.Category] {
		return invalid("invalid category")
	}
	if !validAppliesTo[req.AppliesTo] {
		return invalid("invalid applies_to")
	}
	if !validSources[req.SourceStatus] {
		return invalid("invalid source_status")
	}
	if req.RequiredBeforeStage != nil {
		st := strings.ToUpper(strings.TrimSpace(*req.RequiredBeforeStage))
		if st == "" {
			req.RequiredBeforeStage = nil
		} else if !isMainStage(st) && st != StageClosed {
			return invalid("invalid required_before_stage")
		} else {
			req.RequiredBeforeStage = &st
		}
	}
	return nil
}

func applyRequirement(r *DocumentRequirement, req RequirementRequest) {
	r.Code = req.Code
	r.Category = req.Category
	r.Name = req.Name
	r.Description = strings.TrimSpace(req.Description)
	r.AppliesTo = req.AppliesTo
	r.RequiredBeforeStage = req.RequiredBeforeStage
	r.IsMandatory = req.IsMandatory
	r.LegalBasis = strings.TrimSpace(req.LegalBasis)
	r.SourceStatus = req.SourceStatus
	r.SortOrder = req.SortOrder
	if req.IsActive != nil {
		r.IsActive = *req.IsActive
	}
}

func (s *Service) CreateRequirement(orgID uuid.UUID, req RequirementRequest) (*DocumentRequirement, error) {
	if err := normalizeRequirement(&req); err != nil {
		return nil, err
	}
	r := DocumentRequirement{OrganizationID: orgID, IsActive: true}
	applyRequirement(&r, req)
	if err := s.db.Create(&r).Error; err != nil {
		if isUniqueViolation(err) {
			return nil, ErrCodeTaken
		}
		return nil, err
	}
	return &r, nil
}

func (s *Service) findRequirement(orgID, id uuid.UUID) (*DocumentRequirement, error) {
	var r DocumentRequirement
	err := s.db.Where("id = ? AND organization_id = ?", id, orgID).First(&r).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &r, err
}

func (s *Service) UpdateRequirement(orgID, id uuid.UUID, req RequirementRequest) (*DocumentRequirement, error) {
	if err := normalizeRequirement(&req); err != nil {
		return nil, err
	}
	r, err := s.findRequirement(orgID, id)
	if err != nil {
		return nil, err
	}
	applyRequirement(r, req)
	if err := s.db.Save(r).Error; err != nil {
		if isUniqueViolation(err) {
			return nil, ErrCodeTaken
		}
		return nil, err
	}
	return r, nil
}

func (s *Service) DeleteRequirement(orgID, id uuid.UUID) (*DocumentRequirement, error) {
	r, err := s.findRequirement(orgID, id)
	if err != nil {
		return nil, err
	}
	return r, s.db.Delete(r).Error
}

// ── Checklist and stage gate ────────────────────────────────────────────────

func (s *Service) projectDocuments(orgID, projectID uuid.UUID) ([]Document, error) {
	var docs []Document
	err := s.db.Where("organization_id = ? AND kpbu_project_id = ?", orgID, projectID).
		Order("uploaded_at DESC").Find(&docs).Error
	return docs, err
}

// itemStatus: APPROVED if any approved document, else the latest document's status.
func itemStatus(docs []Document) string {
	if len(docs) == 0 {
		return DocMissing
	}
	for _, d := range docs {
		if d.Status == DocApproved {
			return DocApproved
		}
	}
	return docs[0].Status
}

// missingForStage lists mandatory checklist codes without an approved document
// that gate entry into stage `to`.
func (s *Service) missingForStage(orgID uuid.UUID, p *Project, to string) ([]string, error) {
	var reqs []DocumentRequirement
	if err := s.db.Where("organization_id = ? AND is_active = TRUE AND is_mandatory = TRUE AND required_before_stage = ?", orgID, to).
		Order("sort_order ASC, code ASC").Find(&reqs).Error; err != nil {
		return nil, err
	}
	if len(reqs) == 0 {
		return nil, nil
	}
	var approved []uuid.UUID
	if err := s.db.Model(&Document{}).
		Where("organization_id = ? AND kpbu_project_id = ? AND status = ? AND requirement_id IS NOT NULL", orgID, p.ID, DocApproved).
		Distinct().Pluck("requirement_id", &approved).Error; err != nil {
		return nil, err
	}
	ok := map[uuid.UUID]bool{}
	for _, id := range approved {
		ok[id] = true
	}
	missing := []string{}
	for _, r := range reqs {
		if appliesTo(r, p.Initiative) && !ok[r.ID] {
			missing = append(missing, r.Code)
		}
	}
	return missing, nil
}

// Checklist returns the requirements applicable to a project with their documents.
func (s *Service) Checklist(orgID, projectID uuid.UUID) (*Checklist, error) {
	p, err := s.findProject(orgID, projectID)
	if err != nil {
		return nil, err
	}
	reqs, err := s.ListRequirements(orgID, false)
	if err != nil {
		return nil, err
	}
	docs, err := s.projectDocuments(orgID, projectID)
	if err != nil {
		return nil, err
	}
	byReq := map[uuid.UUID][]Document{}
	other := []Document{}
	for _, d := range docs {
		if d.RequirementID == nil {
			other = append(other, d)
			continue
		}
		byReq[*d.RequirementID] = append(byReq[*d.RequirementID], d)
	}
	out := &Checklist{Items: []ChecklistItem{}, Other: other, Initiative: p.Initiative}
	if next, ok := nextMainStage[p.Stage]; ok {
		out.Summary.NextStage = next
	}
	out.Summary.BlockingNextStage = []string{}
	known := map[uuid.UUID]bool{}
	for _, r := range reqs {
		known[r.ID] = true
		if !appliesTo(r, p.Initiative) {
			continue
		}
		ds := byReq[r.ID]
		if ds == nil {
			ds = []Document{}
		}
		st := itemStatus(ds)
		out.Items = append(out.Items, ChecklistItem{Requirement: r, Status: st, Documents: ds})
		if r.IsMandatory {
			out.Summary.MandatoryTotal++
			if st == DocApproved {
				out.Summary.MandatoryApproved++
			}
			if st != DocApproved && r.RequiredBeforeStage != nil && *r.RequiredBeforeStage == out.Summary.NextStage {
				out.Summary.BlockingNextStage = append(out.Summary.BlockingNextStage, r.Code)
			}
		}
		if st == DocSubmitted {
			out.Summary.PendingReview++
		}
	}
	// Documents of deactivated/deleted requirements stay visible.
	for id, ds := range byReq {
		if !known[id] {
			out.Other = append(out.Other, ds...)
		}
	}
	return out, nil
}

// ── Documents ───────────────────────────────────────────────────────────────

func (s *Service) UploadDocument(orgID, userID, projectID uuid.UUID, in DocumentUpload) (*Document, error) {
	if _, err := s.findProject(orgID, projectID); err != nil {
		return nil, err
	}
	if in.RequirementID != nil {
		if _, err := s.findRequirement(orgID, *in.RequirementID); err != nil {
			if errors.Is(err, ErrNotFound) {
				return nil, invalid("requirement_id not found")
			}
			return nil, err
		}
	}
	if len(in.Data) == 0 {
		return nil, invalid("file is empty")
	}
	if int64(len(in.Data)) > s.maxSize {
		return nil, invalid(fmt.Sprintf("file exceeds maximum size of %d MB", s.maxSize>>20))
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
	if len(name) > 500 || len(in.DocumentNumber) > 200 {
		return nil, invalid("name or document_number is too long")
	}
	id := uuid.New()
	key := fmt.Sprintf("kpbu-documents/%s/%s/%s/%s", orgID, projectID, id, safe)
	if err := project.SaveStoredFile(s.storageRoot, key, in.Data); err != nil {
		return nil, err
	}
	d := Document{
		ID: id, OrganizationID: orgID, KPBUProjectID: projectID, RequirementID: in.RequirementID,
		Name: name, FileURL: key, MimeType: mime, FileSize: int64(len(in.Data)),
		DocumentNumber: strings.TrimSpace(in.DocumentNumber), DocumentDate: in.DocumentDate,
		Notes: strings.TrimSpace(in.Notes), Status: DocSubmitted, UploadedBy: userID, UploadedAt: s.now(),
	}
	if err := s.db.Create(&d).Error; err != nil {
		_ = project.DeleteStoredFile(s.storageRoot, key)
		return nil, err
	}
	return &d, nil
}

func (s *Service) findDocument(orgID, projectID, docID uuid.UUID) (*Document, error) {
	var d Document
	err := s.db.Where("id = ? AND organization_id = ? AND kpbu_project_id = ?", docID, orgID, projectID).First(&d).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &d, err
}

// ReviewDocument approves or rejects a submitted document.
func (s *Service) ReviewDocument(orgID, reviewerID, projectID, docID uuid.UUID, req ReviewRequest) (*Document, error) {
	decision := strings.ToUpper(strings.TrimSpace(req.Decision))
	notes := strings.TrimSpace(req.Notes)
	switch decision {
	case "APPROVE", DocApproved:
		decision = DocApproved
	case "REJECT", DocRejected:
		decision = DocRejected
		if notes == "" {
			return nil, invalid("notes are required when rejecting")
		}
	default:
		return nil, invalid("decision must be APPROVE or REJECT")
	}
	d, err := s.findDocument(orgID, projectID, docID)
	if err != nil {
		return nil, err
	}
	if d.Status != DocSubmitted {
		return nil, invalid("only SUBMITTED documents can be reviewed")
	}
	now := s.now()
	d.Status = decision
	d.ReviewedBy = &reviewerID
	d.ReviewedAt = &now
	d.ReviewNotes = notes
	return d, s.db.Save(d).Error
}

// DocumentFile resolves the stored file path of a document.
func (s *Service) DocumentFile(orgID, projectID, docID uuid.UUID) (*Document, string, error) {
	d, err := s.findDocument(orgID, projectID, docID)
	if err != nil {
		return nil, "", err
	}
	path, _, err := project.OpenStoredFile(s.storageRoot, d.FileURL)
	if err != nil {
		return nil, "", ErrNotFound
	}
	return d, path, nil
}

// DeleteDocument soft-deletes a document that is not approved. The file is kept
// so the audit trail can still reference it.
func (s *Service) DeleteDocument(orgID, projectID, docID uuid.UUID) (*Document, error) {
	d, err := s.findDocument(orgID, projectID, docID)
	if err != nil {
		return nil, err
	}
	if d.Status == DocApproved {
		return nil, invalid("approved documents cannot be deleted")
	}
	return d, s.db.Delete(d).Error
}

// ── Guarantees ──────────────────────────────────────────────────────────────

func (s *Service) enrichGuarantees(p *Project, list []Guarantee) {
	now := today(s.now())
	for i := range list {
		g := &list[i]
		if p.InvestmentValue > 0 && g.Currency == p.Currency {
			pct := math.Round(g.Amount/p.InvestmentValue*10000) / 100
			g.PctOfInvestment = &pct
			g.ExceedsLimit = g.Type == "PERFORMANCE" && pct > PerformanceGuaranteeMaxPct
		}
		if g.ExpiryDate != nil && !g.ExpiryDate.IsZero() {
			d := int(math.Round(today(g.ExpiryDate.Time).Sub(now).Hours() / 24))
			g.DaysToExpiry = &d
		}
	}
}

func (s *Service) ListGuarantees(orgID, projectID uuid.UUID) ([]Guarantee, error) {
	p, err := s.findProject(orgID, projectID)
	if err != nil {
		return nil, err
	}
	var list []Guarantee
	if err := s.db.Where("organization_id = ? AND kpbu_project_id = ?", orgID, projectID).
		Order("created_at ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	s.enrichGuarantees(p, list)
	return list, nil
}

func normalizeGuarantee(req *GuaranteeRequest) error {
	req.Type = strings.ToUpper(strings.TrimSpace(req.Type))
	req.Status = strings.ToUpper(strings.TrimSpace(req.Status))
	req.Currency = strings.ToUpper(strings.TrimSpace(req.Currency))
	if req.Status == "" {
		req.Status = "ACTIVE"
	}
	if req.Currency == "" {
		req.Currency = "IDR"
	}
	if !validGuarTypes[req.Type] {
		return invalid("invalid guarantee type")
	}
	if !validGuarStatus[req.Status] {
		return invalid("invalid guarantee status")
	}
	if req.IssuedDate != nil && req.ExpiryDate != nil && !req.IssuedDate.IsZero() && !req.ExpiryDate.IsZero() &&
		req.ExpiryDate.Before(req.IssuedDate.Time) {
		return invalid("expiry_date must not be before issued_date")
	}
	return nil
}

func applyGuarantee(g *Guarantee, req GuaranteeRequest) {
	g.Type = req.Type
	g.Issuer = strings.TrimSpace(req.Issuer)
	g.GuaranteeNumber = strings.TrimSpace(req.GuaranteeNumber)
	g.Amount = req.Amount
	g.Currency = req.Currency
	g.IssuedDate = req.IssuedDate
	g.ExpiryDate = req.ExpiryDate
	g.Status = req.Status
	g.Notes = strings.TrimSpace(req.Notes)
}

func (s *Service) CreateGuarantee(orgID, projectID uuid.UUID, req GuaranteeRequest) (*Guarantee, error) {
	if err := normalizeGuarantee(&req); err != nil {
		return nil, err
	}
	p, err := s.findProject(orgID, projectID)
	if err != nil {
		return nil, err
	}
	g := Guarantee{OrganizationID: orgID, KPBUProjectID: projectID}
	applyGuarantee(&g, req)
	if err := s.db.Create(&g).Error; err != nil {
		return nil, err
	}
	one := []Guarantee{g}
	s.enrichGuarantees(p, one)
	return &one[0], nil
}

func (s *Service) findGuarantee(orgID, projectID, id uuid.UUID) (*Guarantee, error) {
	var g Guarantee
	err := s.db.Where("id = ? AND organization_id = ? AND kpbu_project_id = ?", id, orgID, projectID).First(&g).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &g, err
}

func (s *Service) UpdateGuarantee(orgID, projectID, id uuid.UUID, req GuaranteeRequest) (*Guarantee, error) {
	if err := normalizeGuarantee(&req); err != nil {
		return nil, err
	}
	p, err := s.findProject(orgID, projectID)
	if err != nil {
		return nil, err
	}
	g, err := s.findGuarantee(orgID, projectID, id)
	if err != nil {
		return nil, err
	}
	applyGuarantee(g, req)
	if err := s.db.Save(g).Error; err != nil {
		return nil, err
	}
	one := []Guarantee{*g}
	s.enrichGuarantees(p, one)
	return &one[0], nil
}

func (s *Service) DeleteGuarantee(orgID, projectID, id uuid.UUID) (*Guarantee, error) {
	g, err := s.findGuarantee(orgID, projectID, id)
	if err != nil {
		return nil, err
	}
	return g, s.db.Delete(g).Error
}
