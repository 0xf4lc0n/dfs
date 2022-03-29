package validation

import (
	"auth/dtos"

	"github.com/go-playground/validator"
)

var validate = validator.New()

func Validate[V dtos.LoginDto | dtos.RegisterDto](v *V) []string {
	var invalidFields []string
	err := validate.Struct(v)

	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			invalidFields = append(invalidFields, err.StructField())
		}
	}

	return invalidFields
}
