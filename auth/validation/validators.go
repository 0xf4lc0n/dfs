package validation

import (
	"auth/dtos"
	"unicode"

	"github.com/go-playground/validator"
)

var validate = validator.New()
var _ = validate.RegisterValidation("password", validatePasswordComplexity)

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

func validatePasswordComplexity(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	var number bool
	var bigLetter bool
	var smallLetter bool
	var specialCharacter bool
	characters := make(map[rune]int)

	if len(password) < 12 || len(password) > 48 {
		return false
	}

	for _, c := range password {
		switch {
		case unicode.IsNumber(c):
			number = true
		case unicode.IsUpper(c):
			bigLetter = true
			characters[c] += 1
		case unicode.IsLower(c):
			smallLetter = true
			characters[c] += 1
		case unicode.IsSymbol(c) || unicode.IsPunct(c):
			specialCharacter = true
		}
	}

	if len(characters) < 6 {
		return false
	}

	return number && bigLetter && smallLetter && specialCharacter
}
