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
	grpcserver "refina-transaction/interface/grpc/server"
	"refina-transaction/interface/http/router"
	"refina-transaction/interface/queue"
	"refina-transaction/internal/repository"
	"refina-transaction/internal/service"
)

func init() {
	log.SetupLogger()

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
}

func main() {
	defer log.Info("Refina API stopped")

	log.Info("Setup Database Connection Start")
	dbInstance := db.GetInstance(env.Cfg.Database)
	log.Info("Setup Database Connection Success")

	log.Info("Setup MinIO Connection Start")
	minioInstance := miniofs.SetupMinio(env.Cfg.Minio) // Initialize MinIO connection
	log.Info("Setup MinIO Connection Success")

	log.Info("Setup RabbitMQ Connection Start")
	queueInstance := queue.GetInstance(env.Cfg.RabbitMQ)
	log.Info("Setup RabbitMQ Connection Success")

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup Outbox Publisher
	log.Info("Setup Outbox Publisher Start")
	outboxRepo := repository.NewOutboxRepository(dbInstance.GetDB())
	outboxPublisher := service.NewOutboxPublisher(outboxRepo, queueInstance)

	// Start outbox publisher worker
	go outboxPublisher.Start(ctx)

	// Start cleanup job (optional)
	go outboxPublisher.StartCleanupJob(ctx)
	log.Info("Outbox Publisher started successfully")

	log.Info("Starting Refina API...")

	// Set up the gRPC client
	grpcManager := client.GetManager()
	err := grpcManager.SetupGRPCClient()
	if err != nil {
		log.Log.Fatalf("Failed to set up gRPC client: %v", err)
	}
	log.Info("gRPC client setup successfully")

	// Set up the HTTP server
	httpServer := router.SetupHTTPServer(dbInstance, minioInstance, queueInstance)
	if httpServer != nil {
		go func() {
			if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Log.Fatalf("Failed to start HTTP server: %s\n", err)
			}
		}()
		log.Info("Starting HTTP server successfully. Running in port: " + env.Cfg.Server.HTTPPort)
	}

	// Set up the gRPC server
	grpcServer, lis, err := grpcserver.SetupGRPCServer(dbInstance)
	if err != nil {
		log.Log.Fatalf("Failed to set up gRPC server: %v", err)
	}
	if grpcServer != nil && lis != nil {
		go func() {
			if err := grpcServer.Serve(*lis); err != nil {
				log.Log.Fatalf("Failed to serve gRPC: %v", err)
			}
		}()
		log.Info("Starting gRPC server successfully. Listening on port: " + env.Cfg.Server.GRPCPort)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Log.Info("Shutting down servers...")

	// Cancel context to stop outbox publisher
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Log.Fatalf("Failed to shutdown HTTP server: %v", err)
	}

	grpcServer.GracefulStop()

	if err := grpcManager.Shutdown(ctx); err != nil {
		log.Log.Fatalf("Failed to shutdown gRPC clients: %v", err)
	}

	log.Log.Info("Servers gracefully stopped")
}
