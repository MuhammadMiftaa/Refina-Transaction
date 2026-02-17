package router

import (
	"net/http"

	"refina-transaction/config/db"
	"refina-transaction/config/env"
	"refina-transaction/config/miniofs"
	"refina-transaction/interface/http/middleware"
	"refina-transaction/interface/http/routes"
	"refina-transaction/interface/queue"

	"github.com/gin-gonic/gin"
)

func SetupHTTPServer(dbInstance db.DatabaseClient, minioInstance *miniofs.MinIOManager, queueInstance queue.RabbitMQClient) *http.Server {
	router := gin.Default()

	router.Use(middleware.CORSMiddleware(), middleware.GinMiddleware())

	router.GET("test", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello World",
		})
	})

	routes.TransactionRoutes(router, dbInstance.GetDB(), minioInstance)
	routes.CategoryRoutes(router, dbInstance.GetDB())

	return &http.Server{
		Addr:    ":" + env.Cfg.Server.HTTPPort,
		Handler: router,
	}
}
