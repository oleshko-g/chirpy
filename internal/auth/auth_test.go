package auth

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

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
