package routes

import (
	"context"
	"fmt"
	"github.com/apex/log"
	"github.com/open-collaboration/server/dtos"
	"github.com/open-collaboration/server/services"
	"github.com/open-collaboration/server/utils"
	"net/http"
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

func RouteAuthenticateUser(
	ctx context.Context,
	writer http.ResponseWriter,
	request *http.Request,
	usersService *services.UsersService,
	authService *services.AuthService,
) error {
	logger := log.FromContext(ctx)

	dto := dtos.LoginDto{}
	err := utils.ReadJson(request, ctx, &dto)
	if err != nil {
		return err
	}

	user, err := usersService.AuthenticateUser(ctx, dto)
	if err != nil {
		return err
	}

	if user != nil {
		sessionToken, err := authService.CreateSession(ctx, user.ID)
		if err != nil {
			logger.WithError(err).Error("Failed to create session")
		}

		userData := dtos.UserDataDto{
			Username: user.Username,
			Email:    user.Email,
		}

		cookieHeader := fmt.Sprintf("%s=%s", "sessionToken", sessionToken)
		writer.Header().Set("Set-Cookie", cookieHeader)

		err = utils.WriteJson(writer, ctx, http.StatusOK, userData)
		if err != nil {
			return err
		}
	} else {
		writer.WriteHeader(401)
	}

	return nil
}
