package routes

import (
	"fmt"
	"github.com/apex/log"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/open-collaboration/server/dtos"
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
func SetupRoutes(server *gin.Engine, providers []interface{}) {
	server.POST("/users", createRouteHandler(RouteRegisterUser, providers))
	server.POST("/login", createRouteHandler(RouteAuthenticateUser, providers))
	server.POST("/projects", createRouteHandler(RouteCreateProject, providers))
	server.GET("/projects", createRouteHandler(RouteFetchProjects, providers))
}

// This method is used to create gin route handlers with a few conveniences.
// It returns a gin route handler that calls the handler you supplied with a
// database reference and automatic error handling. All you have to do is
// supply a routeHandler and the rest will be taken care of for you.
func createRouteHandler(handler interface{}, providers []interface{}) func(*gin.Context) {

	handlerType := reflect.TypeOf(handler)
	if handlerType.Kind() != reflect.Func {
		panic("Supplied handler is not a func!")
	}

	handlerPath := runtime.FuncForPC(reflect.ValueOf(handler).Pointer()).Name()

	handlerPathParts := strings.Split(handlerPath, ".")
	handlerName := handlerPathParts[len(handlerPathParts)-1]

	checkReturnTypes(handlerType, handlerName)

	return func(c *gin.Context) {
		providers = append(providers, c)

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

		// Call the handler
		returnValues := reflect.ValueOf(handler).Call(handlerArgs)
		var routeErr error = nil

		// We know the returned value is error because we checked it earlier
		if !returnValues[0].IsNil() {
			routeErr = returnValues[1].Interface().(error)
		}

		// Handle the error returned by the handler, if any
		if routeErr != nil {
			ginErr, isGinErr := routeErr.(gin.Error)
			validationErr, isValidationErr := routeErr.(validator.ValidationErrors)

			if isValidationErr || (isGinErr && ginErr.IsType(gin.ErrorTypeBind)) {
				errorsMap := make(map[string]string)

				for _, fieldErr := range validationErr {
					errorsMap[fieldErr.Field()] = fieldErr.Tag() + "=" + fieldErr.Param()
				}

				c.JSON(http.StatusBadRequest, &dtos.ErrorDto{
					ErrorCode:    "validation-error",
					ErrorDetails: interface{}(errorsMap),
				})
			} else {
				log.WithError(routeErr).Error("Internal Error")
				c.AbortWithStatus(http.StatusInternalServerError)
			}
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
