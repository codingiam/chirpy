package main

import (
	"codingiam/chirpy/internal/database"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

func (cfg *apiConfig) createChirp(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	type parameters struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}

	var params parameters
	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		writeErrorJson(w, err, "Something went wrong")
		return
	}

	if len(params.Body) > 140 {
		writeErrorJson(w, errors.New("chirp is too long"), "Chirp is too long")
		return
	}

	cleanedBody := replaceProfane(params.Body)

	chirp, err := cfg.sql.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   cleanedBody,
		UserID: params.UserID,
	})
	if err != nil {
		writeErrorJson(w, err, "Something went wrong")
		return
	}

	type response struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserID    uuid.UUID `json:"user_id"`
	}
	resp := response{chirp.ID, chirp.CreatedAt, chirp.UpdatedAt, cleanedBody, params.UserID}

	writeSuccessJson(w, resp, http.StatusCreated)
}

func replaceProfane(body string) string {
	words := strings.Split(body, " ")
	newBody := []string{}
	for _, word := range words {
		w := strings.TrimSpace(strings.ToLower(word))
		switch w {
		case "kerfuffle",
			"sharbert",
			"fornax":
			newBody = append(newBody, "****")
		default:
			newBody = append(newBody, word)
		}
	}
	return strings.Join(newBody, " ")
}

func (cfg *apiConfig) indexChirps(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	chirps, err := cfg.sql.GetChirps(r.Context())
	if err != nil {
		writeErrorJson(w, err, "Something went wrong")
		return
	}

	type response struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserID    uuid.UUID `json:"user_id"`
	}

	var resp []response
	for _, chirp := range chirps {
		resp = append(resp, response{chirp.ID, chirp.CreatedAt, chirp.UpdatedAt, chirp.Body, chirp.UserID})
	}

	writeSuccessJson(w, resp)
}

func (cfg *apiConfig) showChirp(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		writeErrorJson(w, err, "Something went wrong")
		return
	}

	chirp, err := cfg.sql.GetChirpByID(r.Context(), chirpID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		writeErrorJson(w, err, "Something went wrong")
		return
	}

	type response struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserID    uuid.UUID `json:"user_id"`
	}
	resp := response{chirp.ID, chirp.CreatedAt, chirp.UpdatedAt, chirp.Body, chirp.UserID}

	writeSuccessJson(w, resp)
}
