package request

type ResquestUserDTO struct {
	Email        string `json:"email"`
	PasswordHash string `json:"password_hash"`
}
