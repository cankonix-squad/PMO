package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/harmanto-49/cankora/internal/shared/types"
)

// Envelope is the standard API response structure for all endpoints.
type Envelope struct {
	Success bool         `json:"success"`
	Message string       `json:"message,omitempty"`
	Data    interface{}  `json:"data,omitempty"`
	Meta    interface{}  `json:"meta,omitempty"`
	Errors  []FieldError `json:"errors,omitempty"`
}

// FieldError represents a validation error for a specific field.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// OK sends a 200 response with data.
func OK(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Envelope{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// OKPaginated sends a 200 response with paginated data.
func OKPaginated(c *gin.Context, message string, data interface{}, meta types.PaginationMeta) {
	c.JSON(http.StatusOK, Envelope{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}

// Created sends a 201 response.
func Created(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusCreated, Envelope{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// NoContent sends a 204 response (no body).
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// BadRequest sends a 400 response.
func BadRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, Envelope{
		Success: false,
		Message: message,
	})
}

// ValidationError sends a 422 response with field errors.
func ValidationError(c *gin.Context, errors []FieldError) {
	c.JSON(http.StatusUnprocessableEntity, Envelope{
		Success: false,
		Message: "Validation failed",
		Errors:  errors,
	})
}

// Unauthorized sends a 401 response.
func Unauthorized(c *gin.Context, message string) {
	if message == "" {
		message = "Unauthorized"
	}
	c.JSON(http.StatusUnauthorized, Envelope{
		Success: false,
		Message: message,
	})
}

// Forbidden sends a 403 response.
func Forbidden(c *gin.Context, message string) {
	if message == "" {
		message = "You do not have permission to perform this action"
	}
	c.JSON(http.StatusForbidden, Envelope{
		Success: false,
		Message: message,
	})
}

// NotFound sends a 404 response.
func NotFound(c *gin.Context, message string) {
	if message == "" {
		message = "Resource not found"
	}
	c.JSON(http.StatusNotFound, Envelope{
		Success: false,
		Message: message,
	})
}

// Conflict sends a 409 response.
func Conflict(c *gin.Context, message string) {
	c.JSON(http.StatusConflict, Envelope{
		Success: false,
		Message: message,
	})
}

// InternalError sends a 500 response without exposing internal details.
func InternalError(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, Envelope{
		Success: false,
		Message: "An internal error occurred. Please try again later.",
	})
}
