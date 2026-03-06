package main

import (
	"Chirpy/internal"
	"Chirpy/internal/database"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"sort"
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
	}

	decoder := json.NewDecoder(req.Body)
	chirpInfo := chirp{}
	err := decoder.Decode(&chirpInfo)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	token, err := internal.GetBearerToken(req.Header)

	if err != nil {
		fmt.Println(err)
		respondWithError(w, 401, "Unauthorized")
		return
	}

	userId, err := internal.ValidateJWT(token, cfg.secret)

	if err != nil{
		fmt.Println(err)
		respondWithError(w, 401, "Unauthorized")
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

	createdChirp, err := cfg.database.CreateChirp(req.Context(), database.CreateChirpParams{Body: final, UserID: userId})

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
	authorId := req.URL.Query().Get("author_id")
	sortOrder := req.URL.Query().Get("sort")

	chirps, err := cfg.database.GetChirps(req.Context())

	if err != nil {
		respondWithError(w, 400, err.Error())
		return
	}

	chirpsToReturn := []chirpJson{}

	for _, chirp := range(chirps) {
		if authorId != "" {
			if chirp.UserID.String() != authorId {
				continue
			}
		}
		chirpsToReturn = append(chirpsToReturn, chirpJson{
			Id: chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body: chirp.Body,
			UserId: chirp.UserID,
		})
	}

	if sortOrder == "desc" {
		sort.Slice(chirpsToReturn, func(i, j int) bool {return chirpsToReturn[i].CreatedAt.After(chirpsToReturn[j].CreatedAt)})
	} else {
		sort.Slice(chirpsToReturn, func(i, j int) bool {return chirpsToReturn[i].CreatedAt.Before(chirpsToReturn[j].CreatedAt)})
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

func (cfg *apiConfig) handleDeleteChirp(w http.ResponseWriter, req *http.Request) {
	id, err := uuid.Parse(req.PathValue("chirpID"))

	if err != nil {
		respondWithError(w, 404, err.Error())
		return
	}

	token, err := internal.GetBearerToken(req.Header)

	if err != nil {
		respondWithError(w, 401, "Unauthorized")
		return
	}

	userId, err := internal.ValidateJWT(token, cfg.secret)

	if err != nil {
		respondWithError(w, 401, "Unauthorized")
		return
	}

	chirp, err := cfg.database.GetChirp(req.Context(), id)

	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			respondWithError(w, 404, "Chirp not found")
			return
		} else {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	if chirp.UserID != userId {
		respondWithError(w, 403, "Unauthorized")
		return
	}

	_, err = cfg.database.DeleteChirp(req.Context(), id)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(204)
}