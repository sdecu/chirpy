package main

import (
	"encoding/json"
	"fmt"
	"github.com/sdecu/chirpy/internal/database"
	"log"
	"net/http"
	"strings"
)

func main() {
	startServer()
}

func startServer() {
	port := "8080"
	mux := http.NewServeMux()
	db, err := database.NewDB("./database.json")
	if err != nil {
		log.Fatal(err)
	}

	apiCfg := &apiConfig{
		fileserverHits: 0,
		DB:             db,
	}

	healthzHandler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}

	readinessDir := "GET /api/healthz"
	mux.HandleFunc(readinessDir, healthzHandler)

	filepathRoot := "/app/"
	dir := http.Dir(".")
	fileserver := http.FileServer(dir)
	handler := http.StripPrefix(filepathRoot, apiCfg.middlewareMetricsInc(fileserver))
	mux.Handle(filepathRoot, handler)

	mux.HandleFunc("GET /admin/metrics", http.HandlerFunc(apiCfg.writeHit))
	mux.Handle("/api/reset", http.HandlerFunc(apiCfg.reset))

	//json handler
	mux.HandleFunc("POST /api/chirps", func(w http.ResponseWriter, r *http.Request) {
		handleCreateChirp(w, r, db)
	})
	mux.HandleFunc("GET /api/chirps", apiCfg.handlerChirpsRetrieve)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}

type apiConfig struct {
	fileserverHits int
	DB             *database.DB
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits += 1
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) writeHit(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`
        <html>
        <body>
        <h1>Welcome, Chirpy Admin</h1>
        <p>Chirpy has been visited %d times!</p>
        </body>
        </html>`, cfg.fileserverHits)))
}

func (cfg *apiConfig) reset(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits = 0
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits successfully reset"))
}

func respondWithError(w http.ResponseWriter, message string, statusCode int) {
	response := map[string]int{message: statusCode}
	dat, _ := json.Marshal(response)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(dat)
}

func respondWithJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(response)
}

func cleanString(message string) string {
	dirty := strings.Fields(message)
	for i, sub := range dirty {
		lowered := strings.ToLower(sub)
		if lowered == "kerfuffle" || lowered == "sharbert" || lowered == "fornax" {
			dirty[i] = "****"
		}
	}
	clean := strings.Join(dirty, " ")
	return clean
}

func handleCreateChirp(w http.ResponseWriter, r *http.Request, db *database.DB) {
	type parameters struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, "Couldn't decode parameters", http.StatusInternalServerError)
		return
	}

	const maxChirpLength = 140
	if len(params.Body) > maxChirpLength {
		respondWithError(w, "Chirp is too long", http.StatusBadRequest)
		return
	}

	newChirp, err := db.CreateChirp(params.Body)
	if err != nil {
		respondWithError(w, "Couldn't create chirp", http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusCreated, newChirp)
}
