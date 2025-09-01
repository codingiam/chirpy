package auth

import (
	"testing"
)

func TestPasswordHashing(t *testing.T) {
	password := "password"

	passwordHash, err := HashPassword(password)
	if err != nil {
		t.Fatal(err)
	}

	err = CheckPasswordHash(password, passwordHash)
	if err != nil {
		t.Fatal(err)
	}
}
