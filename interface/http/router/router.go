package router

import (
	"net/http"

	"refina-transaction/config/db"
	"refina-transaction/config/env"
	"refina-transaction/config/miniofs"
	"refina-transaction/interface/http/middleware"
	"refina-transaction/interface/http/routes"

	"github.com/gin-gonic/gin"
)

func SetupHTTPServer() *http.Server {
	router := gin.Default()

	router.Use(middleware.CORSMiddleware(), middleware.GinMiddleware())

	router.GET("test", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello World",
		})
	})

	routes.TransactionRoutes(router, db.DB, miniofs.MinioClient)
	routes.CategoryRoutes(router, db.DB)

	return &http.Server{
		Addr:    ":" + env.Cfg.Server.HTTPPort,
		Handler: router,
	}
}
