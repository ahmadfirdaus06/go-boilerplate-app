package services

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/ahmadfirdaus06/go-boilerplate-app/app/externals"
	"github.com/ahmadfirdaus06/go-boilerplate-app/app/models"
	"github.com/ahmadfirdaus06/go-boilerplate-app/app/repo"
	"github.com/ahmadfirdaus06/go-boilerplate-app/app/types"
	"github.com/ahmadfirdaus06/go-boilerplate-app/app/utils"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

type AuthService struct {
	UserRepo *repo.UserRepo[models.User]
}

var JwtSecret = "Abcd1234"

func NewAuthService(appExternals *externals.AllAppExternals) *AuthService {
	mongoExt, mongoExtError := externals.GetExternal[*externals.MongoDBExternal](appExternals)

	if mongoExtError != nil {
		log.Fatalf("%v", mongoExtError)
		return nil
	}

	return &AuthService{
		UserRepo: repo.NewUserRepo[models.User](types.AppDB{MongoDB: mongoExt.DB}, "users"),
	}
}

func (as *AuthService) LoginUser(usernameOrEmail string, password string) (string, error) {
	user, err := as.UserRepo.GetUserByUsernameOrEmail(usernameOrEmail)

	if err != nil {
		return "", err
	}

	if user == nil {
		return "", echo.NewHTTPError(401, "Wrong username / email or password.")
	}

	if ok, err := verifyPassword(password, user.Password, nil); err != nil {
		return "", err
	} else if !ok {
		return "", echo.NewHTTPError(401, "Wrong username / email or password.")
	}

	return utils.GenerateJWTtToken(&jwt.MapClaims{
		"user": map[string]any{
			"_id":   user.ID,
			"email": user.Email,
		},
	}, JwtSecret)
}

var VerificationCodeExpiredDuration = 2 * time.Minute

func (as *AuthService) SendVerificationCode(user *models.User) (*time.Time, error) {
	code := generateVerificationCode()

	var data = struct {
		EmailVerificationCode          string    `json:"emailVerificationCode" bson:"emailVerificationCode"`
		EmailVerificationCodeExpiredAt time.Time `json:"emailVerificationCodeExpiredAt" bson:"emailVerificationCodeExpiredAt"`
	}{
		EmailVerificationCode:          code,
		EmailVerificationCodeExpiredAt: time.Now(),
	}

	updatedUser, updateErr := as.UserRepo.UpdateByID(user.ID, data)

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

func (as *AuthService) VerifyCode(user *models.User, code string) (bool, error) {
	if *user.EmailVerificationCode != code {
		return false, nil
	}

	var data = struct {
		VerificationCode *string   `json:"emailVerificationCode" bson:"emailVerificationCode"`
		EmailVerifiedAt  time.Time `json:"emailVerifiedAt" bson:"emailVerifiedAt"`
	}{
		VerificationCode: nil,
		EmailVerifiedAt:  time.Now(),
	}

	if _, err := as.UserRepo.UpdateByID(user.ID, data); err != nil {
		return false, err
	}

	return true, nil
}
