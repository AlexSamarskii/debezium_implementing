package models

import "github.com/go-playground/validator/v10"

var validate = validator.New()

type Request struct {
	Name  string `json:"name" validate:"required,min=2,max=100"`
	Email string `json:"email" validate:"required,email"`
}

func (r *Request) Validate() error {
	return validate.Struct(r)
}
