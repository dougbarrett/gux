package main

import (
	"fmt"
	"log"
	"os"

	"test-admin/dto"
	"test-admin/models"
	"test-admin/pages"

	"github.com/dougbarrett/gux/api"
	"github.com/dougbarrett/gux/core"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	// Load environment variables
	core.LoadEnv(".env")

	// Set up database
	db, err := gorm.Open(sqlite.Open("test-admin.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect database:", err)
	}

	// Auto-migrate models
	db.AutoMigrate(&models.User{})

	// Seed admin user if not exists
	seedAdminUser(db)

	// Initialize app
	app := core.New()
	app.SetDB(db)
	app.SetTitle("Test Admin")
	app.DarkMode() // Enable dark mode for consistent styling

	// Configure authentication
	app.EnableAuth(core.AuthConfig{
		SessionStore: core.NewMemorySessionStore(),
		CookieName:   "__test_admin_session",
		CookieMaxAge: 86400 * 7, // 7 days
		LoginPath:    "/login",
	})

	// Register API endpoints
	registerAPIEndpoints(app, db)

	// Register routes
	// Public routes (login page)
	app.Routes().
		Hybrid("/login", pages.Login)

	// Admin routes with layout wrapper
	app.RouteGroup("/admin", core.WithBundle("admin")).
		Hybrid("/", pages.AdminLayout(pages.Dashboard)).
		Hybrid("/users", pages.AdminLayout(pages.Users)).
		Hybrid("/users/:id", pages.AdminLayout(pages.UserDetail)).
		Hybrid("/settings", pages.AdminLayout(pages.Settings))

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8095"
	}
	fmt.Printf("Test Admin running at http://localhost:%s\n", port)
	fmt.Println("Login: admin@example.com / admin123")
	app.Run(":" + port)
}

// seedAdminUser creates an admin user if none exists.
func seedAdminUser(db *gorm.DB) {
	var count int64
	db.Model(&models.User{}).Count(&count)
	if count == 0 {
		admin := &models.User{
			Email: "admin@example.com",
			Name:  "Admin User",
			Role:  "admin",
		}
		admin.SetPassword("admin123")
		db.Create(admin)
		log.Println("Created admin user: admin@example.com / admin123")
	}
}

// registerAPIEndpoints sets up typed API endpoints.
func registerAPIEndpoints(app *core.App, db *gorm.DB) {
	// Login endpoint - must be Public() since users aren't authenticated yet
	core.API(app, "POST", "/api/login", func(ctx *core.APIContext, req dto.LoginRequest) (dto.LoginResponse, error) {
		// Find user
		var user models.User
		if err := db.Where("email = ?", req.Email).First(&user).Error; err != nil {
			return dto.LoginResponse{Error: "Invalid email or password"}, nil
		}

		// Check password
		if !user.CheckPassword(req.Password) {
			return dto.LoginResponse{Error: "Invalid email or password"}, nil
		}

		// Create session
		if err := ctx.Login(&core.SessionUser{
			ID:    fmt.Sprintf("%d", user.ID),
			Email: user.Email,
			Name:  user.Name,
			Roles: []string{user.Role},
		}); err != nil {
			return dto.LoginResponse{}, api.InternalError("Failed to create session")
		}

		return dto.LoginResponse{Success: true, Redirect: "/admin"}, nil
	}).Public()

	// Logout endpoint
	core.API(app, "POST", "/api/logout", func(ctx *core.APIContext, req struct{}) (dto.LogoutResponse, error) {
		ctx.Logout()
		return dto.LogoutResponse{Success: true, Redirect: "/login"}, nil
	})
}
