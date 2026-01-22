package main

import (
	"log"

	"github.com/dougbarrett/gux/core"
	"github.com/dougbarrett/gux/examples/minimal/.gux/api"
	"github.com/dougbarrett/gux/examples/minimal/admin"
	"github.com/dougbarrett/gux/examples/minimal/models"
	"github.com/dougbarrett/gux/examples/minimal/pages"
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
	db.AutoMigrate(&models.Counter{})

	// Create initial counter if none exists
	var count int64
	db.Model(&models.Counter{}).Count(&count)
	if count == 0 {
		db.Create(&models.Counter{Name: "default", Value: 0})
	}

	// Initialize API with database for server-side queries
	api.Init(db)

	app := core.New()
	app.SetDB(db)
	app.SetTitle("Gux Counter")

	// Register CRUD for Counter model
	// Creates: GET/POST /__gux_api/crud/counters
	//          GET/PUT/DELETE /__gux_api/crud/counters/:id
	app.CRUD(models.Counter{})

	// Public routes (default bundle)
	app.Routes().
		Hybrid("/", pages.Home).
		Hybrid("/about", pages.About)

	// Admin routes (separate "admin" bundle)
	app.RouteGroup("/admin", core.WithBundle("admin")).
		Hybrid("/", admin.Dashboard).
		Hybrid("/account", admin.Account)

	app.Run(":8081")
}
