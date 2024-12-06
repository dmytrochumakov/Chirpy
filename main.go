package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/dmytrochumakov/chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	envPlatform    string
	db             *database.Queries
	jwtSecret      string
	polkaKey       string
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
		return
	}

	dbURL := os.Getenv("DB_URL")
	envPlatform := os.Getenv("PLATFORM")
	jwtSecret := os.Getenv("JWT_SECRET")
	polkaKey := os.Getenv("POLKA_KEY")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
		return
	}

	const filepathRoot = "."
	const port = "8080"
	dbQueries := database.New(db)

	apiCfg := &apiConfig{
		fileserverHits: atomic.Int32{},
		envPlatform:    envPlatform,
		db:             dbQueries,
		jwtSecret:      jwtSecret,
		polkaKey:       polkaKey,
	}

	mux := http.NewServeMux()
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))))
	mux.HandleFunc("GET /api/healthz", handlerHealthz)
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)
	mux.HandleFunc("POST /api/validate_chirp", handlerValidateChirp)
	mux.HandleFunc("POST /api/users", apiCfg.handlerCreateUser)
	mux.HandleFunc("POST /api/chirps", apiCfg.handlerCreateChirp)
	mux.HandleFunc("GET /api/chirps", apiCfg.handlerGetAllChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.handlerGetChirpByID)
	mux.HandleFunc("POST /api/login", apiCfg.handlerLogin)
	mux.HandleFunc("POST /api/refresh", apiCfg.handlerRefresh)
	mux.HandleFunc("POST /api/revoke", apiCfg.handlerRevoke)
	mux.HandleFunc("PUT /api/users", apiCfg.handlerUpdateUserEmailAndPassword)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", apiCfg.handlerDeleteChirp)
	mux.HandleFunc("POST /api/polka/webhooks", apiCfg.handlerWebhooks)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving on port: %s\n", port)
	log.Fatal(server.ListenAndServe())
}

func writeCleanedBody(w http.ResponseWriter, cleanedBody string) {
	type responseCleanedBody struct {
		CleanedBody string `json:"cleaned_body"`
	}
	resp := responseCleanedBody{}
	resp.CleanedBody = cleanedBody

	w.WriteHeader(http.StatusOK)

	dat, err := json.Marshal(resp)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		write500Error(w)
		return
	}
	w.Write(dat)
}

func writeStatusCodeResponse(w http.ResponseWriter, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
}

func writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	_, writeErr := w.Write(response)
	if writeErr != nil {
		log.Printf("Error writing response: %s", writeErr)
	}
}

func writeError(w http.ResponseWriter, code int, error string) {
	type responseError struct {
		Error string `json:"error"`
	}
	respError := responseError{}

	w.WriteHeader(code)
	respError.Error = error

	dat, err := json.Marshal(respError)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		write500Error(w)
		return
	}
	w.Write(dat)
}

func write500Error(w http.ResponseWriter) {
	writeError(w, 500, "Something went wrong")
}

func write403Error(w http.ResponseWriter) {
	writeError(w, 403, "403 Forbidden")
}

func write404Error(w http.ResponseWriter) {
	writeError(w, 404, "404 Not Found")
}

func write401Error(w http.ResponseWriter) {
	writeError(w, 401, "401 Unauthorized")
}

func DecodeJSON[T any](r *http.Request, target *T) error {
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	return decoder.Decode(target)
}
