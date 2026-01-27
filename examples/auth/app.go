package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/dougbarrett/gux/core"
	"github.com/dougbarrett/gux/examples/auth/dto"
	"github.com/dougbarrett/gux/examples/auth/guxgen/api"
	"github.com/dougbarrett/gux/examples/auth/models"
	"github.com/dougbarrett/gux/examples/auth/pages"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	// Set up database
	db, err := gorm.Open(sqlite.Open("app.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect database:", err)
	}

	// Auto-migrate models
	db.AutoMigrate(&models.User{}, &models.PasswordReset{})

	// Seed users if none exist
	var userCount int64
	db.Model(&models.User{}).Count(&userCount)
	if userCount == 0 {
		// Create admin user (for systems without self-signup)
		admin := models.User{
			Email:    "admin@example.com",
			Name:     "Admin User",
			Verified: true,
			Role:     "admin",
		}
		admin.SetPassword("admin123")
		db.Create(&admin)

		// Create regular demo user
		user := models.User{
			Email:    "demo@example.com",
			Name:     "Demo User",
			Verified: true,
			Role:     "user",
		}
		user.SetPassword("demo123")
		db.Create(&user)

		fmt.Println("Seeded users:")
		fmt.Println("  Admin: admin@example.com / admin123")
		fmt.Println("  User:  demo@example.com / demo123")
	}

	// Initialize API with database
	api.Init(db)

	app := core.New()
	app.SetDB(db)
	app.SetTitle("Auth Example")

	// Enable session-based authentication
	app.EnableAuth(core.AuthConfig{
		SessionStore: core.NewMemorySessionStore(),
		LoginPath:    "/login",
	})

	// Register CRUD for User with DTO and password hooks
	app.CRUD(models.User{},
		core.WithListDTO(dto.UserDTO{}),
		core.WithDetailDTO(dto.UserDTO{}),
		core.WithCreateHook(func(data map[string]interface{}) (interface{}, error) {
			user := &models.User{}
			if email, ok := data["email"].(string); ok {
				user.Email = email
			} else {
				return nil, errors.New("email is required")
			}
			if name, ok := data["name"].(string); ok {
				user.Name = name
			}
			if password, ok := data["password"].(string); ok && password != "" {
				if err := user.SetPassword(password); err != nil {
					return nil, errors.New("failed to hash password")
				}
			} else {
				return nil, errors.New("password is required")
			}
			return user, nil
		}),
	)

	// Login API endpoint - typed with automatic JSON handling
	core.API(app, "POST", "/api/login", func(ctx *core.APIContext, req dto.LoginRequest) (dto.LoginResponse, error) {
		db := ctx.DB().(*gorm.DB)

		// Find user by email
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
			return dto.LoginResponse{Error: "Failed to create session"}, nil
		}

		return dto.LoginResponse{Success: true, Redirect: "/dashboard"}, nil
	})

	// Logout API endpoint - typed with automatic JSON handling
	core.API(app, "POST", "/api/logout", func(ctx *core.APIContext, _ struct{}) (dto.LogoutResponse, error) {
		ctx.Logout()
		return dto.LogoutResponse{Success: true, Redirect: "/"}, nil
	})

	// Public routes
	app.Routes().
		Hybrid("/", pages.Home).
		Hybrid("/login", pages.Login).
		Hybrid("/register", pages.Register).
		Hybrid("/forgot", pages.Forgot).
		Hybrid("/reset/:token", pages.Reset).
		Hybrid("/verify/:token", pages.Verify)

	// Protected routes - requires authentication
	app.Routes().
		Hybrid("/dashboard", pages.Dashboard).Protected()

	// Admin-only routes - demonstrates RBAC
	app.Routes().
		Hybrid("/admin", pages.Admin).Protected().WithRoles("admin")

	app.Run(":8082")
}
