package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/dmytrochumakov/chirpy/internal/auth"
	"github.com/dmytrochumakov/chirpy/internal/database"
)

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type requestParams struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	reqParams := requestParams{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&reqParams)
	if err != nil {
		write500Error(w)
		return
	}
	dbUser, err := cfg.db.GetUserByEmail(r.Context(), reqParams.Email)
	if err != nil {
		write404Error(w)
		return
	}
	err = auth.CheckPasswordHash(dbUser.HashedPassword, reqParams.Password)
	if err != nil {
		write401Error(w)
		return
	}
	accessTokenExpirationTime := time.Hour

	accessToken, err := auth.MakeJWT(
		dbUser.ID,
		cfg.jwtSecret,
		accessTokenExpirationTime,
	)
	if err != nil {
		write500Error(w)
		return
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		write500Error(w)
		return
	}
	dbRefreshToken, err := cfg.db.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     refreshToken,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID:    dbUser.ID,
		ExpiresAt: time.Now().Add(60 * 24 * time.Hour),
		RevokedAt: sql.NullTime{},
	})
	if err != nil {
		write500Error(w)
		return
	}

	type response struct {
		User
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}
	writeJSONResponse(w, 200, response{
		User: User{
			ID:          dbUser.ID,
			CreatedAt:   dbUser.CreatedAt,
			UpdatedAt:   dbUser.UpdatedAt,
			Email:       dbUser.Email,
			IsChirpyRed: dbUser.IsChirpyRed,
		},
		Token:        accessToken,
		RefreshToken: dbRefreshToken.Token,
	})
}

func (cfg *apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Token string `json:"token"`
	}
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		write401Error(w)
		return
	}

	dbUser, err := cfg.db.GetUserFromRefreshToken(r.Context(), refreshToken)
	if err != nil {
		write401Error(w)
		return
	}

	authToken, err := auth.MakeJWT(dbUser.ID, cfg.jwtSecret, time.Hour)
	if err != nil {
		write401Error(w)
		return
	}

	writeJSONResponse(w, http.StatusOK, response{
		Token: authToken,
	})
}

func (cfg *apiConfig) handlerRevoke(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		write401Error(w)
		return
	}

	dbUser, err := cfg.db.GetUserFromRefreshToken(r.Context(), refreshToken)
	if err != nil {
		write401Error(w)
		return
	}

	err = cfg.db.RevokeRefreshToken(r.Context(), database.RevokeRefreshTokenParams{
		RevokedAt: sql.NullTime{
			Time:  time.Now().UTC(),
			Valid: true,
		},
		UpdatedAt: time.Now().UTC(),
		UserID:    dbUser.ID,
	})
	if err != nil {
		write401Error(w)
		return
	}

	writeStatusCodeResponse(w, http.StatusNoContent)
}
