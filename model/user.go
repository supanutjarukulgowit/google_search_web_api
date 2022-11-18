package model

import "time"

type User struct {
	Id          string    `json:"id"`
	Username    string    `json:"username"`
	Password    string    `json:"-"`
	CreatedDate time.Time `json:"created_date"`
}

type SignUpRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type SignInRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type Cookie struct {
	Name      string
	Value     string
	ExpiresAt time.Time
	HTTPOnly  bool
}
