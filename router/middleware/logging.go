package middleware

import (
	"context"
	"github.com/apex/log"
	"github.com/gofrs/uuid"
	"net/http"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestId, err := uuid.NewV4()
		if err != nil {
			log.WithError(err).Error("Failed to generate a request id.")
			w.WriteHeader(500)

			return
		}

		logger := log.WithFields(log.Fields{
			"requestId": requestId,
		})

		ctx := r.Context()
		ctx = log.NewContext(ctx, logger)

		logRouteExecution(r, ctx)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Log the start of the execution of a route handler.
func logRouteExecution(request *http.Request, ctx context.Context) {
	logger := log.FromContext(ctx)

	logger.Infof("Processing %s request to %s", request.Method, request.URL)
}
