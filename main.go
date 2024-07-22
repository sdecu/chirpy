package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	startServer()
}

func startServer() {
	port := "8080"
	mux := http.NewServeMux()
	apiCfg := &apiConfig{
		fileserverHits: 0,
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

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}

type apiConfig struct {
	fileserverHits int
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
