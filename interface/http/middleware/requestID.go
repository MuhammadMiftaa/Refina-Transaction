package middleware

import (
	"refina-transaction/internal/utils/data"

	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
)

func RequestIDMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestID := ctx.GetHeader(data.REQUEST_ID_HEADER)
		if requestID == "" {
			requestID = xid.New().String() + "-X"
		}

		ctx.Set(data.REQUEST_ID_LOCAL_KEY, requestID)
		ctx.Header(data.REQUEST_ID_HEADER, requestID)

		ctx.Next()
	}
}
