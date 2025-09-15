package main

import (
	"log"
	"os"
	"time"

	"github.com/DrWeltschmerz/jwt-auth/pkg/authjwt"
	ginadapter "github.com/DrWeltschmerz/users-adapter-gin/ginadapter"
	gormAdapter "github.com/DrWeltschmerz/users-adapter-gorm/gorm"
	"github.com/DrWeltschmerz/users-core"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	// Use file-based SQLite for persistence, or in-memory for ephemeral
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		// Use in-memory DB for CI/tests by default
		dbPath = ":memory:"
	}
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}
	if err := db.AutoMigrate(&gormAdapter.GormUser{}, &gormAdapter.GormRole{}); err != nil {
		log.Fatalf("failed to migrate db: %v", err)
	}

	// --- SEED ADMIN ROLE AND USER FOR TESTING ---
	// 1. Ensure admin role exists
	var adminRole gormAdapter.GormRole
	if err := db.Where("name = ?", "admin").First(&adminRole).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			adminRole = gormAdapter.GormRole{Name: "admin"}
			if err := db.Create(&adminRole).Error; err != nil {
				log.Fatalf("failed to create admin role: %v", err)
			}
		} else {
			log.Fatalf("failed to query admin role: %v", err)
		}
	}

	// 2. Ensure admin user exists with admin role
	var adminUser gormAdapter.GormUser
	if err := db.Where("email = ?", "admin@example.com").First(&adminUser).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// hash password
			hasher := authjwt.NewBcryptHasher()
			hashed, err := hasher.Hash("adminpass")
			if err != nil {
				log.Fatalf("failed to hash admin password: %v", err)
			}
			adminUser = gormAdapter.GormUser{
				Username:       "admin",
				Email:          "admin@example.com",
				HashedPassword: hashed,
				RoleID:         adminRole.ID,
				LastSeen:       time.Now(),
			}
			if err := db.Create(&adminUser).Error; err != nil {
				log.Fatalf("failed to create admin user: %v", err)
			}
		} else {
			log.Fatalf("failed to query admin user: %v", err)
		}
	} else {
		// ensure admin user has admin role
		if adminUser.RoleID != adminRole.ID {
			adminUser.RoleID = adminRole.ID
			if err := db.Save(&adminUser).Error; err != nil {
				log.Fatalf("failed to update admin user role: %v", err)
			}
		}
	}

	userRepo := gormAdapter.NewGormUserRepository(db)
	roleRepo := gormAdapter.NewGormRoleRepository(db)
	hasher := authjwt.NewBcryptHasher()

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "test-secret-very-long-and-secure"
		os.Setenv("JWT_SECRET", jwtSecret)
	}
	tokenizer := authjwt.NewJWTTokenizer()

	svc := users.NewService(userRepo, roleRepo, hasher, tokenizer)

	r := gin.Default()
	ginadapter.RegisterRoutes(r, svc, tokenizer)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Starting users service on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
