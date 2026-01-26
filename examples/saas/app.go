package main

import (
	"log"

	"github.com/dougbarrett/gux/core"
	"github.com/dougbarrett/gux/examples/saas/.gux/api"
	"github.com/dougbarrett/gux/examples/saas/dto"
	"github.com/dougbarrett/gux/examples/saas/models"
	"github.com/dougbarrett/gux/examples/saas/pages"
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
	db.AutoMigrate(&models.Project{})

	// Seed sample projects if none exist
	var count int64
	db.Model(&models.Project{}).Count(&count)
	if count == 0 {
		db.Create(&models.Project{
			Name:        "Website Redesign",
			Description: "Complete redesign of the company website with modern UI",
			Status:      "active",
		})
		db.Create(&models.Project{
			Name:        "Mobile App",
			Description: "Native mobile application for iOS and Android",
			Status:      "active",
		})
		db.Create(&models.Project{
			Name:        "API Integration",
			Description: "Third-party API integrations for payment and analytics",
			Status:      "completed",
		})
	}

	// Initialize API with database for server-side queries
	api.Init(db)

	app := core.New()
	app.SetDB(db)
	app.SetTitle("SaaS Dashboard")
	app.DarkMode() // Enable dark mode for Tailwind dark: variants

	// Register CRUD for Project model
	// Creates: GET/POST /__gux_api/crud/projects
	//          GET/PUT/DELETE /__gux_api/crud/projects/:id
	app.CRUD(models.Project{},
		core.WithListDTO(dto.ProjectList{}),
		core.WithDetailDTO(dto.ProjectDetail{}),
	)

	// Register routes (single bundle)
	app.Routes().
		Hybrid("/", pages.Dashboard).
		Hybrid("/projects", pages.Projects).
		Hybrid("/projects/new", pages.ProjectNew).
		Hybrid("/projects/:id", pages.ProjectDetail).
		Hybrid("/projects/:id/edit", pages.ProjectEdit).
		Hybrid("/settings", pages.Settings).
		Hybrid("/profile", pages.Profile)

	app.Run(":8082")
}
