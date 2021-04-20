package users

import (
	"errors"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model

	Username     string
	Email        string
	PasswordHash string
}

func (user *User) SetPassword(plainTextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plainTextPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.PasswordHash = string(hash)

	return nil
}

func (user *User) ComparePassword(plainTextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(plainTextPassword))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, nil
		} else {
			return false, err
		}
	} else {
		return true, nil
	}
}
