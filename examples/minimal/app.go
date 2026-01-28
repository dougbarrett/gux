package main

import (
	"errors"
	"log"

	"github.com/dougbarrett/gux/core"
	"github.com/dougbarrett/gux/examples/minimal/guxgen/api"
	"github.com/dougbarrett/gux/examples/minimal/admin"
	"github.com/dougbarrett/gux/examples/minimal/dto"
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
	db.AutoMigrate(&models.Counter{}, &models.User{}, &models.Post{}, &models.Client{})

	// Create initial counter if none exists
	var count int64
	db.Model(&models.Counter{}).Count(&count)
	if count == 0 {
		db.Create(&models.Counter{Name: "default", Value: 0})
	}

	// Seed some users if none exist
	var userCount int64
	db.Model(&models.User{}).Count(&userCount)
	if userCount == 0 {
		// Create users with properly hashed passwords
		admin := models.User{Email: "admin@example.com", Name: "Admin User", Role: "admin"}
		admin.SetPassword("admin123")
		db.Create(&admin)

		user := models.User{Email: "user@example.com", Name: "Regular User", Role: "user"}
		user.SetPassword("user123")
		db.Create(&user)

		// Create some posts for the users
		db.Create(&models.Post{Title: "Getting Started with Gux", Content: "This tutorial covers the basics of building web apps with Gux framework.", UserID: admin.ID})
		db.Create(&models.Post{Title: "Advanced CRUD Patterns", Content: "Learn how to implement custom hooks and DTOs for secure API design.", UserID: admin.ID})
		db.Create(&models.Post{Title: "My First Post", Content: "Hello from a regular user!", UserID: user.ID})
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

	// Register CRUD for User with DTOs and password hashing hooks
	// List: returns UserList (no password hash)
	// Detail: returns UserDetail (includes posts, no password hash)
	app.CRUD(models.User{},
		core.WithListDTO(dto.UserList{}),
		core.WithDetailDTO(dto.UserDetail{}, "Posts"),
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
			if role, ok := data["role"].(string); ok {
				user.Role = role
			} else {
				user.Role = "user" // default role
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
		core.WithUpdateHook(func(existing interface{}, data map[string]interface{}) (interface{}, error) {
			user := existing.(*models.User)
			if email, ok := data["email"].(string); ok {
				user.Email = email
			}
			if name, ok := data["name"].(string); ok {
				user.Name = name
			}
			if role, ok := data["role"].(string); ok {
				user.Role = role
			}
			// Only update password if provided
			if password, ok := data["password"].(string); ok && password != "" {
				if err := user.SetPassword(password); err != nil {
					return nil, errors.New("failed to hash password")
				}
			}
			return user, nil
		}),
	)

	// Register CRUD for Post with DTOs
	app.CRUD(models.Post{},
		core.WithListDTO(dto.PostList{}, "User"),
		core.WithDetailDTO(dto.PostDetail{}, "User"),
	)

	// Register CRUD for Client with DTOs
	app.CRUD(models.Client{},
		core.WithListDTO(dto.ClientList{}, "Salesperson"),
		core.WithDetailDTO(dto.ClientDetail{}, "Salesperson"),
	)

	// Public routes (default bundle)
	app.Routes().
		Hybrid("/", pages.Home).
		Hybrid("/about", pages.About)

	// Admin routes (separate "admin" bundle)
	app.RouteGroup("/admin", core.WithBundle("admin")).
		Hybrid("/", admin.Dashboard).
		Hybrid("/users", admin.Users).
		Hybrid("/users/new", admin.UserNew).
		Hybrid("/users/:id", admin.UserDetail).
		Hybrid("/users/:id/edit", admin.UserEdit).
		Hybrid("/users/:id/posts/new", admin.PostNew).
		Hybrid("/posts", admin.Posts).
		Hybrid("/posts/:id", admin.PostDetail).
		Hybrid("/posts/:id/edit", admin.PostEdit).
		Hybrid("/clients", admin.Clients).
		Hybrid("/clients/new", admin.ClientNew).
		Hybrid("/clients/:id", admin.ClientDetail).
		Hybrid("/clients/:id/edit", admin.ClientEdit).
		Hybrid("/account", admin.Account)

	app.Run(":8081")
}
