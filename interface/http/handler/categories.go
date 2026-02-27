package handler

import (
	"net/http"

	"refina-transaction/config/log"
	"refina-transaction/internal/service"
	"refina-transaction/internal/types/dto"
	"refina-transaction/internal/utils/data"

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
	requestID, _ := c.Get(data.REQUEST_ID_LOCAL_KEY)

	categories, err := categoryHandler.categoryServ.GetAllCategories(ctx)
	if err != nil {
		log.Error("get_all_categories_failed", map[string]any{
			"service":    data.CategoryService,
			"request_id": requestID,
			"error":      err.Error(),
		})
		statusCode, message := mapServiceError(err)
		c.JSON(statusCode, gin.H{
			"statusCode": statusCode,
			"status":     false,
			"message":    message,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"statusCode": 200,
		"status":     true,
		"message":    "Get all categories data",
		"data":       categories,
	})
}

func (categoryHandler *CategoryHandler) GetCategoryByID(c *gin.Context) {
	ctx := c.Request.Context()
	requestID, _ := c.Get(data.REQUEST_ID_LOCAL_KEY)

	id := c.Param("id")

	category, err := categoryHandler.categoryServ.GetCategoryByID(ctx, id)
	if err != nil {
		log.Error("get_category_by_id_failed", map[string]any{
			"service":     data.CategoryService,
			"request_id":  requestID,
			"category_id": id,
			"error":       err.Error(),
		})
		statusCode, message := mapServiceError(err)
		c.JSON(statusCode, gin.H{
			"statusCode": statusCode,
			"status":     false,
			"message":    message,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"statusCode": 200,
		"status":     true,
		"message":    "Get category data by ID",
		"data":       category,
	})
}

func (categoryHandler *CategoryHandler) GetCategoriesByType(c *gin.Context) {
	ctx := c.Request.Context()
	requestID, _ := c.Get(data.REQUEST_ID_LOCAL_KEY)

	typeCategory := c.Param("type")

	categories, err := categoryHandler.categoryServ.GetCategoriesByType(ctx, typeCategory)
	if err != nil {
		log.Error("get_categories_by_type_failed", map[string]any{
			"service":       data.CategoryService,
			"request_id":    requestID,
			"category_type": typeCategory,
			"error":         err.Error(),
		})
		statusCode, message := mapServiceError(err)
		c.JSON(statusCode, gin.H{
			"statusCode": statusCode,
			"status":     false,
			"message":    message,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"statusCode": 200,
		"status":     true,
		"message":    "Get categories data by type",
		"data":       categories,
	})
}

func (categoryHandler *CategoryHandler) CreateCategory(c *gin.Context) {
	ctx := c.Request.Context()
	requestID, _ := c.Get(data.REQUEST_ID_LOCAL_KEY)

	var categoryRequest dto.CategoriesRequest
	if err := c.ShouldBindJSON(&categoryRequest); err != nil {
		log.Warn("create_category_bad_request", map[string]any{
			"service":    data.CategoryService,
			"request_id": requestID,
			"error":      err.Error(),
		})
		c.JSON(http.StatusBadRequest, gin.H{
			"statusCode": 400,
			"status":     false,
			"message":    "invalid request body",
		})
		return
	}

	category, err := categoryHandler.categoryServ.CreateCategory(ctx, categoryRequest)
	if err != nil {
		log.Error("create_category_failed", map[string]any{
			"service":    data.CategoryService,
			"request_id": requestID,
			"name":       categoryRequest.Name,
			"error":      err.Error(),
		})
		statusCode, message := mapServiceError(err)
		c.JSON(statusCode, gin.H{
			"statusCode": statusCode,
			"status":     false,
			"message":    message,
		})
		return
	}

	log.Info("category_created", map[string]any{
		"service":    data.CategoryService,
		"request_id": requestID,
		"name":       categoryRequest.Name,
	})

	c.JSON(http.StatusCreated, gin.H{
		"statusCode": 201,
		"status":     true,
		"message":    "Category created successfully",
		"data":       category,
	})
}

func (categoryHandler *CategoryHandler) UpdateCategory(c *gin.Context) {
	ctx := c.Request.Context()
	requestID, _ := c.Get(data.REQUEST_ID_LOCAL_KEY)

	id := c.Param("id")

	var categoryRequest dto.CategoriesRequest
	if err := c.ShouldBindJSON(&categoryRequest); err != nil {
		log.Warn("update_category_bad_request", map[string]any{
			"service":     data.CategoryService,
			"request_id":  requestID,
			"category_id": id,
			"error":       err.Error(),
		})
		c.JSON(http.StatusBadRequest, gin.H{
			"statusCode": 400,
			"status":     false,
			"message":    "invalid request body",
		})
		return
	}

	category, err := categoryHandler.categoryServ.UpdateCategory(ctx, id, categoryRequest)
	if err != nil {
		log.Error("update_category_failed", map[string]any{
			"service":     data.CategoryService,
			"request_id":  requestID,
			"category_id": id,
			"error":       err.Error(),
		})
		statusCode, message := mapServiceError(err)
		c.JSON(statusCode, gin.H{
			"statusCode": statusCode,
			"status":     false,
			"message":    message,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"statusCode": 200,
		"status":     true,
		"message":    "Category updated successfully",
		"data":       category,
	})
}

func (categoryHandler *CategoryHandler) DeleteCategory(c *gin.Context) {
	ctx := c.Request.Context()
	requestID, _ := c.Get(data.REQUEST_ID_LOCAL_KEY)

	id := c.Param("id")

	category, err := categoryHandler.categoryServ.DeleteCategory(ctx, id)
	if err != nil {
		log.Error("delete_category_failed", map[string]any{
			"service":     data.CategoryService,
			"request_id":  requestID,
			"category_id": id,
			"error":       err.Error(),
		})
		statusCode, message := mapServiceError(err)
		c.JSON(statusCode, gin.H{
			"statusCode": statusCode,
			"status":     false,
			"message":    message,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"statusCode": 200,
		"status":     true,
		"message":    "Category deleted successfully",
		"data":       category,
	})
}
