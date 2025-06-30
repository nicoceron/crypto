package api

import (
	"errors"
	"net/http"
	"os"

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

func handleError(c *gin.Context, err error) {
	var appErr *apperrors.AppError

	if errors.As(err, &appErr) {
		println("🔴 AppError:", appErr.Error())
		c.JSON(appErr.HTTPStatus(), ErrorResponse{
			Error:   appErr.Message,
			Code:    appErr.Code,
			Details: appErr.Details,
		})
		return
	}

	println("🔴 Unknown Error:", err.Error())
	c.JSON(http.StatusInternalServerError, ErrorResponse{
		Error:   err.Error(),
		Code:    apperrors.ErrCodeInternal,
		Details: "Raw error returned for debugging purposes",
	})
}

// CORS middleware to handle cross-origin requests
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		allowedOrigins := []string{
			"http://localhost:5173",
			"http://localhost:3000",
			"http://127.0.0.1:5173",
			"http://127.0.0.1:3000",
		}

		if frontendURL := os.Getenv("FRONTEND_URL"); frontendURL != "" {
			allowedOrigins = append(allowedOrigins, frontendURL)
		}

		var allowedOrigin string
		for _, allowedOrigin = range allowedOrigins {
			if origin == allowedOrigin {
				break
			}
		}

		if origin != allowedOrigin {
			if os.Getenv("ENVIRONMENT") == "development" || os.Getenv("FRONTEND_URL") == "" {
				allowedOrigin = "*"
			} else {
				if frontendURL := os.Getenv("FRONTEND_URL"); frontendURL != "" {
					allowedOrigin = frontendURL
				} else {
					allowedOrigin = allowedOrigins[0]
				}
			}
		}

		c.Header("Access-Control-Allow-Origin", allowedOrigin)
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, HEAD")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept, Origin, X-Api-Key, X-Amz-Date, X-Amz-Security-Token")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Content-Type")
		c.Header("Access-Control-Allow-Credentials", "false")
		c.Header("Access-Control-Max-Age", "86400")

		c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
