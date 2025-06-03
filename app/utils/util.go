package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

// Wrapper for marshal/unmarshal data against struct
func BindData(input any, outputSchema any) error {
	bytes, err := json.Marshal(input)

	if err != nil {
		return err
	}

	if err := json.Unmarshal(bytes, &outputSchema); err != nil {
		return err
	}

	return nil
}

// Generate JWT token
func GenerateJWTtToken(claims *jwt.MapClaims, JWTSecret string) (string, error) {
	if JWTSecret == "" {
		return "", fmt.Errorf("NO SECRET PROVIDED")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(JWTSecret))
}

// Verify JWT token
func VerifyJWTToken(tokenString string, JWTSecret string) (any, error) {
	if JWTSecret == "" {
		return "", fmt.Errorf("NO SECRET PROVIDED")
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return []byte(JWTSecret), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		return claims, nil
	}

	return nil, fmt.Errorf("CLAIMS DOES NOT EXIST")
}

func GetAppConfig(name string) string {
	if name == "" {
		return ""
	}

	return os.Getenv(name)
}

func LoadAppEnv(requiredEnvs []string) {
	err := godotenv.Load()

	if err != nil {
		log.Println("No .env file found. Fallback some env values from runtime.")
	}

	var missingEnvs string

	for _, env := range requiredEnvs {
		if os.Getenv(env) == "" {
			missingEnvs = missingEnvs + fmt.Sprintf("%s is missing or unset. ", env)
		}
	}

	if len(missingEnvs) > 0 {
		log.Fatalln(missingEnvs)
	}

}
