package users

import (
	"context"
	"errors"
	"github.com/apex/log"
	"gorm.io/gorm"
)

var ErrUserNotFound = errors.New("user not found")

type Service struct {
	Db *gorm.DB
}

// Create a user.
func (s *Service) CreateUser(ctx context.Context, newUser NewUserDto) error {
	user := User{
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

// Get a user by id.
// Returns ErrUserNotFound if a user with the specified id cannot be found.
func (s *Service) GetUser(ctx context.Context, id uint) (*User, error) {
	logger := log.FromContext(ctx)

	user := &User{}
	result := s.Db.First(user, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			logger.Debugf("User not found", id)

			return nil, ErrUserNotFound
		} else {
			logger.WithError(result.Error).Error("Database error")

			return nil, result.Error
		}
	}

	return user, nil
}

func (s *Service) FindUserByUsernameOrEmail(ctx context.Context, usernameOrEmail string) (*User, error) {
	logger := log.FromContext(ctx).WithField("usernameOrEmail", usernameOrEmail)

	logger.Debug("Searching for user on database")

	user := &User{}
	result := s.Db.
		Where("username = ?", usernameOrEmail).
		Or("email = ?", usernameOrEmail).
		First(user)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			logger.Debug("User not found")

			return nil, ErrUserNotFound
		} else {
			logger.WithError(result.Error).Error("Could not query for user in database")

			return nil, result.Error
		}
	}

	logger.Debug("User found")

	return user, nil
}
