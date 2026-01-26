package main

import (
	"log"
	"time"

	"github.com/dougbarrett/gux/core"
	"github.com/dougbarrett/gux/examples/admin/.gux/api"
	"github.com/dougbarrett/gux/examples/admin/dto"
	"github.com/dougbarrett/gux/examples/admin/models"
	"github.com/dougbarrett/gux/examples/admin/pages"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	// Set up database
	db, err := gorm.Open(sqlite.Open("admin.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect database:", err)
	}

	// Auto-migrate models
	db.AutoMigrate(&models.User{}, &models.ActivityLog{})

	// Seed data if empty
	var userCount int64
	db.Model(&models.User{}).Count(&userCount)
	if userCount == 0 {
		seedData(db)
	}

	// Initialize API with database for server-side queries
	api.Init(db)

	app := core.New()
	app.SetDB(db)
	app.SetTitle("Admin Panel")
	app.DarkMode() // Enable dark mode for Tailwind dark: variants

	// Register CRUD for User model
	// Creates: GET/POST /__gux_api/crud/users
	//          GET/PUT/DELETE /__gux_api/crud/users/:id
	app.CRUD(models.User{},
		core.WithListDTO(dto.UserList{}),
		core.WithDetailDTO(dto.UserDetail{}),
	)

	// Register CRUD for ActivityLog model (read-only, no create/update from UI)
	// Creates: GET /__gux_api/crud/activitylogs
	//          GET /__gux_api/crud/activitylogs/:id
	app.CRUD(models.ActivityLog{},
		core.WithListDTO(dto.ActivityLogList{}),
		core.WithDetailDTO(dto.ActivityLogList{}),
	)

	// Register routes (single bundle)
	app.Routes().
		Hybrid("/", pages.Dashboard).
		Hybrid("/users", pages.Users).
		Hybrid("/users/new", pages.UserNew).
		Hybrid("/users/:id", pages.UserDetail).
		Hybrid("/users/:id/edit", pages.UserEdit).
		Hybrid("/activity", pages.Activity).
		Hybrid("/settings", pages.Settings)

	app.Run(":8084")
}

// seedData creates initial admin user and sample data.
func seedData(db *gorm.DB) {
	// Create admin user
	admin := &models.User{
		Email:  "admin@example.com",
		Name:   "Admin User",
		Role:   "admin",
		Status: "active",
	}
	admin.SetPassword("admin123")
	db.Create(admin)

	// Create regular users with varied roles and statuses
	now := time.Now()
	lastWeek := now.Add(-7 * 24 * time.Hour)

	user1 := &models.User{
		Email:       "john@example.com",
		Name:        "John Smith",
		Role:        "user",
		Status:      "active",
		LastLoginAt: &now,
	}
	user1.SetPassword("password123")
	db.Create(user1)

	user2 := &models.User{
		Email:       "jane@example.com",
		Name:        "Jane Doe",
		Role:        "editor",
		Status:      "active",
		LastLoginAt: &lastWeek,
	}
	user2.SetPassword("password123")
	db.Create(user2)

	user3 := &models.User{
		Email:  "bob@example.com",
		Name:   "Bob Wilson",
		Role:   "user",
		Status: "suspended",
	}
	user3.SetPassword("password123")
	db.Create(user3)

	// Create sample activity logs
	userID1 := admin.ID
	userID2 := user1.ID
	userID3 := user2.ID
	entityID1 := user1.ID
	entityID2 := user2.ID
	entityID3 := user3.ID

	activityLogs := []models.ActivityLog{
		{
			UserID:      &userID1,
			Action:      "login",
			Entity:      "user",
			EntityID:    &userID1,
			Description: "Admin User logged in",
			IPAddress:   "192.168.1.1",
		},
		{
			UserID:      &userID1,
			Action:      "created",
			Entity:      "user",
			EntityID:    &entityID1,
			Description: "Created user John Smith",
			IPAddress:   "192.168.1.1",
		},
		{
			UserID:      &userID2,
			Action:      "login",
			Entity:      "user",
			EntityID:    &userID2,
			Description: "John Smith logged in",
			IPAddress:   "10.0.0.5",
		},
		{
			UserID:      &userID1,
			Action:      "updated",
			Entity:      "user",
			EntityID:    &entityID2,
			Description: "Updated user Jane Doe role to editor",
			IPAddress:   "192.168.1.1",
		},
		{
			UserID:      &userID3,
			Action:      "updated",
			Entity:      "setting",
			EntityID:    &entityID3,
			Description: "Changed notification preferences",
			IPAddress:   "172.16.0.10",
		},
	}

	for _, log := range activityLogs {
		db.Create(&log)
	}
}
