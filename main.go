package main

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/open-collaboration/server/migrations"
	"github.com/open-collaboration/server/users"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"net/http"
)

func main() {
	server := gin.Default()

	dsn := "host=localhost user=root password=changeme dbname=opencollab port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	migration := migrations.GetMigration(db)
	err = migration.Migrate()
	if err != nil {
		panic(err)
	}

	setupRoutes(server, db)

	err = server.Run()
	if err != nil {
		panic(err)
	}
}

type routeHandler = func(*gin.Context, *gorm.DB) error

// Sets up all routes in the application.
func setupRoutes(server *gin.Engine, db *gorm.DB) {
	server.POST("/users", createRouteHandler(users.RouteRegisterUser, db))
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
			_, isValidationErr := err.(validator.ValidationErrors)

			if isValidationErr || (isGinErr && ginErr.IsType(gin.ErrorTypeBind)) {
				c.AbortWithStatus(http.StatusBadRequest)
			} else {
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}
	}
}
