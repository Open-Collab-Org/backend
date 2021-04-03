package dtos

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
