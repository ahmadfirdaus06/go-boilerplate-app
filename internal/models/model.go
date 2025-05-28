package models

import "time"

type User struct {
	ID                             string     `json:"id"`
	Username                       string     `json:"username"`
	FirstName                      string     `json:"firstName"`
	LastName                       string     `json:"lastName"`
	Email                          string     `json:"email"`
	Password                       string     `json:"password"`
	EmailVerifiedAt                *time.Time `json:"emailVerifiedAt"`
	EmailVerificationCode          *string    `json:"emailVerificationCode"`
	EmailVerificationCodeExpiredAt *time.Time `json:"emailVerificationCodeExpiredAt"`
	CreatedAt                      *time.Time `json:"createdAt"`
	UpdatedAt                      *time.Time `json:"updatedAt"`
}
