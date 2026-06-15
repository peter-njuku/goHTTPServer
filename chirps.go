package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/peter-njuku/goHTTPServer/internal/auth"
	"github.com/peter-njuku/goHTTPServer/internal/database"
)

type chirpRequest struct {
	Body string `json:"body"`
}

type chirpResponse struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

type chirpResponseError struct {
	Error string `json:"error"`
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(data)
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	respondWithJSON(w, code, chirpResponseError{Error: msg})
}

func scrubProfanity(text string) (string, bool) {
	bannedWords := map[string]bool{
		"kerfuffle": true,
		"sharbert":  true,
		"fornax":    true,
	}

	words := strings.Split(text, " ")
	censored := false
	for i, word := range words {
		if bannedWords[strings.ToLower(word)] {
			words[i] = "****"
			censored = true
		}
	}

	return strings.Join(words, " "), censored
}

func (cfg *ApiConfig) handlerValidateChirp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		log.Printf("OOOPS! Wrong API method - %d\n", http.StatusMethodNotAllowed)
		return
	}

	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		log.Print(err)
		return
	}

	userId, err := auth.ValidateJWT(tokenString, cfg.Secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		log.Print(err)
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := chirpRequest{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Something went wrong decoding parameters")
		log.Print(err)
		return
	}
	if len(params.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	cleaned, _ := scrubProfanity(params.Body)
	dbChirp, err := cfg.Db.CreateChirps(r.Context(), database.CreateChirpsParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Body:      cleaned,
		UserID:    userId,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create new chirp")
		log.Print(err)
		return
	}
	respondWithJSON(w, http.StatusOK, chirpResponse{
		ID:        dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		Body:      dbChirp.Body,
		UserID:    dbChirp.UserID,
	})
}

func (cfg *ApiConfig) handlerChirpsRetrieve(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		log.Printf("OOOPS! Wrong API method - %d\n", http.StatusMethodNotAllowed)
		return
	}

	authorIDString := r.URL.Query().Get("author_id")
	sortOrder := r.URL.Query().Get("sort")

	var dbChirps []database.Chirp
	var err error

	if authorIDString != "" {
		authorId, parseErr := uuid.Parse(authorIDString)
		if parseErr != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid author ID format")
			return
		}

		dbChirps, err = cfg.Db.GetChirpsForAuthor(r.Context(), authorId)
	} else {
		dbChirps, err = cfg.Db.GetAllChirps(r.Context())
	}

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not retrive all chirps")
		log.Print(err)
		return
	}

	chirps := []chirpResponse{}
	for _, dbChirp := range dbChirps {
		chirps = append(chirps, chirpResponse{
			ID:        dbChirp.ID,
			CreatedAt: dbChirp.CreatedAt,
			UpdatedAt: dbChirp.UpdatedAt,
			Body:      dbChirp.Body,
			UserID:    dbChirp.UserID,
		})
	}

	if sortOrder == "desc" {
		sort.Slice(chirps, func(i, j int) bool {
			return chirps[i].CreatedAt.After(chirps[j].CreatedAt)
		})
	} else {
		sort.Slice(chirps, func(i, j int) bool {
			return chirps[i].CreatedAt.Before(chirps[j].CreatedAt)
		})
	}
	respondWithJSON(w, http.StatusOK, chirps)
}

func (cfg *ApiConfig) handlerChirpGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		log.Printf("OOOPS! Wrong API method - %d\n", http.StatusMethodNotAllowed)
		return
	}

	idString := r.PathValue("chirpID")
	chirpID, err := uuid.Parse(idString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid Chirp-ID")
		log.Print(err)
		return
	}

	dbChirp, err := cfg.Db.RetriveOneChirp(r.Context(), chirpID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "Chirp not found")
			log.Print(err)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Couldn't retrieve chirp")
		log.Print(err)
		return
	}

	respondWithJSON(w, http.StatusOK, chirpResponse{
		ID:        dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		Body:      dbChirp.Body,
		UserID:    dbChirp.UserID,
	})
}

func (cfg *ApiConfig) HandlerDeleteChirp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		log.Printf("OOOPS! Wrong API method - %d\n", http.StatusMethodNotAllowed)
		return
	}

	idString := r.PathValue("chirpID")
	chirpId, err := uuid.Parse(idString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid Chirp Id")
		log.Print(err)
		return
	}

	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		log.Print(err)
		return
	}
	userId, err := auth.ValidateJWT(tokenString, cfg.Secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		log.Print(err)
		return
	}

	dbChirp, err := cfg.Db.RetriveOneChirp(r.Context(), chirpId)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Chirp not found")
		log.Print(err)
		return
	}

	if dbChirp.UserID != userId {
		respondWithError(w, http.StatusForbidden, "You do not have permission to do the action")
		return
	}

	err = cfg.Db.DeleteChirpById(r.Context(), chirpId)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not delete chirp")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
