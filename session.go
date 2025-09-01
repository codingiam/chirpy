package main

import (
	"codingiam/chirpy/internal/auth"
	"codingiam/chirpy/internal/database"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

func (cfg *apiConfig) createSession(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	type parameters struct {
		Password string `json:"password"`
		Email    string `json:"email"`
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
	jwt, err := auth.MakeJWT(user.ID, cfg.secret, duration)
	if err != nil {
		writeErrorJson(w, err, "Something went wrong")
		return
	}

	token, err := auth.MakeRefreshToken()
	if err != nil {
		writeErrorJson(w, err, "Something went wrong")
		return
	}

	_, err = cfg.sql.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     token,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(60 * 24 * time.Hour),
	})
	if err != nil {
		writeErrorJson(w, err, "Something went wrong")
		return
	}

	type response struct {
		ID           uuid.UUID `json:"id"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
		Email        string    `json:"email"`
		Token        string    `json:"token"`
		RefreshToken string    `json:"refresh_token"`
		IsChirpyRed  bool      `json:"is_chirpy_red"`
	}
	resp := response{user.ID, user.CreatedAt, user.UpdatedAt, user.Email, jwt, token, user.IsChirpyRed}

	writeSuccessJson(w, resp)
}

func (cfg *apiConfig) refreshSession(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		writeErrorJson(w, err, "Something went wrong")
		return
	}

	refreshToken, err := cfg.sql.GetRefreshTokenByToken(r.Context(), token)
	if err != nil || refreshToken.ExpiresAt.Before(time.Now()) {
		w.WriteHeader(http.StatusUnauthorized)
		writeErrorJson(w, err, "Something went wrong")
		return
	}

	duration := 1 * time.Hour
	jwt, err := auth.MakeJWT(refreshToken.UserID, cfg.secret, duration)
	if err != nil {
		writeErrorJson(w, err, "Something went wrong")
		return
	}

	type response struct {
		Token string `json:"token"`
	}
	resp := response{jwt}

	writeSuccessJson(w, resp)
}

func (cfg *apiConfig) revokeSession(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		writeErrorJson(w, err, "Something went wrong")
		return
	}

	refreshToken, err := cfg.sql.GetRefreshTokenByToken(r.Context(), token)
	if err != nil || refreshToken.ExpiresAt.Before(time.Now()) {
		w.WriteHeader(http.StatusUnauthorized)
		writeErrorJson(w, err, "Something went wrong")
		return
	}

	_, err = cfg.sql.RevokeRefreshTokenByToken(r.Context(), token)
	if err != nil {
		writeErrorJson(w, err, "Something went wrong")
		return
	}
	writeSuccessJson(w, nil, http.StatusNoContent)
}
