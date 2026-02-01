package main

import (
	"fmt"
	"log"

	"github.com/dougbarrett/gux/core"
	"github.com/dougbarrett/gux/examples/audit/admin"
	"github.com/dougbarrett/gux/examples/audit/dto"
	"github.com/dougbarrett/gux/examples/audit/guxgen/api"
	"github.com/dougbarrett/gux/examples/audit/models"
	"github.com/dougbarrett/gux/examples/audit/pages"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Success  bool   `json:"success"`
	Error    string `json:"error,omitempty"`
	Redirect string `json:"redirect,omitempty"`
}

func main() {
	// Database setup
	db, err := gorm.Open(sqlite.Open("audit-demo.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}

	// Auto-migrate all models including AuditEntry
	db.AutoMigrate(&models.User{}, &models.Document{}, &core.AuditEntry{})

	// Initialize API with database
	api.Init(db)

	// Create app
	app := core.New()
	app.SetDB(db)
	app.SetTitle("Document Manager")
	app.DarkMode()

	// Auth setup
	app.EnableAuth(core.AuthConfig{
		SessionStore: core.NewMemorySessionStore(),
		LoginPath:    "/login",
	})

	// Login API endpoint
	core.API(app, "POST", "/api/login", func(ctx *core.APIContext, req LoginRequest) (LoginResponse, error) {
		loginDB := ctx.DB().(*gorm.DB)

		var user models.User
		if err := loginDB.Where("email = ?", req.Email).First(&user).Error; err != nil {
			return LoginResponse{Success: false, Error: "Invalid email or password"}, nil
		}

		if !user.CheckPassword(req.Password) {
			return LoginResponse{Success: false, Error: "Invalid email or password"}, nil
		}

		ctx.Login(&core.SessionUser{
			ID:    fmt.Sprintf("%d", user.ID),
			Email: user.Email,
			Name:  user.Name,
			Roles: []string{user.Role},
		})

		return LoginResponse{Success: true, Redirect: "/admin/dashboard"}, nil
	}).Public()

	// CRUD with audit logging
	app.CRUD(models.User{},
		core.WithListDTO(dto.UserList{}),
		core.WithDetailDTO(dto.UserDetail{}),
		core.WithAuditLog("PasswordHash"),
		core.WithRoles("admin"),
	)

	app.CRUD(models.Document{},
		core.WithListDTO(dto.DocumentList{}),
		core.WithDetailDTO(dto.DocumentDetail{}),
		core.WithAuditLog(),
	)

	// Audit entries — read-only for admins
	app.CRUD(core.AuditEntry{},
		core.WithListDTO(dto.AuditEntryList{}),
		core.WithRoles("admin"),
	)

	// Public routes
	app.Routes().
		Hybrid("/", pages.Home).
		Hybrid("/login", pages.Login)

	// Admin routes (separate bundle)
	app.RouteGroup("/admin", core.WithBundle("admin")).
		Hybrid("/dashboard", admin.Dashboard).
		Hybrid("/documents", admin.Documents).
		Hybrid("/documents/new", admin.DocumentNew).
		Hybrid("/documents/:id", admin.DocumentDetail).
		Hybrid("/documents/:id/edit", admin.DocumentEdit).
		Hybrid("/users", admin.Users).
		Hybrid("/users/:id", admin.UserDetail).
		Hybrid("/audit", admin.AuditLog)

	// Seed data
	seedData(db)

	fmt.Println("Starting server on :8085")
	fmt.Println("Login with: admin@example.com / admin123")
	app.Run(":8085")
}

func seedData(db *gorm.DB) {
	var count int64
	db.Model(&models.User{}).Count(&count)
	if count > 0 {
		return
	}

	fmt.Println("Seeding database...")

	// Create admin user
	adminUser := models.User{
		Email: "admin@example.com",
		Name:  "Admin User",
		Role:  "admin",
	}
	adminUser.SetPassword("admin123")
	db.Create(&adminUser)

	// Create regular users
	user1 := models.User{
		Email: "john@example.com",
		Name:  "John Doe",
		Role:  "user",
	}
	user1.SetPassword("password123")
	db.Create(&user1)

	user2 := models.User{
		Email: "jane@example.com",
		Name:  "Jane Smith",
		Role:  "user",
	}
	user2.SetPassword("password123")
	db.Create(&user2)

	// Create documents
	db.Create(&models.Document{
		Title:    "Company Policy Update",
		Content:  "This document outlines the new company policies effective Q1 2026.",
		Status:   "published",
		AuthorID: adminUser.ID,
	})

	db.Create(&models.Document{
		Title:    "Q4 Financial Report - Draft",
		Content:  "Preliminary financial results for Q4 2025.",
		Status:   "draft",
		AuthorID: user1.ID,
	})

	db.Create(&models.Document{
		Title:    "Product Roadmap 2026",
		Content:  "Strategic product roadmap for the upcoming year.",
		Status:   "published",
		AuthorID: user2.ID,
	})

	// Create sample audit entries for demo
	entries := []core.AuditEntry{
		{
			Action:     core.AuditActionCreate,
			EntityType: "User",
			EntityID:   adminUser.ID,
			UserID:     fmt.Sprintf("%d", adminUser.ID),
			UserEmail:  adminUser.Email,
			IPAddress:  "127.0.0.1",
			Changes:    `{"email":"admin@example.com","name":"Admin User","role":"admin"}`,
		},
		{
			Action:     core.AuditActionCreate,
			EntityType: "Document",
			EntityID:   1,
			UserID:     fmt.Sprintf("%d", adminUser.ID),
			UserEmail:  adminUser.Email,
			IPAddress:  "127.0.0.1",
			Changes:    `{"title":"Company Policy Update","status":"published"}`,
		},
		{
			Action:     core.AuditActionUpdate,
			EntityType: "Document",
			EntityID:   2,
			UserID:     fmt.Sprintf("%d", user1.ID),
			UserEmail:  user1.Email,
			IPAddress:  "192.168.1.100",
			Changes:    `{"status":{"old":"draft","new":"review"},"content":{"old":"Initial draft...","new":"Preliminary financial results..."}}`,
		},
		{
			Action:     core.AuditActionCreate,
			EntityType: "User",
			EntityID:   user2.ID,
			UserID:     fmt.Sprintf("%d", adminUser.ID),
			UserEmail:  adminUser.Email,
			IPAddress:  "127.0.0.1",
			Changes:    `{"email":"jane@example.com","name":"Jane Smith","role":"user"}`,
		},
	}
	for _, e := range entries {
		db.Create(&e)
	}

	fmt.Println("Database seeded successfully")
}
