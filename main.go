package main

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/open-collaboration/server/dtos"
	"github.com/open-collaboration/server/migrations"
	"github.com/open-collaboration/server/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"net/http"
)

func main() {
	server := gin.Default()

	dsn := "host=localhost user=root password=changeme dbname=opencollab port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return
	}

	migration := migrations.GetMigration(db)
	err = migration.Migrate()
	if err != nil {
		return
	}

	setupRoutes(server, db)

	err = server.Run()
	if err != nil {
		return
	}
}

func setupRoutes(server *gin.Engine, db *gorm.DB) {
	server.POST("/users", func(c *gin.Context) {
		registerUser(c, db)
	})
}

func registerUser(c *gin.Context, db *gorm.DB) {
	newUser := dtos.NewUserDto{}
	err := c.ShouldBind(&newUser)

	if err != nil {
		ginErr, isGinErr := err.(gin.Error)
		_, isValidationErr := err.(validator.ValidationErrors)

		if isValidationErr || (isGinErr && ginErr.IsType(gin.ErrorTypeBind)) {
			c.AbortWithStatus(http.StatusBadRequest)
		} else {
			c.AbortWithStatus(http.StatusInternalServerError)
		}

		return
	}

	err = models.CreateUser(db, newUser)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)

		return
	}

	c.Status(201)
}