package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/apex/log"
	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	"github.com/open-collaboration/server/httpUtils"
	"net/http"
	"reflect"
	"runtime"
	"strings"
)

type RouteResponse struct {
	Status int
	Json   map[string]interface{}
}

// Sets up all routes in the application.
func SetupRoutes(router *mux.Router, providers []interface{}) {
	for _, provider := range providers {
		providerType := reflect.TypeOf(provider)
		if providerType.Kind() != reflect.Ptr {
			log.
				WithField("provider", providerType.Name()).
				Warn("By-value provider detected, this is allowed but not recommended. Consider passing the provider as a pointer.")
		}
	}

	router.HandleFunc("/users", createRouteHandler(RouteRegisterUser, providers)).Methods("POST")
	router.HandleFunc("/login", createRouteHandler(RouteAuthenticateUser, providers)).Methods("POST")
	router.HandleFunc("/projects", createRouteHandler(RouteCreateProject, providers)).Methods("POST")
	router.HandleFunc("/projects", createRouteHandler(RouteListProjects, providers)).Methods("GET")

	err := router.Walk(logRoute)
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

	handlerType := reflect.TypeOf(handler)
	if handlerType.Kind() != reflect.Func {
		panic("Supplied handler is not a func!")
	}

	handlerPath := runtime.FuncForPC(reflect.ValueOf(handler).Pointer()).Name()

	handlerPathParts := strings.Split(handlerPath, ".")
	handlerName := handlerPathParts[len(handlerPathParts)-1]

	checkReturnTypes(handlerType, handlerName)

	return func(writer http.ResponseWriter, request *http.Request) {
		// Populate two arrays with the underlying types of each value in providers
		// and whether that value is a pointer or not.
		providerTypes := make([]reflect.Type, len(providers))
		providerTypesPtrs := make([]bool, len(providers))
		for i, provider := range providers {
			providerType, isPtr := getType(reflect.TypeOf(provider))

			providerTypes[i] = providerType
			providerTypesPtrs[i] = isPtr
		}

		// Create an array of arguments that will be used to call handler.
		// These arguments are found by matching the handler parameter types
		// against the provider types, when there is a match the respective provider
		// will be used as argument for the respective parameter.
		handlerParamCount := handlerType.NumIn()
		handlerArgs := make([]reflect.Value, handlerParamCount)
		for i := 0; i < handlerParamCount; i++ {
			paramType, isPtr := getType(handlerType.In(i))

			if paramType.Implements(reflect.TypeOf((*http.ResponseWriter)(nil)).Elem()) {
				if isPtr {
					panic(fmt.Sprintf("Handler %s: Cannot pass *http.ResponseWriter (pointer). Use http.Request (value) instead.", handlerName))
				} else {
					handlerArgs[i] = reflect.ValueOf(writer)
				}

				continue
			}

			if paramType == reflect.TypeOf(http.Request{}) {
				if isPtr {
					handlerArgs[i] = reflect.ValueOf(request)
				} else {
					panic(fmt.Sprintf("Handler %s: Cannot pass http.Request (value). Use *http.Request (pointer) instead.", handlerName))
				}

				continue
			}

			for providerIdx, provider := range providers {
				// Check if param and provider are of the same type
				// (and whether or not they're both pointers)
				if isPtr == providerTypesPtrs[providerIdx] && paramType == providerTypes[providerIdx] {
					handlerArgs[i] = reflect.ValueOf(provider)

					break
				}
			}

			if !handlerArgs[i].IsValid() {
				panic(fmt.Sprintf("The provider %s was not found for parameter %d of handler %s", paramType.Name(), i, handlerPath))
			}
		}

		ctx := context.Background()
		requestId, err := uuid.NewV4()
		if err != nil {
			log.WithError(err).Error("Failed to generate a request id.")
			writer.WriteHeader(500)

			return
		}

		logger := log.WithFields(log.Fields{
			"requestId": requestId,
		})
		log.NewContext(ctx, logger)

		// Call the handler
		returnValues := reflect.ValueOf(handler).Call(handlerArgs)
		var routeErr error = nil

		// We know the returned value is error because we checked it earlier
		if !returnValues[0].IsNil() {
			routeErr = returnValues[0].Interface().(error)
		}

		// Handle the error returned by the handler, if any
		if routeErr != nil {
			handleRouteError(writer, ctx, routeErr)
		}
	}
}

func getType(t reflect.Type) (reflect.Type, bool) {
	paramType := t
	isPtr := paramType.Kind() == reflect.Ptr
	if isPtr {
		paramType = paramType.Elem()

		if paramType.Kind() == reflect.Ptr {
			panic("Double pointers are not supported by the Dependency Injection system.")
		}
	}

	return paramType, isPtr
}

func checkReturnTypes(handlerType reflect.Type, handlerName string) {
	msg := fmt.Sprintf("%s: Handlers should have an error return type", handlerName)

	returnValueCount := handlerType.NumOut()
	if returnValueCount != 1 {
		panic(msg)
	}

	firstRetVal := handlerType.Out(0)
	if firstRetVal != reflect.TypeOf((*error)(nil)).Elem() {
		panic(msg)
	}
}

func handleRouteError(writer http.ResponseWriter, ctx context.Context, routeErr error) {
	logger := log.FromContext(ctx)

	code := "unknown-error"
	details := map[string]interface{}{}

	switch e := routeErr.(type) {
	case *json.SyntaxError:
		code = "json-syntax-error"
		details["offset"] = fmt.Sprintf("%d", e.Offset)

	case *json.UnmarshalTypeError:
		code = "json-type-error"
		details["field"] = e.Field
	}

	err := httpUtils.WriteJson(writer, ctx, map[string]interface{}{
		"code":    code,
		"details": details,
	})
	if err != nil {
		logger.WithError(err).Error("Failed to write error response")
		writer.WriteHeader(500)
	}

	writer.WriteHeader(400)
}

func logRoute(route *mux.Route, _ *mux.Router, _ []*mux.Route) error {

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
