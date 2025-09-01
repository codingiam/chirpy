package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

func writeErrorJson(w http.ResponseWriter, err error, message ...string) {
	log.Printf("Error: %s", err)

	type response struct {
		Error string `json:"error"`
	}
	resp := response{Error: strings.Join(message, " ")}

	w.WriteHeader(http.StatusBadRequest)
	_ = json.NewEncoder(w).Encode(resp)
}

func writeSuccessJson(w http.ResponseWriter, resp any, statusCode ...int) {
	if statusCode != nil {
		w.WriteHeader(statusCode[0])
	} else {
		w.WriteHeader(http.StatusOK)
	}
	if resp != nil {
		_ = json.NewEncoder(w).Encode(resp)
	}
}

func healthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	out := fmt.Sprintf("<html>\n  <body>\n    <h1>Welcome, Chirpy Admin</h1>\n    <p>Chirpy has been visited %d times!</p>\n  </body>\n</html>", cfg.fileserverHits.Load())
	w.Write([]byte(out))
}

func (cfg *apiConfig) resetHandler(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	cfg.fileserverHits.Store(0)

	err := cfg.sql.DeleteUsers(r.Context())
	if err != nil {
		writeErrorJson(w, err, "Couldn't truncate users")
		return
	}
}
