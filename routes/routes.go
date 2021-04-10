package routes

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ItsaMeTuni/godi"
	"github.com/apex/log"
	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	"github.com/open-collaboration/server/httpUtils"
	"net/http"
)

type RouteResponse struct {
	Status int
	Json   map[string]interface{}
}

// Sets up all routes in the application.
func SetupRoutes(router *mux.Router, providers []interface{}) {
	router.HandleFunc("/users", createRouteHandler(RouteRegisterUser, providers)).Methods("POST")
	router.HandleFunc("/login", createRouteHandler(RouteAuthenticateUser, providers)).Methods("POST")
	router.HandleFunc("/projects", createRouteHandler(RouteCreateProject, providers)).Methods("POST")
	router.HandleFunc("/projects", createRouteHandler(RouteListProjects, providers)).Methods("GET")
	router.HandleFunc("/projects/{projectId}", createRouteHandler(RouteGetProject, providers)).Methods("GET")

	err := router.Walk(logRouteDeclaration)
	if err != nil {
		log.WithError(err).Error("Failed to log routes")
		panic("Failed to log routes")
	}
}

// This method is used to create gin route handlers with a few conveniences.
// It returns a gin route handler that calls the handler you supplied with a
// database reference and automatic error handling. All you have to do is
// supply a routeHandler and the rest will be taken care of for you.
func createRouteHandler(handler interface{}, providers []interface{}) func(http.ResponseWriter, *http.Request) {

	err := godi.AssertFn(handler, []interface{}{errors.New("")})
	if err != nil {
		panic(err)
	}

	return func(writer http.ResponseWriter, request *http.Request) {
		requestId, err := uuid.NewV4()
		if err != nil {
			log.WithError(err).Error("Failed to generate a request id.")
			writer.WriteHeader(500)

			return
		}

		ctx := context.Background()
		logger := log.WithFields(log.Fields{
			"requestId": requestId,
		})
		ctx = log.NewContext(ctx, logger)

		logRouteExecution(request, ctx)

		reqProviders := append(
			godi.Providers{
				ctx,
				writer,
				request,
			},
			providers...,
		)

		returnValues, err := godi.Inject(handler, reqProviders)
		if err != nil {
			logger.WithError(err).Error("Failed to call handler and inject dependencies on it.")
			return
		}

		// Handlers should return an error (or nil), as we asserted above
		if !returnValues[0].IsNil() {
			e := returnValues[0].Interface().(error)
			handleRouteError(writer, ctx, e)
		}
	}
}

// Handle an error that was returned by a route.
//
// Json syntax and unmarshalling errors return to the client a
// 400 response with an error description.
// All other errors return a 500 without a body.
func handleRouteError(writer http.ResponseWriter, ctx context.Context, routeErr error) {
	logger := log.FromContext(ctx)

	code := "unknown-error"
	details := map[string]interface{}{}

	switch e := routeErr.(type) {
	default:
		writer.WriteHeader(500)

	case *json.SyntaxError:
		code = "json-syntax-error"
		details["offset"] = fmt.Sprintf("%d", e.Offset)
		writer.WriteHeader(400)

	case *json.UnmarshalTypeError:
		code = "json-type-error"
		details["field"] = e.Field
		writer.WriteHeader(400)

	}

	err := httpUtils.WriteJson(writer, ctx, map[string]interface{}{
		"code":    code,
		"details": details,
	})
	if err != nil {
		logger.WithError(err).Error("Failed to write error response")
		writer.WriteHeader(500)
	}
}

// Log a route declaration in the format "Route: [METHOD] <path>".
// This basically just to inform that a route has been properly
// configured.
func logRouteDeclaration(route *mux.Route, _ *mux.Router, _ []*mux.Route) error {

	methods, err := route.GetMethods()
	if err != nil {
		return err
	}

	pathTemplate, err := route.GetPathTemplate()
	if err != nil {
		return err
	}

	log.Infof("Route: %s %s", methods, pathTemplate)

	return nil
}

// Log the start of the execution of a route handler.
func logRouteExecution(request *http.Request, ctx context.Context) {
	logger := log.FromContext(ctx)

	logger.Infof("Processing %s request to %s", request.Method, request.URL)
}
