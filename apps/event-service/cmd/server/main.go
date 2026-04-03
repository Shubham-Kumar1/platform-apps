package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"

	"github.com/eventhub/event-service/internal/db"
	"github.com/eventhub/event-service/internal/handler"
)

func main() {
	// MySQL (own DB for events)
	mysqlDSN := os.Getenv("MYSQL_DSN")
	if mysqlDSN == "" {
		mysqlDSN = "root:password@tcp(mysql:3306)/events?parseTime=true"
	}
	mysqlDB, err := sql.Open("mysql", mysqlDSN)
	if err != nil {
		log.Fatalf("Failed to connect to MySQL: %v", err)
	}
	mysqlDB.SetMaxOpenConns(25)
	mysqlDB.SetMaxIdleConns(5)
	mysqlDB.SetConnMaxLifetime(5 * time.Minute)

	// Central PostgreSQL (read-only, for user/org validation)
	centralDSN := os.Getenv("CENTRAL_DB_URL")
	if centralDSN == "" {
		centralDSN = "postgresql://ro_tenant_alpha:password@pgbouncer.platform-services:5432/eventhub_central?sslmode=disable"
	}
	centralDB, err := sql.Open("postgres", centralDSN)
	if err != nil {
		log.Fatalf("Failed to connect to Central DB: %v", err)
	}
	centralDB.SetMaxOpenConns(10)
	centralDB.SetMaxIdleConns(3)

	waitForDB(mysqlDB, "MySQL")
	waitForDB(centralDB, "CentralDB")
	runMigrations(mysqlDB)

	userStore := db.NewUserStore(centralDB)
	eventHandler := handler.NewEventHandler(mysqlDB, userStore)
	bookingHandler := handler.NewBookingHandler(mysqlDB, userStore)
	healthHandler := handler.NewHealthHandler(mysqlDB, centralDB)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(middleware.Heartbeat("/ping"))

	r.Get("/healthz", healthHandler.Liveness)
	r.Get("/readyz", healthHandler.Readiness)

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/events", func(r chi.Router) {
			r.Get("/", eventHandler.List)
			r.Post("/", eventHandler.Create)
			r.Get("/{eventID}", eventHandler.Get)
			r.Put("/{eventID}", eventHandler.Update)
			r.Delete("/{eventID}", eventHandler.Delete)
		})
		r.Route("/events/{eventID}/bookings", func(r chi.Router) {
			r.Get("/", bookingHandler.List)
			r.Post("/", bookingHandler.Create)
		})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	srv := &http.Server{Addr: ":" + port, Handler: r}

	go func() {
		log.Printf("Event Service listening on :%s", port)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
	mysqlDB.Close()
	centralDB.Close()
}

func waitForDB(d *sql.DB, name string) {
	for i := 0; i < 30; i++ {
		if err := d.Ping(); err == nil {
			log.Printf("%s is ready", name)
			return
		}
		log.Printf("Waiting for %s... (%d/30)", name, i+1)
		time.Sleep(2 * time.Second)
	}
	log.Fatalf("%s did not become ready in time", name)
}

func runMigrations(d *sql.DB) {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS events (
			id VARCHAR(36) PRIMARY KEY,
			org_id VARCHAR(36) NOT NULL,
			title VARCHAR(255) NOT NULL,
			description TEXT,
			venue VARCHAR(255),
			start_time DATETIME NOT NULL,
			end_time DATETIME NOT NULL,
			max_attendees INT DEFAULT 0,
			status ENUM('draft','published','cancelled','completed') DEFAULT 'draft',
			created_by VARCHAR(36) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			INDEX idx_org_id (org_id),
			INDEX idx_status (status)
		)`,
		`CREATE TABLE IF NOT EXISTS ticket_types (
			id VARCHAR(36) PRIMARY KEY,
			event_id VARCHAR(36) NOT NULL,
			name VARCHAR(100) NOT NULL,
			price DECIMAL(10,2) NOT NULL DEFAULT 0,
			quantity INT NOT NULL DEFAULT 0,
			sold INT NOT NULL DEFAULT 0,
			FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS bookings (
			id VARCHAR(36) PRIMARY KEY,
			user_id VARCHAR(36) NOT NULL,
			event_id VARCHAR(36) NOT NULL,
			ticket_type_id VARCHAR(36) NOT NULL,
			status ENUM('confirmed','cancelled','refunded') DEFAULT 'confirmed',
			booked_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_user_id (user_id),
			INDEX idx_event_id (event_id),
			FOREIGN KEY (event_id) REFERENCES events(id),
			FOREIGN KEY (ticket_type_id) REFERENCES ticket_types(id)
		)`,
	}
	for _, q := range queries {
		if _, err := d.Exec(q); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
	}
	log.Println("MySQL migrations completed")
}
