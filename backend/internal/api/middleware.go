package api

import (
	"errors"
	"net/http"

	apperrors "stock-analyzer/pkg/errors"

	"github.com/gin-gonic/gin"
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code"`
	Details string `json:"details,omitempty"`
}

// ErrorHandler middleware handles application errors and converts them to HTTP responses
func ErrorHandler() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(error); ok {
			handleError(c, err)
		} else {
			// Unknown panic
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error: "Internal server error",
				Code:  apperrors.ErrCodeInternal,
			})
		}
		c.Abort()
	})
}

// HandleError is a helper function to handle errors in handlers
func HandleError(c *gin.Context, err error) {
	handleError(c, err)
}

// handleError processes the error and sends appropriate HTTP response
func handleError(c *gin.Context, err error) {
	var appErr *apperrors.AppError

	// Check if it's our custom error type
	if errors.As(err, &appErr) {
		// Log the full error for debugging
		println("🔴 AppError:", appErr.Error())
		c.JSON(appErr.HTTPStatus(), ErrorResponse{
			Error:   appErr.Message,
			Code:    appErr.Code,
			Details: appErr.Details,
		})
		return
	}

	// For unknown errors, log the full error and return detailed message for debugging
	println("🔴 Unknown Error:", err.Error())
	c.JSON(http.StatusInternalServerError, ErrorResponse{
		Error:   err.Error(), // Return the actual error message for debugging
		Code:    apperrors.ErrCodeInternal,
		Details: "Raw error returned for debugging purposes",
	})
}

// CORS middleware to handle cross-origin requests
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set CORS headers for all requests
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, HEAD")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept, Origin, X-Api-Key, X-Amz-Date, X-Amz-Security-Token")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Content-Type")
		c.Header("Access-Control-Allow-Credentials", "false")
		c.Header("Access-Control-Max-Age", "86400")

		// Add cache control headers to prevent browser caching issues
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")

		// Handle preflight OPTIONS requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
