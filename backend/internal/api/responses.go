package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
)

// NewSuccessResponse creates a successful API Gateway response.
func NewSuccessResponse(statusCode int, body interface{}) events.APIGatewayProxyResponse {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		log.Printf("Failed to marshal success response: %v", err)
		return NewErrorResponse(http.StatusInternalServerError, "Failed to create response")
	}

	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Body:       string(jsonBody),
		Headers:    map[string]string{"Content-Type": "application/json"},
	}
}

// NewErrorResponse creates a failure API Gateway response.
func NewErrorResponse(statusCode int, message string) events.APIGatewayProxyResponse {
	errorBody := ErrorResponse{
		Error: message,
	}
	jsonBody, err := json.Marshal(errorBody)
	if err != nil {
		// This is a fallback and should rarely happen.
		log.Printf("Failed to marshal error response: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       `{"error":"failed to create error response"}`,
			Headers:    map[string]string{"Content-Type": "application/json"},
		}
	}

	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Body:       string(jsonBody),
		Headers:    map[string]string{"Content-Type": "application/json"},
	}
} 