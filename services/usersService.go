package services

import (
	"context"
	"errors"
	"github.com/open-collaboration/server/dtos"
	"github.com/open-collaboration/server/logging"
	"github.com/open-collaboration/server/models"
	"gorm.io/gorm"
)

type UsersService struct {
	Db *gorm.DB
}

func (s *UsersService) CreateUser(ctx context.Context, newUser dtos.NewUserDto) error {
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

func (s *UsersService) AuthenticateUser(ctx context.Context, authUser dtos.LoginDto) (*models.User, error) {
	logger := logging.LoggerFromCtx(ctx).
		WithField("username", authUser.UsernameOrEmail)

	logger.Debug("Attempting to authenticate user")

	logger.Debug("Searching for user in database...")

	user := &models.User{}
	result := s.Db.
		Where("username = ?", authUser.UsernameOrEmail).
		Or("email = ?", authUser.UsernameOrEmail).
		First(user)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			logger.Debug("User not found")

			return nil, nil
		} else {
			logger.WithError(result.Error).Error("Could not query for user in database")

			return nil, result.Error
		}
	}

	logger.Debug("User found, comparing passwords...")

	passwordMatch, err := user.ComparePassword(authUser.Password)
	if err != nil {
		logger.WithError(err).Error("Error comparing passwords")

		return nil, err
	} else if passwordMatch {
		logger.Debug("Passwords match, user authenticated")

		return user, nil
	} else {
		logger.Debug("Wrong password")

		return nil, nil
	}
}
