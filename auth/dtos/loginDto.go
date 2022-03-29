package dtos

type LoginDto struct {
	Email    string `json:"email" validate:"required,email,min=6,max=48"`
	Password string `json:"password" validate:"required,min=12,max=48"`
}
