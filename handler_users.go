package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/dmytrochumakov/chirpy/internal/auth"
	"github.com/dmytrochumakov/chirpy/internal/database"
	"github.com/google/uuid"
)

type User struct {
	ID          uuid.UUID `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Email       string    `json:"email"`
	IsChirpyRed bool      `json:"is_chirpy_red"`
}

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	type requestParams struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	reqParans := requestParams{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&reqParans)
	if err != nil {
		write500Error(w)
		return
	}
	hashedPassword, err := auth.HashPassword(reqParans.Password)
	if err != nil {
		write500Error(w)
		return
	}
	dbUser, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
		ID:             uuid.New(),
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
		Email:          reqParans.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		write500Error(w)
		return
	}

	respUser := User{
		ID:          dbUser.ID,
		CreatedAt:   dbUser.CreatedAt,
		UpdatedAt:   dbUser.UpdatedAt,
		Email:       dbUser.Email,
		IsChirpyRed: dbUser.IsChirpyRed,
	}

	dat, err := json.Marshal(respUser)

	w.WriteHeader(201)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		write500Error(w)
		return
	}

	w.Write(dat)

}

func (cfg *apiConfig) handlerUpdateUserEmailAndPassword(w http.ResponseWriter, r *http.Request) {
	authToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		write401Error(w)
		return
	}
	userID, err := auth.ValidateJWT(authToken, cfg.jwtSecret)
	if err != nil {
		write401Error(w)
		return
	}

	type parameters struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	params := parameters{}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&params)
	if err != nil {
		write500Error(w)
		return
	}
	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		write500Error(w)
		return
	}
	dbUser, err := cfg.db.UpdateUserEmailAndPassword(r.Context(), database.UpdateUserEmailAndPasswordParams{
		Email:          params.Email,
		HashedPassword: hashedPassword,
		UpdatedAt:      time.Now().UTC(),
		ID:             userID,
	})
	if err != nil {
		write500Error(w)
		return
	}

	writeJSONResponse(w, http.StatusOK, User{
		ID:          dbUser.ID,
		CreatedAt:   dbUser.CreatedAt,
		UpdatedAt:   dbUser.UpdatedAt,
		Email:       dbUser.Email,
		IsChirpyRed: dbUser.IsChirpyRed,
	})
}
