package auth

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/harmanto-49/cankora/internal/platform/response"
)

// Handler wires the auth service to HTTP routes.
type Handler struct {
	service      *Service
	tokenService *TokenService
}

// NewHandler creates a new auth Handler.
func NewHandler(svc *Service, tokenSvc *TokenService) *Handler {
	return &Handler{service: svc, tokenService: tokenSvc}
}

// RegisterRoutes mounts the auth routes onto a router group.
// The group should be at /api/v1/auth.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/login", h.Login)
	rg.POST("/refresh", h.Refresh)
	rg.POST("/forgot-password", h.ForgotPassword)
	rg.POST("/reset-password", h.ResetPassword)
}

// RegisterProtectedRoutes mounts routes that require a valid JWT.
func (h *Handler) RegisterProtectedRoutes(rg *gin.RouterGroup) {
	rg.POST("/logout", h.Logout)
	rg.POST("/logout-all", h.LogoutAll)
	rg.POST("/change-password", h.ChangePassword)
	rg.GET("/me", h.Me)
}

// Login godoc
// POST /api/v1/auth/login
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	pair, err := h.service.Login(c.Request.Context(), &req,
		c.Request.UserAgent(), c.ClientIP())
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidCredentials):
			response.Unauthorized(c, "Invalid email or password")
		case errors.Is(err, ErrAccountLocked):
			c.JSON(http.StatusTooManyRequests, response.Envelope{
				Success: false,
				Message: "Account locked due to too many failed attempts. Try again in 15 minutes.",
			})
		case errors.Is(err, ErrAccountInactive):
			response.Forbidden(c, "Account is inactive. Contact your administrator.")
		default:
			response.InternalError(c)
		}
		return
	}

	response.OK(c, "Login successful", pair)
}

// Refresh godoc
// POST /api/v1/auth/refresh
func (h *Handler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	pair, err := h.service.Refresh(c.Request.Context(), req.RefreshToken,
		c.Request.UserAgent(), c.ClientIP())
	if err != nil {
		response.Unauthorized(c, "Invalid or expired refresh token")
		return
	}

	response.OK(c, "Token refreshed", pair)
}

// ForgotPassword godoc
// POST /api/v1/auth/forgot-password
func (h *Handler) ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Always return 200 to prevent email enumeration
	_, _ = h.service.ForgotPassword(c.Request.Context(), req.Email)
	response.OK(c, "If the email exists, a reset link has been sent.", nil)
}

// ResetPassword godoc
// POST /api/v1/auth/reset-password
func (h *Handler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.service.ResetPassword(c.Request.Context(), &req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.OK(c, "Password has been reset successfully", nil)
}

// Logout godoc
// POST /api/v1/auth/logout  [requires auth]
func (h *Handler) Logout(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "")
		return
	}

	if err := h.service.Logout(c.Request.Context(), claims.JTI); err != nil {
		response.InternalError(c)
		return
	}

	response.OK(c, "Logged out successfully", nil)
}

// LogoutAll godoc
// POST /api/v1/auth/logout-all  [requires auth]
func (h *Handler) LogoutAll(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "")
		return
	}

	if err := h.service.LogoutAll(c.Request.Context(), claims.UserID); err != nil {
		response.InternalError(c)
		return
	}

	response.OK(c, "All sessions terminated", nil)
}

// ChangePassword godoc
// POST /api/v1/auth/change-password  [requires auth]
func (h *Handler) ChangePassword(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "")
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.service.ChangePassword(c.Request.Context(), claims.UserID, &req); err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			response.BadRequest(c, "Current password is incorrect")
			return
		}
		response.InternalError(c)
		return
	}

	response.OK(c, "Password changed successfully", nil)
}

// Me godoc
// GET /api/v1/auth/me  [requires auth]
func (h *Handler) Me(c *gin.Context) {
	claims := claimsFromGin(c)
	if claims == nil {
		response.Unauthorized(c, "")
		return
	}

	user, err := h.service.repo.FindUserByID(c.Request.Context(), claims.UserID)
	if err != nil {
		response.NotFound(c, "User not found")
		return
	}

	roles, err := h.service.fetchUserRoles(c.Request.Context(), user.ID)
	if err != nil {
		response.InternalError(c)
		return
	}

	info := UserInfo{
		ID:             user.ID,
		OrganizationID: user.OrganizationID,
		OrgUnitID:      user.OrgUnitID,
		FirstName:      user.FirstName,
		LastName:       user.LastName,
		Email:          user.Email,
		JobTitle:       user.JobTitle,
		AvatarURL:      user.AvatarURL,
		MustChangePwd:  user.MustChangePwd,
		Roles:          roles,
	}

	response.OK(c, "OK", info)
}

// claimsFromGin is a local helper to extract Claims from context.
func claimsFromGin(c *gin.Context) *Claims {
	v, exists := c.Get(string(ContextKeyClaims))
	if !exists {
		return nil
	}
	claims, _ := v.(*Claims)
	return claims
}
