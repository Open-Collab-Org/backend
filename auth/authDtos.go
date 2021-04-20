package auth

type LoginDto struct {
	UsernameOrEmail string `json:"usernameOrEmail"`
	Password        string `json:"password"`
	RecaptchaToken  string `json:"recaptchaToken"`
}
