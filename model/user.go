package model

import "time"

type User struct {
	Id          string    `json:"id"`
	Username    string    `json:"username"`
	Password    string    `json:"-"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	CreatedDate time.Time `json:"created_date"`
}

type SignUpRequest struct {
	Username  string `json:"username" validate:"required"`
	Password  string `json:"password" validate:"required"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type SignInRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type SignInResponse struct {
	Token  string `json:"token"`
	UserID string `json:"user_id"`
}
