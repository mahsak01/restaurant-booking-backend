// @title Restaurant Booking API
// @version 1.0
// @description RESTful API for restaurant table booking system
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@restaurant.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

package main

import (
	_ "embed"
	"log"
	"os"

	"restaurant-booking-backend/config"
	"restaurant-booking-backend/models"
	"restaurant-booking-backend/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/swagger"
)

//go:embed docs/swagger.json
var swaggerJSON []byte

func main() {
	// Load configuration
	config.LoadConfig()

	// Connect to database
	config.InitDB()

	// Auto migrate database
	if err := config.DB.AutoMigrate(
		&models.User{},
		&models.Table{},
		&models.MenuItem{},
		&models.Reservation{},
		&models.Notification{},
		&models.Category{},
	); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}
	log.Println("Database migration completed")

	// Create Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
			})
		},
	})

	// Setup Logger middleware - logs all requests
	app.Use(logger.New(logger.Config{
		Format:     "${time} | ${status} | ${latency} | ${method} | ${path} | ${ip} | ${error}\n",
		TimeFormat: "2006-01-02 15:04:05",
		TimeZone:   "Asia/Tehran",
		Output:     os.Stdout,
	}))

	// Setup CORS middleware
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3000",
		AllowCredentials: true,
		AllowHeaders:     "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With",
		AllowMethods:     "POST, OPTIONS, GET, PUT, DELETE, PATCH",
	}))

	// Serve Swagger JSON
	app.Get("/swagger/doc.json", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "application/json")
		return c.Send(swaggerJSON)
	})

	// Swagger UI route
	app.Get("/swagger/*", swagger.New(swagger.Config{
		URL:         "/swagger/doc.json",
		DeepLinking: true,
	}))

	// Setup routes
	routes.SetupRoutes(app)

	// Read port from environment variable or use default port
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start server
	log.Printf("Server is running on port %s...", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
