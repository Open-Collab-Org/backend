package services

import (
	"errors"
	"github.com/open-collaboration/server/dtos"
	"github.com/open-collaboration/server/models"
	"gorm.io/gorm"
)

type UsersService struct {
	Db *gorm.DB
}

func (s *UsersService) CreateUser(newUser dtos.NewUserDto) error {
	user := models.User{
		Username: newUser.Username,
		Email:    newUser.Email,
	}

	err := user.SetPassword(newUser.Password)
	if err != nil {
		return err
	}

	result := s.Db.Create(&user)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (s *UsersService) AuthenticateUser(authUser dtos.LoginDto) (*models.User, error) {
	user := &models.User{}

	result := s.Db.
		Where("username = ?", authUser.UsernameOrEmail).
		Or("email = ?", authUser.UsernameOrEmail).
		First(user)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		} else {
			return nil, result.Error
		}
	}

	passwordMatch, err := user.ComparePassword(authUser.Password)
	if err != nil {
		return nil, err
	} else if passwordMatch {
		return user, nil
	} else {
		return nil, nil
	}
}
