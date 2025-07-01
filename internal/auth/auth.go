package auth

import (
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

func signUserJWT(userID uuid.UUID, jwtSecret string, expiresAfter time.Duration) (string, error) {
	t := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.RegisteredClaims{
			Issuer:    "chirpy",
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiresAfter)),
			Subject:   userID.String(),
		},
	)
	return t.SignedString([]byte(jwtSecret))
}

func validateUserJWT(tokenString, jwtSecret string) (uuid.UUID, error) {
	claims := jwt.RegisteredClaims{}
	userJWT, err := jwt.ParseWithClaims(tokenString, &claims,
		func(token *jwt.Token) (interface{}, error) {
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
