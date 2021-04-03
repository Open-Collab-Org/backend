package logging

import (
	"context"
	"github.com/apex/log"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/open-collaboration/server/consts"
)

func LoggerFromCtx(ctx context.Context) *log.Entry {
	logger := ctx.Value(consts.LoggerKey).(*log.Entry)
	if logger == nil {
		logger = log.Log.WithFields(log.Fields{})
	}

	return logger
}

func LoggerMiddleware(c *gin.Context) {
	path := c.Request.URL.Path
	clientIP := c.ClientIP()
	method := c.Request.Method
	requestId, _ := uuid.NewV4()

	requestLogger := log.Log.WithField("request_id", requestId)

	requestLogger.WithFields(log.Fields{
		"client_ip": clientIP,
		"method":    method,
		"path":      path,
	}).Info("Begin request")

	c.Set(consts.LoggerKey, requestLogger)
	c.Set(consts.RequestIdKey, requestId)

	c.Next()

	statusCode := c.Writer.Status()

	requestLogger.WithFields(log.Fields{
		"request_id":  requestId,
		"status_code": statusCode,
	}).Info("End request")
}
