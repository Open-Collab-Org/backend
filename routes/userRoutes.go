package routes

import (
	"context"
	"github.com/dgrijalva/jwt-go"
	"github.com/open-collaboration/server/dtos"
	"github.com/open-collaboration/server/services"
	"github.com/open-collaboration/server/utils"
	"net/http"
	"os"
)

// Registers a user.
// Accepts a dtos.NewUserDto as Json.
func RouteRegisterUser(writer http.ResponseWriter, request *http.Request, usersService *services.UsersService) error {
	dto := dtos.NewUserDto{}
	err := utils.ReadJson(request, context.TODO(), &dto)
	if err != nil {
		return err
	}

	err = usersService.CreateUser(context.TODO(), dto)
	if err != nil {
		return err
	}

	writer.WriteHeader(201)

	return nil
}

func RouteAuthenticateUser(writer http.ResponseWriter, request *http.Request, usersService *services.UsersService) error {
	dto := dtos.LoginDto{}
	err := utils.ReadJson(request, context.TODO(), &dto)
	if err != nil {
		return err
	}

	user, err := usersService.AuthenticateUser(context.TODO(), dto)
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

		err = utils.WriteJson(writer, context.TODO(), http.StatusOK, authenticatedUser)
		if err != nil {
			return err
		}
	} else {
		writer.WriteHeader(401)
	}

	return nil
}
