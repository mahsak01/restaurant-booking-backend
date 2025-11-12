package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// BaseController base controller
type BaseController struct{}

// SuccessResponse returns success response
func (bc *BaseController) SuccessResponse(c *gin.Context, data interface{}, message string) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": message,
		"data":    data,
	})
}

// ErrorResponse returns error response
func (bc *BaseController) ErrorResponse(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, gin.H{
		"success": false,
		"message": message,
	})
}

// ValidationErrorResponse returns validation error response
func (bc *BaseController) ValidationErrorResponse(c *gin.Context, errors interface{}) {
	c.JSON(http.StatusBadRequest, gin.H{
		"success": false,
		"message": "Validation error",
		"errors":  errors,
	})
}

