package httpapi

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"regexp"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	requestIDHeader = "X-Request-ID"
	requestIDKey    = "request_id"
)

var requestIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)

func requestMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		started := time.Now()
		id := c.GetHeader(requestIDHeader)
		if !requestIDPattern.MatchString(id) {
			generatedID, err := newRequestID()
			if err != nil {
				logger.Error("request id generation failed", zap.Error(err))
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})

				return
			}
			id = generatedID
		}

		c.Set(requestIDKey, id)
		c.Header(requestIDHeader, id)
		c.Next()

		route := c.FullPath()
		if route == "" {
			route = "unmatched"
		}

		logger.Info(
			"http request completed",
			zap.String("request_id", id),
			zap.String("method", c.Request.Method),
			zap.String("route", route),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("duration", time.Since(started)),
		)
	}
}

func recoveryMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			panicValue := recover()
			if panicValue == nil {
				return
			}

			logger.Error(
				"http panic recovered",
				zap.String("request_id", requestID(c)),
				zap.Any("panic", panicValue),
				zap.ByteString("stack", debug.Stack()),
			)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}()

		c.Next()
	}
}

func securityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		c.Header("Referrer-Policy", "no-referrer")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Next()
	}
}

func requestID(c *gin.Context) string {
	value, exists := c.Get(requestIDKey)
	if !exists {
		return ""
	}

	id, ok := value.(string)
	if !ok {
		return ""
	}

	return id
}

func newRequestID() (string, error) {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", err
	}

	return hex.EncodeToString(value[:]), nil
}
