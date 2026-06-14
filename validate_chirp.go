package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

type chirpRequest struct {
	Body string `json:"body"`
}

type chirpResponseValid struct {
	Valid bool `json:"valid"`
}

type chirpResponseError struct {
	Error string `json:"error"`
}

type chirpResponseCleaned struct {
	CleanedBody string `json:"cleaned_body"`
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

func handlerValidateChirp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		log.Printf("OOOPS! Wrong API method - %d\n", http.StatusMethodNotAllowed)
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := chirpRequest{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Something went wrong")
		return
	}
	if len(params.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	if cleaned, censored := scrubProfanity(params.Body); censored {
		respondWithJSON(w, http.StatusOK, chirpResponseCleaned{CleanedBody: cleaned})
		return
	}
	respondWithJSON(w, http.StatusOK, chirpResponseValid{Valid: true})
}
