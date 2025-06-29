package auth

import "golang.org/x/crypto/bcrypt"

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
