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

type HandlerMetadata struct {
	handler     interface{}
	handlerType reflect.Type
	handlerName string
	handlerPath string
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

	handlerMetadata := HandlerMetadata{
		handler,
		handlerType,
		handlerName,
		handlerPath,
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

		rsProviders := []interface{}{
			ctx,
			&writer,
			request,
		}

		err = callHandler(handlerMetadata, providers, rsProviders)
		if err != nil {
			handleRouteError(writer, ctx, err)
		}
	}
}

// Calls a handler
// TODO: document this properly
// Note: if you want to make a provider accessible as an interface, you
// have to pass a pointer to the interface as provider. Example:
// Assume myProvider implements MyIface
// If we want the handler to receive a MyIface parameter,
// rsProviders has to be []interface{}{ &MyIface(myProvider) }.
func callHandler(
	handlerMetadata HandlerMetadata,
	providers []interface{},
	rsProviders []interface{},
) error {
	// Populate two arrays with the underlying types of each value in providers
	// and whether that value is a pointer or not.
	providerTypes := make([]reflect.Type, len(providers))
	providerTypesPtrs := make([]bool, len(providers))
	for i, provider := range providers {
		providerType, isPtr := getType(reflect.TypeOf(provider))

		providerTypes[i] = providerType
		providerTypesPtrs[i] = isPtr
	}

	rsProviderTypes := make([]reflect.Type, len(rsProviders))
	rsProviderTypesPtrs := make([]bool, len(rsProviders))
	for i, rsProvider := range rsProviders {
		rsProviderType, isPtr := getType(reflect.TypeOf(rsProvider))

		rsProviderTypes[i] = rsProviderType
		rsProviderTypesPtrs[i] = isPtr
	}

	// Create an array of arguments that will be used to call handler.
	// These arguments are found by matching the handler parameter types
	// against the provider types, when there is a match the respective provider
	// will be used as argument for the respective parameter.
	handlerParamCount := handlerMetadata.handlerType.NumIn()
	handlerArgs := make([]reflect.Value, handlerParamCount)
paramLoop:
	for i := 0; i < handlerParamCount; i++ {
		paramType, isPtr := getType(handlerMetadata.handlerType.In(i))

		for rsProviderIdx, rsProvider := range rsProviders {
			rsProviderType := rsProviderTypes[rsProviderIdx]
			rsProviderIsPtr := rsProviderTypesPtrs[rsProviderIdx]

			rsProviderIsIface := rsProviderType.Kind() == reflect.Interface

			if rsProviderIsIface && paramType == rsProviderType {
				// If we're here, rsProviderType is of kind interface. This tells us that rsProvider
				// is a pointer for sure, since the only way to get an interface type is through
				// the Elem() of the pointer-to-interface type.
				// This means that if the handler is expecting an interface argument, we have to
				// dereference the value of rsProvider (with reflect.Indirect) before passing it to the handler.
				// If the handler is expecting a pointer to the interface, we can just pass the value of rsProvider,
				// since it already is a pointer to the interface.
				if isPtr {
					handlerArgs[i] = reflect.ValueOf(rsProvider)
				} else {
					handlerArgs[i] = reflect.Indirect(reflect.ValueOf(rsProvider))
				}

				continue
			}

			if rsProviderType == paramType && rsProviderIsPtr == isPtr {
				handlerArgs[i] = reflect.ValueOf(rsProvider)

				continue paramLoop
			}
		}

		for providerIdx, provider := range providers {
			// Check if param and provider are of the same type
			// (and whether or not they're both pointers)
			if isPtr == providerTypesPtrs[providerIdx] && paramType == providerTypes[providerIdx] {
				handlerArgs[i] = reflect.ValueOf(provider)

				continue paramLoop
			}
		}

		if !handlerArgs[i].IsValid() {
			panic(
				fmt.Sprintf("The provider %s was not found for parameter %d of handler %s",
					paramType.Name(),
					i,
					handlerMetadata.handlerPath,
				),
			)
		}
	}

	// Call the handler
	returnValues := reflect.ValueOf(handlerMetadata.handler).Call(handlerArgs)
	routeErr := error(nil)

	// We know the returned value is error because we checked it earlier
	if !returnValues[0].IsNil() {
		routeErr = returnValues[0].Interface().(error)
	}

	return routeErr
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
