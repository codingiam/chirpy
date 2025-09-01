package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

func (cfg *apiConfig) createUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	type parameters struct {
		Email string `json:"email"`
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

	user, err := cfg.sql.CreateUser(r.Context(), email)
	if err != nil {
		writeErrorJson(w, err, "Couldn't create user")
		return
	}

	type response struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}
	resp := response{user.ID, user.CreatedAt, user.UpdatedAt, user.Email}

	writeSuccessJson(w, resp, http.StatusCreated)
}
