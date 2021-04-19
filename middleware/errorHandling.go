package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/apex/log"
	"github.com/open-collaboration/server/utils"
	"net/http"
)

type RouteError struct {
	Err error
}

// Checks if there's an error in the request's context. If there
// is, an appropriate response will be returned based on the error found.
// If we don't know how to handle the error a 500 will be sent.
func ErrorHandlingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = context.WithValue(ctx, RouteError{}, &RouteError{})

		next.ServeHTTP(w, r.WithContext(ctx))

		logger := log.FromContext(ctx)

		routeErr := r.Context().Value((error)(nil)).(error)
		if routeErr != nil {
			code := "unknown-error"
			details := map[string]interface{}{}

			logger.WithError(routeErr).Debug("Route resulted in error")

			var status int

			switch e := routeErr.(type) {
			default:
				if errors.Is(routeErr, ErrUnauthenticated) {
					status = http.StatusUnauthorized
					code = "unauthenticated-error"
				} else {
					status = http.StatusInternalServerError
				}

			case *json.SyntaxError:
				code = "json-syntax-error"
				details["offset"] = fmt.Sprintf("%d", e.Offset)
				status = http.StatusBadRequest

			case *json.UnmarshalTypeError:
				code = "json-type-error"
				details["field"] = e.Field
				status = http.StatusBadRequest

			}

			err := utils.WriteJson(w, ctx, status, map[string]interface{}{
				"code":    code,
				"details": details,
			})
			if err != nil {
				logger.WithError(err).Error("Failed to write error response")
			}
		}
	})
}

func SetHandlerError(r *http.Request, err error) {
	routeErr := r.Context().Value(RouteError{}).(*RouteError)

	if routeErr.Err == nil {
		routeErr.Err = err
	} else {
		logger := log.FromContext(r.Context())
		logger.
			WithError(routeErr.Err).
			WithError(err).
			Error("More than one error occurred in a route handler")

		panic("More than one error occurred in a route handler")
	}
}
