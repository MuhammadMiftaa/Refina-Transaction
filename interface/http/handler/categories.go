package handler

import (
	"net/http"

	"refina-transaction/internal/service"
	"refina-transaction/internal/types/dto"

	"github.com/gin-gonic/gin"
)

type CategoryHandler struct {
	categoryServ service.CategoriesService
}

func NewCategoryHandler(categoryServ service.CategoriesService) *CategoryHandler {
	return &CategoryHandler{categoryServ}
}

func (categoryHandler *CategoryHandler) GetAllCategories(c *gin.Context) {
	ctx := c.Request.Context()

	categories, err := categoryHandler.categoryServ.GetAllCategories(ctx)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"statusCode": 400,
			"message":    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     true,
		"statusCode": 200,
		"message":    "Get all categories data",
		"data":       categories,
	})
}

func (categoryHandler *CategoryHandler) GetCategoryByID(c *gin.Context) {
	ctx := c.Request.Context()

	id := c.Param("id")

	category, err := categoryHandler.categoryServ.GetCategoryByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"statusCode": 400,
			"message":    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     true,
		"statusCode": 200,
		"message":    "Get category data by ID",
		"data":       category,
	})
}

func (categoryHandler *CategoryHandler) GetCategoriesByType(c *gin.Context) {
	ctx := c.Request.Context()

	typeCategory := c.Param("type")

	categories, err := categoryHandler.categoryServ.GetCategoriesByType(ctx, typeCategory)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"statusCode": 400,
			"message":    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     true,
		"statusCode": 200,
		"message":    "Get categories data by type",
		"data":       categories,
	})
}

func (categoryHandler *CategoryHandler) CreateCategory(c *gin.Context) {
	ctx := c.Request.Context()

	var categoryRequest dto.CategoriesRequest
	if err := c.ShouldBindJSON(&categoryRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"statusCode": 400,
			"message":    "Invalid request",
		})
		return
	}

	category, err := categoryHandler.categoryServ.CreateCategory(ctx, categoryRequest)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"statusCode": 400,
			"message":    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     true,
		"statusCode": 200,
		"message":    "Category created successfully",
		"data":       category,
	})
}

func (categoryHandler *CategoryHandler) UpdateCategory(c *gin.Context) {
	ctx := c.Request.Context()

	id := c.Param("id")

	var categoryRequest dto.CategoriesRequest
	if err := c.ShouldBindJSON(&categoryRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"statusCode": 400,
			"message":    "Invalid request",
		})
		return
	}

	category, err := categoryHandler.categoryServ.UpdateCategory(ctx, id, categoryRequest)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"statusCode": 400,
			"message":    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     true,
		"statusCode": 200,
		"message":    "Category updated successfully",
		"data":       category,
	})
}

func (categoryHandler *CategoryHandler) DeleteCategory(c *gin.Context) {
	ctx := c.Request.Context()

	id := c.Param("id")

	category, err := categoryHandler.categoryServ.DeleteCategory(ctx, id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":     false,
			"statusCode": 400,
			"message":    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     true,
		"statusCode": 200,
		"message":    "Category deleted successfully",
		"data":       category,
	})
}
