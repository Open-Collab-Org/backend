package users

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"os"
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

func RouteAuthenticateUser(c *gin.Context, db *gorm.DB) error {
	authUser := LoginDto{}
	err := c.ShouldBind(&authUser)

	if err != nil {
		return err
	}

	user, err := AuthenticateUser(db, authUser)
	if err != nil {
		return err
	}

	if user != nil {
		claims := jwt.MapClaims{"userId": user.ID}

		jwtKey := os.Getenv("JWT_SIGNING_KEY")

		token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(jwtKey))
		if err != nil {
			return err
		}

		authenticatedUser := AuthenticatedUserDto{
			Token: token,
			User: UserDataDto{
				Username: user.Username,
				Email:    user.Email,
			},
		}

		c.JSON(200, authenticatedUser)
	} else {
		c.AbortWithStatus(401)
	}

	return nil
}
