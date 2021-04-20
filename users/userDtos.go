package users

type NewUserDto struct {
	Username       string `json:"username" validate:"required,min=4,max=32"`
	Email          string `json:"email" validate:"required,email"`
	Password       string `json:"password" validate:"required,min=6,max=255"`
	RecaptchaToken string `json:"recaptchaToken"`
}

type UserDataDto struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}
