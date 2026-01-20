package middleware

import (
	"net/http"
	"strings"

	helper "refina-transaction/internal/utils"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(ctx *gin.Context) {
		jwt := ctx.GetHeader("Authorization")
		if jwt == "" {
			if ctx.Request.URL.Path != "/" {
				ctx.JSON(http.StatusUnauthorized, gin.H{
					"statusCode": 401,
					"status":     false,
					"error":      "Unauthorized",
				})
				ctx.Abort()
				return
			} else {
				ctx.Redirect(http.StatusSeeOther, "/login")
				ctx.Abort()
				return
			}
		}

		tokenParts := strings.Split(jwt, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"statusCode": 401,
				"status":     false,
				"error":      "Invalid Authorization format",
			})
			ctx.Abort()
			return
		}

		token := tokenParts[1]
		userData, err := helper.VerifyToken(token)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"statusCode": 401,
				"status":     false,
				"error":      "Unauthorized",
			})
			ctx.Abort()
			return
		}

		ctx.Set("user_data", userData)
		ctx.Next()
	})
}
