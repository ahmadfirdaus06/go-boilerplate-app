package services

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"go-boilerplate-backend/app/externals"
	"go-boilerplate-backend/app/models"
	"go-boilerplate-backend/app/repo"
	"go-boilerplate-backend/app/types"
	"log"
	"strings"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/argon2"
)

type UserService struct {
	UserRepo *repo.UserRepo[models.User]
}

func NewUserService(appExternals *externals.AllAppExternals) *UserService {
	mongoExt, mongoExtError := externals.GetExternal[*externals.MongoDBExternal](appExternals)

	if mongoExtError != nil {
		log.Fatalf("%v", mongoExtError)
		return nil
	}

	userRepo := repo.NewUserRepo[models.User](types.AppDB{MongoDB: mongoExt.DB}, "users")

	return &UserService{
		UserRepo: userRepo,
	}
}

func (s *UserService) RegisterUser(inputs any) (any, error) {
	data, err := json.Marshal(inputs)
	if err != nil {
		return nil, err
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	if usersWithExistingUsername, err := s.UserRepo.GetAll(false, &types.GetAllFiltersAndSorts{
		QueryParamsFilters: []types.QueryParamsFilter{
			{Field: "username", Operator: types.OpEq, Value: raw["username"].(string)},
		},
	}, nil); err != nil {
		return nil, err
	} else if len(usersWithExistingUsername.Records) > 0 {
		return nil, echo.NewHTTPError(400, "Username already taken.")
	}

	if usersWithExistingEmail, err := s.UserRepo.GetAll(false, &types.GetAllFiltersAndSorts{
		QueryParamsFilters: []types.QueryParamsFilter{
			{Field: "email", Operator: types.OpEq, Value: raw["email"].(string)},
		},
	}, nil); err != nil {
		return nil, err
	} else if len(usersWithExistingEmail.Records) > 0 {
		return nil, echo.NewHTTPError(400, "Email already taken.")
	}

	delete(raw, "confirmPassword")

	raw["emailVerifiedAt"] = nil

	hashedPassword, hashPasswordErr := generateHash(raw["password"].(string), nil)

	if hashPasswordErr != nil {
		return nil, hashPasswordErr
	}

	raw["password"] = hashedPassword

	created, createErr := s.UserRepo.Create(raw)

	if createErr != nil {
		return nil, createErr
	}

	return created, nil
}

type ArgonParams struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

var DefaultParams = &ArgonParams{
	Memory:      64 * 1024, // 64 MB
	Iterations:  3,
	Parallelism: 2,
	SaltLength:  16,
	KeyLength:   32,
}

func generateHash(password string, p *ArgonParams) (string, error) {
	if p == nil {
		p = DefaultParams
	}

	salt := make([]byte, p.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(password), salt, p.Iterations, p.Memory, p.Parallelism, p.KeyLength)

	return fmt.Sprintf("%s$%s",
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	), nil
}

func verifyPassword(password, encodedHash string, p *ArgonParams) (bool, error) {
	if p == nil {
		p = DefaultParams
	}

	parts := strings.Split(encodedHash, "$")
	if len(parts) != 2 {
		return false, fmt.Errorf("invalid hash format")
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[0])
	if err != nil {
		return false, err
	}

	hash, err := base64.RawStdEncoding.DecodeString(parts[1])
	if err != nil {
		return false, err
	}

	computedHash := argon2.IDKey([]byte(password), salt, p.Iterations, p.Memory, p.Parallelism, p.KeyLength)
	return subtle.ConstantTimeCompare(hash, computedHash) == 1, nil
}
