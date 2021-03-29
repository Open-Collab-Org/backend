package users

import (
	"errors"
	"gorm.io/gorm"
)

type NewUserDto struct {
	Username       string `json:"username" binding:"required,min=4,max=32"`
	Email          string `json:"email" binding:"required,email"`
	Password       string `json:"password" binding:"required,min=6,max=255"`
	RecaptchaToken string `json:"recaptchaToken"`
}

type LoginDto struct {
	UsernameOrEmail string `json:"usernameOrEmail"`
	Password        string `json:"password"`
	RecaptchaToken  string `json:"recaptchaToken"`
}

type UserDataDto struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}

type AuthenticatedUserDto struct {
	Token string      `json:"token"`
	User  UserDataDto `json:"user"`
}

func CreateUser(db *gorm.DB, newUser NewUserDto) error {
	user := User{
		Username: newUser.Username,
		Email:    newUser.Email,
	}

	err := user.SetPassword(newUser.Password)
	if err != nil {
		return err
	}

	db.Create(&user)

	return nil
}

func AuthenticateUser(db *gorm.DB, authUser LoginDto) (*User, error) {
	user := &User{}

	result := db.
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
