package main

import (
	"Chirpy/internal/database"
	"encoding/json"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
)

type chirpJson struct {
	Id uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body string `json:"body"`
	UserId uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) handleChirp(w http.ResponseWriter, req *http.Request) {
	type chirp struct {
		Body string `json:"body"`
		UserId uuid.UUID `json:"user_id"`
	}

	decoder := json.NewDecoder(req.Body)
	chirpInfo := chirp{}
	err := decoder.Decode(&chirpInfo)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if len(chirpInfo.Body) > 140 {
		respondWithError(w, 400, "Chirp too long")
		return
	}

	badWords := []string{"kerfuffle", "sharbert", "fornax"}
	arr := strings.Split(chirpInfo.Body, " ")

	final := ""
	for i, v := range arr {
		if slices.Contains(badWords, strings.ToLower(v)) {
			if i == len(arr) - 1 {
				final += "****"
			} else {
				final += "**** "
			}
		} else {
			if i == len(arr) - 1 {
				final += v
			} else {
				final += v + " "
			}
		}
	}

	createdChirp, err := cfg.database.CreateChirp(req.Context(), database.CreateChirpParams{Body: final, UserID: chirpInfo.UserId})

	if err != nil {
		respondWithError(w, 400, err.Error())
		return
	}

	respondWithJson(w, 201, chirpJson{
		Id: createdChirp.ID,
		CreatedAt: createdChirp.CreatedAt,
		UpdatedAt: createdChirp.UpdatedAt,
		Body: createdChirp.Body,
		UserId: createdChirp.UserID,
	})
}

func (cfg *apiConfig) handleGetChirps(w http.ResponseWriter, req *http.Request) {
	chirps, err := cfg.database.GetChirps(req.Context())

	if err != nil {
		respondWithError(w, 400, err.Error())
		return
	}

	 chirpsToReturn := []chirpJson{}

	 for _, chirp := range(chirps) {
		chirpsToReturn = append(chirpsToReturn, chirpJson{
			Id: chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body: chirp.Body,
			UserId: chirp.UserID,
		})
	 }


	respondWithJson(w, 200, chirpsToReturn)
}

func (cfg *apiConfig) handleGetChirp(w http.ResponseWriter, req *http.Request) {
	id, err := uuid.Parse(req.PathValue("chirpID"))

	if err != nil {
		respondWithError(w, 404, err.Error())
		return
	}

	chirp, err := cfg.database.GetChirp(req.Context(), id)

	if err != nil {
		respondWithError(w, 404, err.Error())
		return
	}

	respondWithJson(w, 200, chirpJson{
		Id: chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body: chirp.Body,
		UserId: chirp.UserID,
	})
}