package main

import (
	"errors"
	"log"

	"github.com/dougbarrett/gux/core"
	"github.com/dougbarrett/gux/examples/auth/.gux/api"
	"github.com/dougbarrett/gux/examples/auth/dto"
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

	// Seed demo user if none exists
	var userCount int64
	db.Model(&models.User{}).Count(&userCount)
	if userCount == 0 {
		user := models.User{Email: "demo@example.com", Name: "Demo User", Verified: true}
		user.SetPassword("demo123")
		db.Create(&user)
	}

	// Initialize API with database
	api.Init(db)

	app := core.New()
	app.SetDB(db)
	app.SetTitle("Auth Example")

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

	// Auth routes (all public, single bundle)
	app.Routes().
		Hybrid("/", pages.Home).
		Hybrid("/login", pages.Login).
		Hybrid("/register", pages.Register).
		Hybrid("/forgot", pages.Forgot).
		Hybrid("/reset/:token", pages.Reset).
		Hybrid("/verify/:token", pages.Verify).
		Hybrid("/dashboard", pages.Dashboard)

	app.Run(":8082")
}
