package main

import (
	"codingiam/chirpy/internal/auth"
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

func (cfg *apiConfig) createSession(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	type parameters struct {
		Password         string `json:"password"`
		Email            string `json:"email"`
		ExpiresInSeconds int    `json:"expires_in_seconds,omitempty"`
	}

	var params parameters
	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		writeErrorJson(w, err, "Something went wrong")
		return
	}

	email := strings.TrimSpace(params.Email)
	if len(email) < 5 {
		writeErrorJson(w, errors.New("email is too short"), "Email is too short")
		return
	}

	user, err := cfg.sql.GetUserByEmail(r.Context(), email)
	if err != nil {
		writeErrorJson(w, err, "Couldn't get user")
		return
	}

	err = auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		writeErrorJson(w, err, "Incorrect email or password")
		return
	}

	duration := 1 * time.Hour
	if params.ExpiresInSeconds > 0 {
		duration = time.Duration(math.Min(float64(duration), float64(time.Duration(params.ExpiresInSeconds)*time.Second)))
	}
	jwt, err := auth.MakeJWT(user.ID, cfg.secret, duration)
	if err != nil {
		writeErrorJson(w, err, "Something went wrong")
		return
	}

	type response struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
		Token     string    `json:"token"`
	}
	resp := response{user.ID, user.CreatedAt, user.UpdatedAt, user.Email, jwt}

	writeSuccessJson(w, resp)
}
