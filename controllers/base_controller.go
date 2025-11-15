package controllers

import (
	"github.com/gofiber/fiber/v2"
)

// BaseController base controller
type BaseController struct{}

// SuccessResponse returns success response
func (bc *BaseController) SuccessResponse(c *fiber.Ctx, data interface{}, message string) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": message,
		"data":    data,
	})
}

// ErrorResponse returns error response
func (bc *BaseController) ErrorResponse(c *fiber.Ctx, statusCode int, message string) error {
	return c.Status(statusCode).JSON(fiber.Map{
		"success": false,
		"message": message,
	})
}

// ValidationErrorResponse returns validation error response
func (bc *BaseController) ValidationErrorResponse(c *fiber.Ctx, errors interface{}) error {
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		"success": false,
		"message": "Validation error",
		"errors":  errors,
	})
}
