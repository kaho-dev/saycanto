package main

import (
	"database/sql"
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/kaho-dev/saycanto/backend/internal/dictionary"
	"github.com/kaho-dev/saycanto/backend/internal/health"
	"github.com/kaho-dev/saycanto/backend/internal/shared/httputil"
	_ "modernc.org/sqlite"
)

func main() {
	dbPath := flag.String("db", envOr("DB_PATH", "./dictionary.db"), "path to SQLite dictionary database")
	port := flag.String("port", envOr("PORT", "8080"), "HTTP listen port")
	corsOrigin := flag.String("cors", envOr("CORS_ORIGIN", "*"), "allowed CORS origin")
	flag.Parse()

	db, err := sql.Open("sqlite", *dbPath)
	if err != nil {
		log.Fatalf("failed to open sqlite db: %v", err)
	}
	defer db.Close()

	if err := dictionary.InitSchema(db); err != nil {
		log.Fatalf("failed to init schema: %v", err)
	}

	// Constructor-based Dependency Injection
	dictRepo := dictionary.NewSQLiteRepository(db)
	dictService := dictionary.NewService(dictRepo)
	dictController := dictionary.NewController(dictService)
	healthController := health.NewController()

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(10 * time.Second))
	r.Use(httputil.CORSMiddleware(*corsOrigin))

	healthController.RegisterRoutes(r)
	dictController.RegisterRoutes(r)

	addr := ":" + *port
	log.Printf("saycanto API listening on %s (db=%s)", addr, *dbPath)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
