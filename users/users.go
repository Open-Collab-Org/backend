package users

import (
	"gorm.io/gorm"
)

type NewUserDto struct {
	Username string `json:"username" binding:"required,min=4,max=32"`
	Email string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6,max=255"`
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