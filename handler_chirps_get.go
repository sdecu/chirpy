package main

import (
	db "github.com/sdecu/chirpy/internal/database" // Alias the import
	"net/http"
	"sort"
)

func (cfg *apiConfig) handlerChirpsRetrieve(w http.ResponseWriter, r *http.Request) {
	dbChirps, err := cfg.DB.GetChirps()
	if err != nil {
		respondWithError(w, "Couldn't retrieve chirps", http.StatusInternalServerError)
		return
	}

	chirps := []db.Chirp{}
	for _, dbChirp := range dbChirps {
		chirps = append(chirps, db.Chirp{
			ID:   dbChirp.ID,
			Body: dbChirp.Body,
		})
	}

	sort.Slice(chirps, func(i, j int) bool {
		return chirps[i].ID < chirps[j].ID
	})

	respondWithJSON(w, http.StatusOK, chirps)
}
