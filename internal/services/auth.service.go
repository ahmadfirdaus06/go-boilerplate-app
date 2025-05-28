package services

import (
	"fmt"
	"go-boilerplate-backend/internal/externals"
	"go-boilerplate-backend/internal/models"
	"go-boilerplate-backend/internal/repo"
	"go-boilerplate-backend/internal/utils"
	"math/rand"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

type AuthService struct {
	UserRepo *repo.UserRepo[models.User]
}

var JwtSecret = "Abcd1234"

func NewAuthService(externals *externals.AllAppExternals) *AuthService {
	return &AuthService{
		UserRepo: repo.NewUserRepo[models.User](externals.JsonDB, "users"),
	}
}

func (as *AuthService) LoginUser(usernameOrEmail string, password string) (string, error) {
	users, err := as.UserRepo.GetAll([]externals.JsonDBPipelineStage{
		{
			Op: "$match",
			Value: map[string]interface{}{
				"$or": []interface{}{
					map[string]interface{}{"email": usernameOrEmail},
					map[string]interface{}{"username": usernameOrEmail},
				},
			},
		},
	})

	if err != nil {
		return "", err
	}

	if len(users) == 0 {
		return "", echo.NewHTTPError(401, "Wrong username / email or password.")
	}

	if ok, err := verifyPassword(password, users[0].Password, nil); err != nil {
		return "", err
	} else if !ok {
		return "", echo.NewHTTPError(401, "Wrong username / email or password.")
	}

	return utils.GenerateJWTtToken(&jwt.MapClaims{
		"user": map[string]string{
			"id":    users[0].ID,
			"email": users[0].Email,
		},
	}, JwtSecret)
}

var VerificationCodeExpiredDuration = 2 * time.Minute

func (as *AuthService) SendVerificationCode(user *models.User) (*time.Time, error) {
	code := generateVerificationCode()

	updatedUser, updateErr := as.UserRepo.UpdateByID(user.ID, map[string]string{"emailVerificationCode": code, "emailVerificationCodeExpiredAt": (time.Now().Add(VerificationCodeExpiredDuration)).Format(time.RFC3339)})

	if updateErr != nil {
		return nil, updateErr
	}

	//TODO: send email using queue

	return updatedUser.EmailVerificationCodeExpiredAt, nil
}

func generateVerificationCode() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	code := r.Intn(900000) + 100000
	return fmt.Sprintf("%06d", code)
}
