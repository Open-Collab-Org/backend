package auth

import (
	"fmt"
	"github.com/apex/log"
	"github.com/open-collaboration/server/users"
	"github.com/open-collaboration/server/utils"
	"net/http"
)

// @Summary Authenticate user
// @Tags users
// @Router /login [post]
// @Param credentials body dtos.LoginDto true "The user's credentials"
// @Success 200 {object} dtos.UserDataDto "User successfully authenticated"
// @Header 200 {string} Set-Cookie "Session token. E.g. sessionToken=72f34c69-6eb0-47cf-83ed-c2b5ad3989df"
// @Failure 401
func RouteAuthenticateUser(
	writer http.ResponseWriter,
	request *http.Request,
	authService Service,
) error {
	ctx := request.Context()

	logger := log.FromContext(ctx)

	dto := LoginDto{}
	err := utils.ReadJson(ctx, request, &dto)
	if err != nil {
		return err
	}

	user, err := authService.AuthenticateUser(ctx, dto)
	if err != nil {
		return err
	}

	if user != nil {
		sessionToken, err := authService.CreateSession(ctx, user.ID)
		if err != nil {
			logger.WithError(err).Error("Failed to create session")
		}

		userData := users.UserDataDto{
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
