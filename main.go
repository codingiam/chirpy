package main

import (
	"codingiam/chirpy/internal/database"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	sql            *database.Queries
	platform       string
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func healthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func validateChirp(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	type parameters struct {
		Body string `json:"body"`
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

	type response struct {
		CleanedBody string `json:"cleaned_body"`
	}
	resp := response{CleanedBody: cleanedBody}

	writeSuccessJson(w, resp)
}

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
	_ = json.NewEncoder(w).Encode(resp)
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

func main() {
	godotenv.Load()

	const filepathRoot = "."
	const port = "8080"

	platform := os.Getenv("PLATFORM")

	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalln(err)
	}

	dbQueries := database.New(db)

	cfg := &apiConfig{
		fileserverHits: atomic.Int32{},
		sql:            dbQueries,
		platform:       platform,
	}

	mux := http.NewServeMux()

	handle := http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))
	mux.Handle("/app/", cfg.middlewareMetricsInc(handle))

	mux.HandleFunc("GET /api/healthz", healthz)
	mux.HandleFunc("POST /api/validate_chirp", validateChirp)
	mux.HandleFunc("POST /api/users", cfg.createUser)

	mux.HandleFunc("GET /admin/metrics", cfg.metricsHandler)
	mux.HandleFunc("POST /admin/reset", cfg.resetHandler)

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)

	err = http.ListenAndServe(":"+port, mux)
	if err != nil {
		log.Fatal(err)
	}
}
