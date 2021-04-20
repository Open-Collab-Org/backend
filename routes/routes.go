package routes

import (
	"context"
	"errors"
	"github.com/ItsaMeTuni/godi"
	"github.com/apex/log"
	"github.com/gorilla/mux"
	"github.com/open-collaboration/server/middleware"
	"github.com/open-collaboration/server/services"
	"net/http"
	"reflect"
)

type RouteResponse struct {
	Status int
	Json   map[string]interface{}
}

// Sets up all routes in the application.
func SetupRoutes(providers []interface{}) *mux.Router {
	rootRouter := mux.NewRouter()

	rootRouter.Use(middleware.LoggingMiddleware)
	rootRouter.Use(middleware.ErrorHandlingMiddleware)
	rootRouter.Use(middleware.CorsMiddleware)

	authService := getProvider(providers, (*services.AuthService)(nil)).(*services.AuthService)
	rootRouter.Use(middleware.SessionMiddleware(authService))

	// Setup routes
	rootRouter.HandleFunc("/users", createRouteHandler(RouteRegisterUser, providers)).Methods("POST")
	rootRouter.HandleFunc("/login", createRouteHandler(RouteAuthenticateUser, providers)).Methods("POST")
	rootRouter.HandleFunc("/projects", createRouteHandler(RouteListProjects, providers)).Methods("GET")
	rootRouter.HandleFunc("/projects", createRouteHandler(RouteCreateProject, providers)).Methods("POST")
	rootRouter.HandleFunc("/projects/{projectId}", createRouteHandler(RouteGetProject, providers)).Methods("GET")

	// Swagger
	swaggerUi := http.FileServer(http.Dir("swagger-ui/"))
	rootRouter.
		PathPrefix("/swagger-ui").
		Handler(http.StripPrefix("/swagger-ui/", swaggerUi)).
		Methods("GET")

	// Log routes
	err := rootRouter.Walk(logRouteDeclaration)
	if err != nil {
		log.WithError(err).Error("Failed to log routes")
		panic("Failed to log routes")
	}

	return rootRouter
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
		ctx := request.Context()
		logger := log.FromContext(ctx)

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

}

// Log a route declaration in the format "Route: [METHOD] <path>".
// This basically just to inform that a route has been properly
// configured.
func logRouteDeclaration(route *mux.Route, _ *mux.Router, _ []*mux.Route) error {

	methods, err := route.GetMethods()
	if err != nil {
		return nil
	}

	pathTemplate, err := route.GetPathTemplate()
	if err != nil {
		return nil
	}

	log.Infof("Route: %s %s", methods, pathTemplate)

	return nil
}

func getProvider(providers []interface{}, providerType interface{}) interface{} {
	pType := reflect.TypeOf(providerType)

	for _, provider := range providers {
		if reflect.TypeOf(provider).AssignableTo(pType) {
			return provider
		}
	}

	return nil
}
