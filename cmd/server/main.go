package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/a2sh3r/gophkeeper/internal/auth"
	"github.com/a2sh3r/gophkeeper/internal/config"
	"github.com/a2sh3r/gophkeeper/internal/db"
	"github.com/a2sh3r/gophkeeper/internal/logger"
	"github.com/a2sh3r/gophkeeper/internal/server"
	"github.com/a2sh3r/gophkeeper/internal/storage"
	"github.com/a2sh3r/gophkeeper/pkg/version"
	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
	"go.uber.org/zap"
)

func main() {
	var (
		showVersion = flag.Bool("version", false, "Show version information")
	)
	flag.Parse()

	if *showVersion {
		fmt.Println(version.Info())
		os.Exit(0)
	}

	cfg := config.Load()

	if err := logger.Initialize(cfg.Server.LogLevel); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	var store server.Storage

	switch cfg.Database.Type {
	case "postgres":
		logger.Log.Info("Using PostgreSQL database", zap.String("host", cfg.Database.Host))
		database, err := db.New(cfg.GetDSN())
		if err != nil {
			logger.Log.Fatal("Failed to connect to PostgreSQL", zap.Error(err))
		}
		defer func() {
			if err := database.Close(); err != nil {
				logger.Log.Error("Failed to close database", zap.Error(err))
			}
		}()
		store = storage.NewPostgresStorage(database.Conn())
	case "memory":
		logger.Log.Info("Using in-memory storage")
		store = storage.NewMemoryStorage()
	default:
		logger.Log.Fatal("Unsupported database type", zap.String("type", cfg.Database.Type))
	}

	jwtManager := auth.NewJWTManager(cfg.JWT.Secret, cfg.JWT.TokenExpiry)

	srv := server.NewServer(store, jwtManager)

	router := mux.NewRouter()
	srv.RegisterRoutes(router)

	n := negroni.New()
	n.Use(negroni.NewLogger())
	n.Use(negroni.NewRecovery())
	n.UseHandler(router)

	addr := cfg.GetServerAddr()
	logger.Log.Info("Starting GophKeeper server",
		zap.String("address", addr),
		zap.String("version", version.ShortInfo()),
		zap.String("database", cfg.Database.Type))

	if err := http.ListenAndServe(addr, n); err != nil {
		logger.Log.Fatal("Server failed to start", zap.Error(err))
	}
}
