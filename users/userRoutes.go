package users

import (
	"github.com/open-collaboration/server/utils"
	"net/http"
)

// @Summary Register a new user
// @Tags users
// @Router /users [post]
// @Param userData body dtos.NewUserDto true "The user's data"
// @Success 201
func RouteRegisterUser(
	writer http.ResponseWriter,
	request *http.Request,
	usersService Service,
) error {
	dto := NewUserDto{}
	err := utils.ReadJson(request.Context(), request, &dto)
	if err != nil {
		return err
	}

	err = usersService.CreateUser(request.Context(), dto)
	if err != nil {
		return err
	}

	writer.WriteHeader(201)

	return nil
}
