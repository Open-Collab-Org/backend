package users

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Registers a user.
// Accepts a dtos.NewUserDto as body.
func RouteRegisterUser(c *gin.Context, db *gorm.DB) error {
	newUser := NewUserDto{}
	err := c.ShouldBind(&newUser)

	if err != nil {
		return err
	}

	err = CreateUser(db, newUser)
	if err != nil {
		return err
	}

	c.Status(201)

	return nil
}
