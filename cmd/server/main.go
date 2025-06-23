package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tomassar/credit-score-evaluator/internal/config"
	"github.com/tomassar/credit-score-evaluator/internal/handlers"
	"github.com/tomassar/credit-score-evaluator/internal/services"
	"github.com/tomassar/credit-score-evaluator/internal/storage"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database
	db, err := storage.NewSQLiteDB(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize services
	borrowerService := services.NewBorrowerService(db)
	ruleService := services.NewRuleService(db)
	scoringService := services.NewScoringService(ruleService)

	// Initialize session store
	sessionStore := sessions.NewCookieStore([]byte(cfg.SessionSecret))
	sessionStore.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 30, // 30 days
		HttpOnly: true,
		Secure:   cfg.HTTPSEnabled,
		SameSite: http.SameSiteStrictMode,
	}

	// Initialize handlers
	h := handlers.New(borrowerService, ruleService, scoringService, sessionStore)

	// Setup routes
	router := mux.NewRouter()
	setupRoutes(router, h)

	// Create server
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on port %s", cfg.Port)
		if cfg.HTTPSEnabled {
			if err := server.ListenAndServeTLS(cfg.CertFile, cfg.KeyFile); err != nil && err != http.ErrServerClosed {
				log.Fatalf("Failed to start HTTPS server: %v", err)
			}
		} else {
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("Failed to start server: %v", err)
			}
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

func setupRoutes(router *mux.Router, h *handlers.Handler) {
	// Static files
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/",
		http.FileServer(http.Dir("web/static/"))))

	// API routes
	api := router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/upload", h.UploadCSV).Methods("POST")
	api.HandleFunc("/borrowers", h.GetBorrowers).Methods("GET")
	api.HandleFunc("/borrowers/export", h.ExportBorrowers).Methods("GET")
	api.HandleFunc("/rules", h.GetRules).Methods("GET")
	api.HandleFunc("/rules", h.CreateRule).Methods("POST")
	api.HandleFunc("/rules/{id}", h.UpdateRule).Methods("PUT")
	api.HandleFunc("/rules/{id}", h.DeleteRule).Methods("DELETE")
	api.HandleFunc("/recalculate", h.RecalculateScores).Methods("POST")

	// Page routes
	router.HandleFunc("/", h.HomePage).Methods("GET")
	router.HandleFunc("/upload", h.UploadPage).Methods("GET")
	router.HandleFunc("/borrowers", h.BorrowersPage).Methods("GET")
	router.HandleFunc("/rules", h.RulesPage).Methods("GET")

	// HTMX partial routes
	htmx := router.PathPrefix("/htmx").Subrouter()
	htmx.HandleFunc("/borrowers-table", h.BorrowersTable).Methods("GET")
	htmx.HandleFunc("/rules-list", h.RulesList).Methods("GET")
	htmx.HandleFunc("/upload-form", h.UploadForm).Methods("GET")
	htmx.HandleFunc("/rule-modal", h.RuleModal).Methods("GET")
	htmx.HandleFunc("/rule-modal/{id}", h.RuleModal).Methods("GET")
	htmx.HandleFunc("/rules", h.ProcessRule).Methods("POST")
	htmx.HandleFunc("/rules/{id}", h.ProcessRule).Methods("POST")
	htmx.HandleFunc("/rules/{id}/delete", h.DeleteRuleHTMX).Methods("DELETE", "POST")
	htmx.HandleFunc("/flash-message", h.FlashMessage).Methods("GET")
	htmx.HandleFunc("/dashboard-stats", h.DashboardStats).Methods("GET")
	htmx.HandleFunc("/recalculate", h.RecalculateScoresHTMX).Methods("POST")
	htmx.HandleFunc("/upload-status", h.UploadStatus).Methods("GET")
}
