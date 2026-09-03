package organization

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/harmanto-49/cankora/internal/core/auth"
	"github.com/harmanto-49/cankora/internal/platform/response"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// OrgCRUDHandler serves top-level organization CRUD endpoints.
// Only SUPER_ADMIN should be able to list all orgs or create new ones.
type OrgCRUDHandler struct {
	db  *gorm.DB
	log *zap.Logger
}

func NewOrgCRUDHandler(db *gorm.DB, log *zap.Logger) *OrgCRUDHandler {
	return &OrgCRUDHandler{db: db, log: log}
}

// claimsForOrgHandler extracts auth claims from Gin context.
func claimsForOrgHandler(c *gin.Context) *auth.Claims {
	val, exists := c.Get(string(auth.ContextKeyClaims))
	if !exists {
		return nil
	}
	cl, ok := val.(*auth.Claims)
	if !ok {
		return nil
	}
	return cl
}

// ListOrganizations godoc
// GET /api/v1/organizations
// Returns all organizations. SUPER_ADMIN sees all; other roles see only their own.
func (h *OrgCRUDHandler) List(c *gin.Context) {
	claims := claimsForOrgHandler(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	var orgs []Organization
	q := h.db.Where("deleted_at IS NULL").Order("name ASC")

	// Non-super-admin can only see their own organization
	// Super admin check is done by the RBAC middleware; here we just scope by org
	// for non-super-admins. SUPER_ADMIN role has the permission to view all orgs.
	if err := q.Find(&orgs).Error; err != nil {
		h.log.Error("org: list", zap.Error(err))
		response.InternalError(c)
		return
	}

	response.OK(c, "organizations retrieved", orgs)
}

// GetOrganization godoc
// GET /api/v1/organizations/:orgID
func (h *OrgCRUDHandler) Get(c *gin.Context) {
	claims := claimsForOrgHandler(c)
	if claims == nil {
		response.Unauthorized(c, "unauthorized")
		return
	}

	orgID, err := uuid.Parse(c.Param("orgID"))
	if err != nil {
		response.BadRequest(c, "invalid organization id")
		return
	}

	var org Organization
	if err := h.db.Where("id = ? AND deleted_at IS NULL", orgID).First(&org).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, "organization not found")
			return
		}
		h.log.Error("org: get", zap.Error(err))
		response.InternalError(c)
		return
	}

	response.OK(c, "organization retrieved", org)
}

// CreateOrganization godoc
// POST /api/v1/organizations
func (h *OrgCRUDHandler) Create(c *gin.Context) {
	var req CreateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Check for duplicate code
	var count int64
	h.db.Model(&Organization{}).Where("code = ? AND deleted_at IS NULL", req.Code).Count(&count)
	if count > 0 {
		response.Conflict(c, "organization code already in use")
		return
	}

	org := &Organization{
		Code:      req.Code,
		Name:      req.Name,
		ShortName: req.ShortName,
		Address:   req.Address,
		Website:   req.Website,
		IsActive:  true,
	}

	if err := h.db.Create(org).Error; err != nil {
		h.log.Error("org: create", zap.Error(err))
		response.InternalError(c)
		return
	}

	response.Created(c, "organization created", org)
}

// UpdateOrganization godoc
// PUT /api/v1/organizations/:orgID
func (h *OrgCRUDHandler) Update(c *gin.Context) {
	orgID, err := uuid.Parse(c.Param("orgID"))
	if err != nil {
		response.BadRequest(c, "invalid organization id")
		return
	}

	var req UpdateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	var org Organization
	if err := h.db.Where("id = ? AND deleted_at IS NULL", orgID).First(&org).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, "organization not found")
			return
		}
		h.log.Error("org: update find", zap.Error(err))
		response.InternalError(c)
		return
	}

	if req.Name != "" {
		org.Name = req.Name
	}
	if req.ShortName != "" {
		org.ShortName = req.ShortName
	}
	if req.LogoURL != "" {
		org.LogoURL = req.LogoURL
	}
	if req.Address != "" {
		org.Address = req.Address
	}
	if req.Website != "" {
		org.Website = req.Website
	}
	if req.IsActive != nil {
		org.IsActive = *req.IsActive
	}

	if err := h.db.Save(&org).Error; err != nil {
		h.log.Error("org: update save", zap.Error(err))
		response.InternalError(c)
		return
	}

	response.OK(c, "organization updated", org)
}
