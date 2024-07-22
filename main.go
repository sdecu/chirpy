package main

import (
	"log"
	"net/http"
	"strconv"
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

	readinessDir := "/healthz"
	mux.HandleFunc(readinessDir, healthzHandler)

	filepathRoot := "/app/"
	dir := http.Dir(".")
	fileserver := http.FileServer(dir)
	handler := http.StripPrefix(filepathRoot, apiCfg.middlewareMetricsInc(fileserver))
	mux.Handle(filepathRoot, handler)

	mux.Handle("/metrics", http.HandlerFunc(apiCfg.metrics))
	mux.Handle("/reset", http.HandlerFunc(apiCfg.reset))

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
		cfg.fileserverHits++
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) metrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits: " + strconv.Itoa(cfg.fileserverHits)))
}

func (cfg *apiConfig) reset(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits = 0
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits successfully reset"))
}
