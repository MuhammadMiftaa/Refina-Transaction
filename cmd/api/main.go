package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"refina-transaction/config/db"
	"refina-transaction/config/env"
	"refina-transaction/config/log"
	"refina-transaction/config/miniofs"
	"refina-transaction/interface/grpc/client"
	"refina-transaction/interface/http/router"
)

func init() {
	log.SetupLogger() // Initialize the logger configuration

	var err error
	var missing []string
	if missing, err = env.LoadByViper(); err != nil {
		log.Error("Failed to read JSON config file:" + err.Error())
		if missing, err = env.LoadNative(); err != nil {
			log.Log.Fatalf("Failed to load environment variables: %v", err)
		}
		log.SetupLogger()
		log.Info("Environment variables by .env file loaded successfully")
	} else {
		log.SetupLogger()
		log.Info("Environment variables by Viper loaded successfully")
	}

	if len(missing) > 0 {
		for _, envVar := range missing {
			log.Warn("Missing environment variable: " + envVar)
		}
	}

	log.Info("Setup Database Connection Start")
	db.SetupDatabase(env.Cfg.Database) // Initialize the database connection and run migrations
	log.Info("Setup Database Connection Success")

	log.Info("Setup MinIO Connection Start")
	miniofs.SetupMinio(env.Cfg.Minio) // Initialize MinIO connection
	log.Info("Setup MinIO Connection Success")

	log.Info("Starting Refina API...")
}

func main() {
	defer log.Info("Refina API stopped")

	// Set up the gRPC client
	grpcManager := client.GetManager()
	err := grpcManager.SetupGRPCClient()
	if err != nil {
		log.Log.Fatalf("Failed to set up gRPC client: %v", err)
	}

	// Set up the HTTP server
	httpServer := router.SetupHTTPServer()
	if httpServer != nil {
		go func() {
			if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Log.Fatalf("Failed to start HTTP server: %s\n", err)
			}
		}()
		log.Info("Starting HTTP server successfully")
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Log.Info("Shutting down servers...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Log.Fatalf("Failed to shutdown HTTP server: %v", err)
	}

	if err := grpcManager.Shutdown(ctx); err != nil {
		log.Log.Fatalf("Failed to shutdown gRPC clients: %v", err)
	}

	log.Log.Info("Servers gracefully stopped")
}
