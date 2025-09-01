package main

import (
	"codingiam/chirpy/internal/auth"
	"codingiam/chirpy/internal/database"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

func (cfg *apiConfig) polkaWebhook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil || apiKey != cfg.polkaKey {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	type parameters struct {
		Event string `json:"event"`
		Data  struct {
			UserID uuid.UUID `json:"user_id"`
		} `json:"data"`
	}

	var params parameters
	err = json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		writeErrorJson(w, err, "Something went wrong")
		return
	}

	if params.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	userID := params.Data.UserID
	err = cfg.sql.UpdateUserIsChirpyRed(r.Context(), database.UpdateUserIsChirpyRedParams{
		ID: userID, IsChirpyRed: true,
	})
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		writeErrorJson(w, err, "Something went wrong")
		return
	}

	writeSuccessJson(w, nil, http.StatusNoContent)
}
