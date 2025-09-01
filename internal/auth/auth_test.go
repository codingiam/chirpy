package auth

import (
	"net/http"
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

func TestGetBearerToken(t *testing.T) {
	headers := http.Header{}
	headers.Add("Authorization", "bearer TOKEN_STRING")

	if token, _ := GetBearerToken(headers); token != "TOKEN_STRING" {
		t.Fatalf("Bearer token does not match")
	}
}
