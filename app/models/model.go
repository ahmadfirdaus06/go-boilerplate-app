package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type User struct {
	ID                             bson.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Username                       string        `json:"username" bson:"username"`
	FirstName                      string        `json:"firstName" bson:"firstName"`
	LastName                       string        `json:"lastName" bson:"lastName"`
	Email                          string        `json:"email" bson:"email"`
	Password                       string        `json:"password" bson:"password"`
	EmailVerifiedAt                *time.Time    `json:"emailVerifiedAt" bson:"emailVerifiedAt"`
	EmailVerificationCode          *string       `json:"emailVerificationCode" bson:"emailVerificationCode"`
	EmailVerificationCodeExpiredAt *time.Time    `json:"emailVerificationCodeExpiredAt" bson:"emailVerificationCodeExpiredAt"`
	CreatedAt                      *time.Time    `json:"createdAt" bson:"createdAt"`
	UpdatedAt                      *time.Time    `json:"updatedAt" bson:"updatedAt"`
}
