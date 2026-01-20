package routes

import (
	"refina-transaction/interface/http/handler"
	"refina-transaction/interface/http/middleware"
	"refina-transaction/internal/repository"
	"refina-transaction/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func CategoryRoutes(version *gin.Engine, db *gorm.DB) {
	txManager := repository.NewTxManager(db)
	categoryRepo := repository.NewCategoryRepository(db)

	categoryServ := service.NewCategoriesService(txManager, categoryRepo)
	categoryHandler := handler.NewCategoryHandler(categoryServ)

	category := version.Group("/categories")
	category.Use(middleware.AuthMiddleware())

	category.GET("", categoryHandler.GetAllCategories)
	category.GET("/:id", categoryHandler.GetCategoryByID)
	category.GET("/type/:type", categoryHandler.GetCategoriesByType)
	category.POST("", categoryHandler.CreateCategory)
	category.PUT("/:id", categoryHandler.UpdateCategory)
	category.DELETE("/:id", categoryHandler.DeleteCategory)
}
