package main

import (
	"encoding/json"
	"net/http"
	"sort"
	"time"

	"github.com/dmytrochumakov/chirpy/internal/auth"
	"github.com/dmytrochumakov/chirpy/internal/database"
	"github.com/google/uuid"
)

type SortType string

const (
	SortTypeDESC SortType = "desc"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    string    `json:"user_id"`
}

func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, r *http.Request) {
	type requestBody struct {
		Body string `json:"body"`
	}
	reqBody := requestBody{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&reqBody)
	if err != nil {
		write500Error(w)
		return
	}
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		write401Error(w)
		return
	}
	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		write401Error(w)
		return
	}

	dbChirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Body:      reqBody.Body,
		UserID:    userID,
	})
	if err != nil {
		write500Error(w)
		return
	}

	writeJSONResponse(w, 201, Chirp{
		ID:        dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		Body:      dbChirp.Body,
		UserID:    userID.String(),
	})
}

func (cfg *apiConfig) handlerGetAllChirps(w http.ResponseWriter, r *http.Request) {
	authorID := r.URL.Query().Get("author_id")
	if authorID != "" {
		cfg.handlerGetChirpByUserID(w, r, authorID)
		return
	}
	dbChirps, err := cfg.db.GetAllChirps(r.Context())
	if err != nil {
		write500Error(w)
		return
	}
	res := make([]Chirp, len(dbChirps))
	for i, dbChirp := range dbChirps {
		res[i] = Chirp{
			ID:        dbChirp.ID,
			CreatedAt: dbChirp.CreatedAt,
			UpdatedAt: dbChirp.UpdatedAt,
			Body:      dbChirp.Body,
			UserID:    dbChirp.UserID.String(),
		}
	}
	sortType := r.URL.Query().Get("sort")
	if sortType == string(SortTypeDESC) {
		sort.Slice(res, func(i, j int) bool {
			return res[i].CreatedAt.After(res[j].CreatedAt)
		})
	}
	writeJSONResponse(w, 200, res)
}

func (cfg *apiConfig) handlerGetChirpByID(w http.ResponseWriter, r *http.Request) {
	parsedUUID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		write500Error(w)
		return
	}
	dbChirp, err := cfg.db.GetChirpByID(r.Context(), parsedUUID)
	if err != nil {
		write404Error(w)
		return
	}
	writeJSONResponse(w, 200, Chirp{
		ID:        dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		Body:      dbChirp.Body,
		UserID:    dbChirp.UserID.String(),
	})
}

func (cfg *apiConfig) handlerDeleteChirp(w http.ResponseWriter, r *http.Request) {
	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		write500Error(w)
		return
	}

	authToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		write401Error(w)
		return
	}
	userID, err := auth.ValidateJWT(authToken, cfg.jwtSecret)
	if err != nil {
		write403Error(w)
		return
	}

	dbChirp, err := cfg.db.GetChirpByChirpIDAndUserID(r.Context(), database.GetChirpByChirpIDAndUserIDParams{
		ID:     chirpID,
		UserID: userID,
	})
	if err != nil {
		write403Error(w)
		return
	}
	err = cfg.db.DeleteChirpByID(r.Context(), dbChirp.ID)
	if err != nil {
		write500Error(w)
		return
	}
	writeStatusCodeResponse(w, http.StatusNoContent)
}

func (cfg *apiConfig) handlerGetChirpByUserID(w http.ResponseWriter, r *http.Request, userID string) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		write500Error(w)
		return
	}
	dbChirps, err := cfg.db.GetChirpsByUserID(r.Context(), userUUID)
	if err != nil {
		write403Error(w)
		return
	}

	res := make([]Chirp, len(dbChirps))
	for i, dbChirp := range dbChirps {
		res[i] = Chirp{
			ID:        dbChirp.ID,
			CreatedAt: dbChirp.CreatedAt,
			UpdatedAt: dbChirp.UpdatedAt,
			Body:      dbChirp.Body,
			UserID:    dbChirp.UserID.String(),
		}
	}
	sortType := r.URL.Query().Get("sort")
	if sortType == string(SortTypeDESC) {
		sort.Slice(res, func(i, j int) bool {
			return res[i].CreatedAt.After(res[j].CreatedAt)
		})
	}
	writeJSONResponse(w, 200, res)
}
