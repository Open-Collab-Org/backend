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

// @Summary Register a new user
// @Tags users
// @Router /users [post]
// @Param userData body dtos.NewUserDto true "The user's data"
// @Success 201
func RouteRegisterUser(
	ctx context.Context,
	writer http.ResponseWriter,
	request *http.Request,
	usersService *services.UsersService,
) error {
	dto := dtos.NewUserDto{}
	err := utils.ReadJson(request, ctx, &dto)
	if err != nil {
		return err
	}

	err = usersService.CreateUser(ctx, dto)
	if err != nil {
		return err
	}

	writer.WriteHeader(201)

	return nil
}

// @Summary Authenticate user
// @Tags users
// @Router /login [post]
// @Param credentials body dtos.LoginDto true "The user's credentials"
// @Success 200 {object} dtos.UserDataDto "User successfully authenticated"
// @Header 200 {string} Set-Cookie "Session token. E.g. sessionToken=72f34c69-6eb0-47cf-83ed-c2b5ad3989df"
// @Failure 401
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
