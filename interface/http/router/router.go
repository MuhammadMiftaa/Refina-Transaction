package router

import (
	"refina-transaction/config/db"
	"refina-transaction/config/miniofs"
	"refina-transaction/interface/http/middleware"
	"refina-transaction/interface/http/routes"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	router := gin.Default()

	router.Use(middleware.CORSMiddleware(), middleware.GinMiddleware())

	router.GET("test", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello World",
		})
	})

	routes.TransactionRoutes(router, db.DB, miniofs.MinioClient)
	routes.CategoryRoutes(router, db.DB)

	return router
}
