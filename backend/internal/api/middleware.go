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
		c.JSON(appErr.HTTPStatus(), ErrorResponse{
			Error:   appErr.Message,
			Code:    appErr.Code,
			Details: appErr.Details,
		})
		return
	}

	// For unknown errors, return generic internal server error
	c.JSON(http.StatusInternalServerError, ErrorResponse{
		Error: "Internal server error",
		Code:  apperrors.ErrCodeInternal,
	})
}

// CORS middleware to handle cross-origin requests
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
} 