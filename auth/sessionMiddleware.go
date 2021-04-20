package auth

import (
	"context"
	"errors"
	"github.com/apex/log"
	"github.com/gorilla/mux"
	"net/http"
)

type Session struct {
	token  string
	userId uint
}

var ErrUnauthenticated = errors.New("unauthenticated")

// Checks the incoming request for a session token. If the session token
// exists and is valid, a session is added to the request's context.
// You can get the session with
//	r.Context().Value(Session{})
func SessionMiddleware(authService *Service) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := log.FromContext(r.Context())

			logger.Debug("Checking request session token")

			ctx := r.Context()

			session, err := getSessionFromRequest(r, authService)
			if err != nil {
				if !errors.Is(err, http.ErrNoCookie) && !errors.Is(err, ErrInvalidSessionToken) {
					logger.WithError(err).Error("Failed to get request's session")
					w.WriteHeader(http.StatusInternalServerError)

					return
				}
			} else {
				logger.Debug("Session token found")
				ctx = context.WithValue(r.Context(), Session{}, session)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Get a Session from the request. Returns an http.ErrNoCookie if the cookie is not set
// or services.ErrInvalidSessionToken if the session token is invalid. Otherwise a Session
// is returned.
func getSessionFromRequest(r *http.Request, authService *Service) (Session, error) {
	sessionToken, err := r.Cookie("sessionToken")
	if err != nil {
		return Session{}, err
	}

	userId, err := authService.AuthenticateSession(r.Context(), sessionToken.Value)
	if err != nil {
		return Session{}, err
	}

	return Session{
		token:  sessionToken.Value,
		userId: userId,
	}, nil
}

func CheckSession(r *http.Request) (Session, error) {
	session := r.Context().Value(Session{})
	if session == nil {
		return Session{}, ErrUnauthenticated
	}

	return session.(Session), nil
}
