package dtos

type RegisterDto struct {
	Name     string `json:"name" validate:"required,min=6,max=32"`
	Email    string `json:"email" validate:"required,email,min=6,max=48"`
	Password string `json:"password" validate:"required,min=12,max=48,password"`
}
