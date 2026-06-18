package middleware

import (
	"incubator-backend/pkg/logger"
	"time"

	"github.com/gin-gonic/gin"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		latency := time.Since(startTime)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()

		logger.Infof("[%s] %s %s - %d - %v",
			method,
			path,
			clientIP,
			statusCode,
			latency,
		)
	}
}
