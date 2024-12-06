package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

func handlerHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	html := `
		<html>
		<body>
			<h1>Welcome, Chirpy Admin</h1>
			<p>Chirpy has been visited %d times!</p>
		</body>
		</html>
	`
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(html, cfg.fileserverHits.Load())))
}

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	if cfg.envPlatform != "dev" {
		write403Error(w)
		return
	}
	err := cfg.db.DeleteAllUsers(r.Context())
	if err != nil {
		write500Error(w)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	cfg.fileserverHits.Store(0)
	w.Write([]byte(fmt.Sprintf("Hits: %d", cfg.fileserverHits.Load())))
}

func handlerValidateChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)

	if err != nil {
		log.Printf("Error decoding parameters %s", err)
		write500Error(w)
		return
	}

	if len(params.Body) > 140 {
		log.Println("params.Body is too long")
		writeError(w, 400, "Chirp is too long")
		return
	}

	profaneWords := []string{"kerfuffle", "sharbert", "fornax"}
	splitBody := strings.Split(params.Body, " ")
	cleanedBodyArray := []string{}

	for _, str := range splitBody {
		newStr := str
		for _, profaneWord := range profaneWords {
			if strings.ToLower(str) == profaneWord {
				newStr = "****"
			}
		}
		cleanedBodyArray = append(cleanedBodyArray, newStr)
	}
	cleanedBody := strings.Join(cleanedBodyArray, " ")
	writeCleanedBody(w, cleanedBody)
}
