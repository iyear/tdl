package validator

import (
	"github.com/go-playground/validator/v10"
)

var v *validator.Validate

func init() {
	v = validator.New()
}

func Struct(s interface{}) error {
	return v.Struct(s)
}
