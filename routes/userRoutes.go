package routes

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/open-collaboration/server/dtos"
	"github.com/open-collaboration/server/services"
	"os"
)

// Registers a user.
// Accepts a dtos.NewUserDto as Json.
func RouteRegisterUser(c *gin.Context, usersService *services.UsersService) error {
	newUser := dtos.NewUserDto{}
	err := c.ShouldBind(&newUser)

	if err != nil {
		return err
	}

	err = usersService.CreateUser(newUser)
	if err != nil {
		return err
	}

	c.Status(201)

	return nil
}

func RouteAuthenticateUser(c *gin.Context, usersService *services.UsersService) error {
	authUser := dtos.LoginDto{}
	err := c.ShouldBind(&authUser)

	if err != nil {
		return err
	}

	user, err := usersService.AuthenticateUser(authUser)
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

		authenticatedUser := dtos.AuthenticatedUserDto{
			Token: token,
			User: dtos.UserDataDto{
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
