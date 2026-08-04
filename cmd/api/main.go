package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/g-mnt/piano-paralelo-api/internal/config"
	"github.com/g-mnt/piano-paralelo-api/internal/db"
	"github.com/g-mnt/piano-paralelo-api/internal/handler"
	"github.com/g-mnt/piano-paralelo-api/internal/middleware"
	"github.com/g-mnt/piano-paralelo-api/internal/seed"
)

func main() {
	seedOnly := flag.Bool("seed", false, "run seed and exit")
	flag.Parse()

	cfg := config.Load()

	pool, err := db.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	// Run migrations
	if err := runMigrations(cfg.DatabaseURL); err != nil {
		log.Fatalf("migrations failed: %v", err)
	}

	// Run seed
	if err := seed.Run(pool); err != nil {
		log.Fatalf("seed failed: %v", err)
	}

	if *seedOnly {
		log.Println("seed complete, exiting")
		return
	}

	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// CORS
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"http://localhost:5173"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	r.Use(cors.New(corsConfig))

	// Health
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Auth (public)
	r.POST("/auth/register", handler.Register(pool, cfg.JWTSecret))
	r.POST("/auth/login", handler.Login(pool, cfg.JWTSecret))

	// Authenticated routes
	auth := r.Group("/")
	auth.Use(middleware.RequireAuth(cfg.JWTSecret))
	{
		auth.GET("/profile", handler.GetProfile(pool))
		auth.PUT("/profile", handler.UpdateProfile(pool))

		auth.GET("/curriculum/weeks", handler.ListWeeks(pool))
		auth.GET("/curriculum/weeks/:n", handler.GetWeek(pool))

		auth.POST("/sessions", handler.GetOrCreateSession(pool))
		auth.PATCH("/sessions/:id/tasks/:taskId", handler.ToggleTask(pool))
		auth.GET("/sessions/streak", handler.GetStreak(pool))

		auth.GET("/repertoire", handler.ListPieces(pool))
		auth.PATCH("/repertoire/:id/progress", handler.UpdateProgress(pool))
	}

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("starting server on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func runMigrations(databaseURL string) error {
	m, err := migrate.New("file://migrations", databaseURL)
	if err != nil {
		return fmt.Errorf("migrate.New: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migrate up: %w", err)
	}
	return nil
}
