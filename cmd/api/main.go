package main

import (
	"refina-transaction/config/db"
	"refina-transaction/config/env"
	"refina-transaction/config/log"
	"refina-transaction/config/miniofs"
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

	r := router.SetupRouter() // Set up the HTTP router
	r.Run(":" + env.Cfg.Server.Port)
	log.Info("Starting HTTP server on port " + env.Cfg.Server.Port)
}
