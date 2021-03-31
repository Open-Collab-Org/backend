package main

import (
	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/open-collaboration/server/migrations"
	"github.com/open-collaboration/server/projects"
	"github.com/open-collaboration/server/users"
	apex_gin "github.com/thedanielforum/apex-gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"net/http"
	"os"
)

type ErrorDto struct {
	ErrorCode    string      `json:"errorCode"`
	ErrorDetails interface{} `json:"errorDetails"`
}

func main() {

	// Setup logger
	log.SetHandler(cli.New(os.Stdout))
	log.SetLevel(log.DebugLevel)

	// Load env variables
	err := godotenv.Load()
	if err != nil {
		log.WithError(err).Error("Failed to load environment variables.")
		panic(err)
	}

	// Setup db connection
	dsn := "host=localhost user=root password=changeme dbname=opencollab port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.WithError(err).Error("Failed to connect to database.")
		panic(err)
	}

	// Run db migrations
	migration := migrations.GetMigration(db)
	err = migration.Migrate()
	if err != nil {
		log.WithError(err).Error("Failed to run database migrations.")
		panic(err)
	}

	// Setup server
	server := gin.New()

	server.Use(cors.Default())
	server.Use(apex_gin.Handler("request"))

	setupRoutes(server, db)

	// Start server
	addr := os.Getenv("HOST")
	err = server.Run(addr)
	if err != nil {
		log.WithError(err).Error("Failed to start the server.")
		panic(err)
	}
}

type routeHandler = func(*gin.Context, *gorm.DB) error

// Sets up all routes in the application.
func setupRoutes(server *gin.Engine, db *gorm.DB) {
	server.POST("/users", createRouteHandler(users.RouteRegisterUser, db))
	server.POST("/login", createRouteHandler(users.RouteAuthenticateUser, db))
	server.POST("/projects", createRouteHandler(projects.RouteCreateProject, db))
}

// This method is used to create gin route handlers with a few conveniences.
// It returns a gin route handler that calls the handler you supplied with a
// database reference and automatic error handling. All you have to do is
// supply a routeHandler and the rest will be taken care of for you.
func createRouteHandler(handler routeHandler, db *gorm.DB) func(*gin.Context) {
	return func(c *gin.Context) {
		err := handler(c, db)
		if err != nil {
			ginErr, isGinErr := err.(gin.Error)
			validationErr, isValidationErr := err.(validator.ValidationErrors)

			if isValidationErr || (isGinErr && ginErr.IsType(gin.ErrorTypeBind)) {
				errorsMap := make(map[string]string)

				for _, fieldErr := range validationErr {
					errorsMap[fieldErr.Field()] = fieldErr.Tag() + "=" + fieldErr.Param()
				}

				c.JSON(http.StatusBadRequest, &ErrorDto{
					ErrorCode:    "validation-error",
					ErrorDetails: interface{}(errorsMap),
				})
			} else {
				log.WithError(err).Error("Internal Error")
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}
	}
}
