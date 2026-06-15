package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/peter-njuku/goHTTPServer/internal/auth"
	"github.com/peter-njuku/goHTTPServer/internal/database"
)

type UserRequest struct {
	Password         string `json:"password"`
	Email            string `json:"email"`
	ExpiresInSeconds *int   `json:"expires_in_seconds,omitempty"`
}

type UserCreatedResponse struct {
	Id        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	Token     string    `json:"token"`
}

func (cfg *ApiConfig) HandlerCreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		log.Printf("OOOPS! Wrong API method - %d\n", http.StatusMethodNotAllowed)
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := UserRequest{}

	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Something went wrong decoding the request")
		return
	}

	if params.Password == "" {
		respondWithError(w, http.StatusBadRequest, "Password is required")
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not secure password")
		log.Print(err)
		return
	}

	dbUser, err := cfg.Db.CreateUser(r.Context(), database.CreateUserParams{
		ID:             uuid.New(),
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
		Email:          params.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create a user")
		log.Print(err)
		return
	}

	respondWithJSON(w, http.StatusCreated, UserCreatedResponse{
		Id:        dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email:     dbUser.Email,
	})
}

func (cfg *ApiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		log.Printf("OOOPS! Wrong API method - %d\n", http.StatusMethodNotAllowed)
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := UserRequest{}

	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong decoding the request")
		log.Print(err)
		return
	}
	if params.Email == "" || params.Password == "" {
		respondWithError(w, http.StatusBadRequest, "Both Password and Email are required")
		return
	}

	dbUser, err := cfg.Db.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}

	match, err := auth.CheckPasswordHash(params.Password, dbUser.HashedPassword)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error checking password")
	}
	if !match {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}

	defaultExpiry := time.Hour
	maxExpiry := time.Hour
	expirationDuration := defaultExpiry

	if params.ExpiresInSeconds != nil {
		requestDuration := time.Duration(*params.ExpiresInSeconds) * time.Second
		if requestDuration > maxExpiry || requestDuration <= 0 {
			expirationDuration = maxExpiry
		} else {
			expirationDuration = requestDuration
		}
	}

	log.Print(expirationDuration)

	tokenString, err := auth.MakeJWT(dbUser.ID, cfg.Secret, expirationDuration)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not create access token")
		log.Print(err)
		return
	}

	respondWithJSON(w, http.StatusOK, UserCreatedResponse{
		Id:        dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email:     dbUser.Email,
		Token:     tokenString,
	})
}
