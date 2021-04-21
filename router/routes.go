package router

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ItsaMeTuni/godi"
	"github.com/apex/log"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/open-collaboration/server/auth"
	"github.com/open-collaboration/server/projects"
	"github.com/open-collaboration/server/router/middleware"
	"github.com/open-collaboration/server/users"
	"github.com/open-collaboration/server/utils"
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
	rootRouter.Use(middleware.CorsMiddleware)

	authService := getProvider(providers, (*auth.Service)(nil)).(auth.Service)
	rootRouter.Use(auth.SessionMiddleware(authService))

	// Setup routes
	rootRouter.HandleFunc("/users", createRouteHandler(users.RouteRegisterUser, providers)).Methods("POST")
	rootRouter.HandleFunc("/login", createRouteHandler(auth.RouteAuthenticateUser, providers)).Methods("POST")
	rootRouter.HandleFunc("/projects", createRouteHandler(projects.RouteListProjects, providers)).Methods("GET")
	rootRouter.HandleFunc("/projects", createRouteHandler(projects.RouteCreateProject, providers)).Methods("POST")
	rootRouter.HandleFunc("/projects/{projectId}", createRouteHandler(projects.RouteUpdateProject, providers)).Methods("POST")
	rootRouter.HandleFunc("/projects/{projectId}", createRouteHandler(projects.RouteGetProject, providers)).Methods("GET")

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

	if routeErr != nil {
		code := "unknown-error"
		details := map[string]interface{}{}

		logger.WithError(routeErr).Debug("Route resulted in error")

		var status int

		switch e := routeErr.(type) {
		default:
			if errors.Is(routeErr, auth.ErrUnauthenticated) {
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
			details[e.Field] = e.Error()
			status = http.StatusBadRequest

		case validator.ValidationErrors:
			code = "validation-error"
			for _, fieldError := range e {
				details[fieldError.StructField()] = fieldError.Tag()
			}
			status = http.StatusBadRequest
		}

		err := utils.WriteJson(writer, ctx, status, map[string]interface{}{
			"code":    code,
			"details": details,
		})
		if err != nil {
			logger.WithError(err).Error("Failed to write error response")
		}
	}
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

// Get a provider of the given type.
// Example:
//  providers := []interface{}{ &Foo{} }
//  foo := getProvider(providers, &Foo{}).(*Foo)
//
// To get a provider through an interface type, you should use
// a nil pointer to the interface as the provider type.
// Example:
//  type Foo struct {}
//  type Bar interface {} // Foo implements Bar
//  providers := []interface{}{ &Foo{} }
//  bar := getProvider(providers, (*Bar)(nil)).(Bar)
//
// Note: pointers to interfaces are handled differently. If you provide
// a pointer to an interface (e.g. *Bar), getProvider will try to find a
// provider that implements the interface (Bar), not a provider that is a pointer
// to a Bar (*Bar).
func getProvider(providers []interface{}, providerType interface{}) interface{} {
	pType := reflect.TypeOf(providerType)

	if pType.Kind() == reflect.Ptr && pType.Elem().Kind() == reflect.Interface {
		pType = pType.Elem()
	}

	for _, provider := range providers {
		if reflect.TypeOf(provider).AssignableTo(pType) {
			return provider
		}
	}

	return nil
}
