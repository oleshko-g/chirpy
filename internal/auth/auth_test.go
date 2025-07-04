package auth

import (
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestSignUserJWT(t *testing.T) {
	userUUID, err := uuid.NewRandom()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		t.FailNow()
	}

	// Test: JWT signed
	jwtSecret := "pirch"
	tokenString, err := signUserJWT(userUUID, jwtSecret, 5*time.Second)
	assert.NoError(t, err)
	tokenStruct, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) { return []byte(jwtSecret), nil })
	assert.NoError(t, err)
	assert.True(t, tokenStruct.Valid)
}

func TestValidateUserJWT(t *testing.T) {
	userUUID, err := uuid.NewRandom()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		t.FailNow()
	}

	token, err := signUserJWT(userUUID, "pirch", 5*time.Second)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		t.FailNow()
	}

	// Test: valid jwtSecret
	userUUIDFromClaims, err := validateUserJWT(token, "pirch")
	assert.NoError(t, uuid.Validate(userUUIDFromClaims.String()))
	assert.NoError(t, err)

	// Test: valid jwtSecret has expired
	time.Sleep(5 * time.Second)
	_, err = validateUserJWT(token, "pirch")
	assert.Error(t, err)
}

func TestGetBearerToken(t *testing.T) {
	userUUID, err := uuid.NewRandom()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		t.FailNow()
	}
	jwtSecret := "pirch"
	tokenString, err := signUserJWT(userUUID, jwtSecret, 5*time.Second)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		t.FailNow()
	}
	headers := make(http.Header)
	headers.Add("Authorization", "Bearer "+tokenString)
	jwtString, err := GetBearerToken(&headers)
	assert.NoError(t, err)
	assert.NotEmpty(t, jwtString)
}
