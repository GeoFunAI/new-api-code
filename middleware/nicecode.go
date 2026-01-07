package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func NiceCode() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		method := c.Request.Method

		// 只对 AI 模型调用路径进行验证
		needCheck := false

		// /v1/* AI 模型路径
		if strings.HasPrefix(path, "/v1/") {
			// 排除模型列表查询 (GET 请求)
			if path == "/v1/models" || strings.HasPrefix(path, "/v1/models/") {
				if method == "GET" {
					needCheck = false
				} else {
					// POST /v1/models/* 是 Gemini API，需要检查
					needCheck = true
				}
			} else {
				// 其他 /v1/* 路径都是 AI 模型调用
				needCheck = true
			}
		}

		// /v1beta/* Gemini AI 模型路径
		if strings.HasPrefix(path, "/v1beta/") {
			// 排除模型列表查询 (GET 请求)
			if strings.HasPrefix(path, "/v1beta/models") || strings.HasPrefix(path, "/v1beta/openai/models") {
				if method == "GET" {
					needCheck = false
				} else {
					// POST 是 AI 调用
					needCheck = true
				}
			}
		}

		// 排除的路径 (非 AI 模型调用)
		// /pg/* - Playground
		// /mj/* - Midjourney
		// /suno/* - Suno
		if strings.HasPrefix(path, "/pg/") ||
			strings.HasPrefix(path, "/mj/") ||
			strings.HasPrefix(path, "/suno/") {
			needCheck = false
		}

		// 如果需要检查，验证 X-Request-ID
		if needCheck {
			requestId := c.GetHeader("X-Request-ID")
			if requestId == "" {
				// 返回错误
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "plz use right domain to access this api",
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}