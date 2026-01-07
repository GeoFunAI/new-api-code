package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func NiceCode() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestId := c.GetHeader("X-Request-ID")
		if requestId == "" {
			// 返回错误
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "plz use right domain to access this api",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}