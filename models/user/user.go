package user

import "time"

type User struct {
	ID                   string    `bson:"_id,omitempty"`
	FirstName            string    `validate:"required"`
	LastName             string    `validate:"required"`
	Email                string    `validate:"required,email"`
	Password             string    `validate:"required,min=6,max=30"`
	Verified             bool      `bson:"verified" json:"verified"`
	ResetToken           string    `bson:"resetToken,omitempty" json:"resetToken,omitempty"`
	ResetTokenExpiration time.Time `bson:"resetTokenExpiration,omitempty" json:"resetTokenExpiration,omitempty"`
}

type UserRequestBody struct {
	FirstName string `json:"firstName" validate:"required"`
	LastName  string `json:"lastName" validate:"required"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=6,max=30"`
}
