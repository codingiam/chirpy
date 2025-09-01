package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
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

func TestJWT(t *testing.T) {
	userID := uuid.New()
	tokenSecret := "secret"
	jwt, err := MakeJWT(userID, tokenSecret, 5*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	decodedUUID, err := ValidateJWT(jwt, tokenSecret)
	if err != nil {
		t.Fatal(err)
	}
	if decodedUUID != userID {
		t.Fatalf("UUID does not match")
	}
}
