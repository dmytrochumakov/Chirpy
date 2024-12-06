package main

import (
	"net/http"

	"github.com/dmytrochumakov/chirpy/internal/auth"
	"github.com/dmytrochumakov/chirpy/internal/database"
	"github.com/google/uuid"
)

type EventType string

const (
	EventTypeUserUpgraded EventType = "user.upgraded"
)

func (cfg *apiConfig) handlerWebhooks(w http.ResponseWriter, r *http.Request) {
	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		write401Error(w)
		return
	}
	if apiKey != cfg.polkaKey {
		write401Error(w)
		return
	}
	type ParametersData struct {
		UserID string `json:"user_id"`
	}
	type parameters struct {
		Event string         `json:"event"`
		Data  ParametersData `json:"data"`
	}
	params := parameters{}
	err = DecodeJSON(r, &params)
	if err != nil {
		write500Error(w)
		return
	}
	userID, err := uuid.Parse(params.Data.UserID)
	if err != nil {
		write500Error(w)
		return
	}
	if params.Event == string(EventTypeUserUpgraded) {
		_, err := cfg.db.UpdateUserChirpyRedByUserID(r.Context(), database.UpdateUserChirpyRedByUserIDParams{
			IsChirpyRed: true,
			ID:          userID,
		})
		if err != nil {
			write404Error(w)
			return
		}
		writeStatusCodeResponse(w, http.StatusNoContent)
	} else {
		writeStatusCodeResponse(w, http.StatusNoContent)
		return
	}
}
