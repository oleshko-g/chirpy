package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {

	hashedPasswordData, errHash := bcrypt.GenerateFromPassword([]byte(password), 0)
	if errHash != nil {
		return "", errHash
	}

	return string(hashedPasswordData), nil
}

func CheckPasswordHash(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func SignUserJWT(userID uuid.UUID, jwtSecret string, expiresIn time.Duration) (string, error) {
	t := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.RegisteredClaims{
			Issuer:    "chirpy",
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiresIn)),
			Subject:   userID.String(),
		},
	)
	return t.SignedString([]byte(jwtSecret))
}

func ValidateUserJWT(tokenString, jwtSecret string) (uuid.UUID, error) {
	claims := jwt.RegisteredClaims{}
	userJWT, err := jwt.ParseWithClaims(tokenString, &claims,
		func(token *jwt.Token) (interface{}, error) {
			if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
				return nil, fmt.Errorf("error signing method must be HS256. Token's method is %s", token.Method.Alg())
			}
			return []byte(jwtSecret), nil
		},
	)
	if err != nil {
		return uuid.UUID{}, err
	}
	id, err := userJWT.Claims.GetSubject()
	if err != nil {
		return uuid.UUID{}, err
	}
	return uuid.Parse(id)
}

func GetBearerToken(headers *http.Header) (string, error) {
	authHeader, err := getAuthHeader(headers)
	if err != nil {
		return authHeader, err
	}

	bearerString, ok := strings.CutPrefix(authHeader, "Bearer ")
	if !ok {
		return "", fmt.Errorf("error token string doesn't start with \"Bearer \"")
	}

	return strings.TrimLeft(bearerString, " "), nil
}

func MakeRefreshToken() (string, error) {
	randomData := make([]byte, 32)
	rand.Read(randomData)
	return hex.EncodeToString(randomData), nil
}

func GetApiKey(headers *http.Header) (string, error) {
	authHeader, err := getAuthHeader(headers)
	if err != nil {
		return authHeader, err
	}
	apiKey, ok := strings.CutPrefix(authHeader, "ApiKey ")
	if !ok {
		return "", fmt.Errorf("error token string doesn't start with \"ApiKey \"")
	}

	return strings.TrimLeft(apiKey, " "), nil
}

func getAuthHeader(headers *http.Header) (string, error) {
	authHeader, ok := (*headers)["Authorization"]
	if !ok {
		return "", fmt.Errorf("error no Authorization header")
	}

	if authHeader[0] == "" {
		return "", fmt.Errorf("error Authorization header is not set")
	}

	return authHeader[0], nil
}
