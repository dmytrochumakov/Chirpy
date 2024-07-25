package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"	
	"strings"
)

func main() {
	const filepathRoot = "."
	const port = "8080"

	apiCfg := apiConfig{
		fileserverHits: 0,
	}

	mux := http.NewServeMux()
	mux.Handle("/app/*", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))))
	mux.HandleFunc("GET /api/healthz" , healthz)
	mux.HandleFunc("GET /admin/metrics" , apiCfg.handlerMetrics)
	mux.HandleFunc("/api/reset" , apiCfg.resetFileserverHits)
	mux.HandleFunc("POST /api/validate_chirp" , validate_chirp)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving on port: %s\n", port)
	log.Fatal(srv.ListenAndServe())
}

func healthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

type apiConfig struct {
	fileserverHits int
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits++
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig)handlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`
<html>

<body>
	<h1>Welcome, Chirpy Admin</h1>
	<p>Chirpy has been visited %d times!</p>
</body>

</html>
	`, cfg.fileserverHits)))
}

func (cfg *apiConfig) resetFileserverHits(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits = 0
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits reset to 0"))
}

func validate_chirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {        
        Body string `json:"body"`        
    }
	type returnVals struct {
		Valid bool `json:"valid"`
	}
	type returnCleanBodyVals struct {
		CleanBody string `json:"cleaned_body"`
	}

    decoder := json.NewDecoder(r.Body)
    params := parameters{}
    err := decoder.Decode(&params)
    if err != nil {		
		respondWithError(w, http.StatusBadRequest, "Couldn't decode parameters")
		return
    }

	const maxChirpLength = 140
	if len(params.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	lowerBody := strings.ToLower(params.Body)	
	badWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}
	if strings.Contains(lowerBody, "kerfuffle") || strings.Contains(lowerBody, "sharbert") || strings.Contains(lowerBody, "fornax") {
		
		replacedBody := getCleanedBody(params.Body, badWords)
		respondWithJSON(w, http.StatusOK, returnCleanBodyVals{
			CleanBody: replacedBody,	
		})
		return
	} else {
		respondWithJSON(w, http.StatusOK, returnCleanBodyVals{
			CleanBody: params.Body,
		})
		return
	}

	respondWithJSON(w, http.StatusOK, returnVals{
		Valid: true,
	})
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	if code > 499 {
		log.Printf("Responding with 5XX error: %s", msg)
	}
	type errorResponse struct {
		Error string `json:"error"`
	}
	respondWithJSON(w, code, errorResponse{
		Error: msg,
	})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	dat, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(code)
	w.Write(dat)
}

func getCleanedBody(body string, badWords map[string]struct{}) string {
	words := strings.Split(body, " ")
	for i, word := range words {
		loweredWord := strings.ToLower(word)
		if _, ok := badWords[loweredWord]; ok {
			words[i] = "****"
		}
	}
	cleaned := strings.Join(words, " ")
	return cleaned
}