package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
)

func (cfg *apiConfig) handleCreateUser(w http.ResponseWriter, req *http.Request) {
	type recievedBody struct {
		Email string `json:"email"`
	}

	type returnBody struct {
		Id uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email string `json:"email"`
	}

	decoder := json.NewDecoder(req.Body)
	user := recievedBody{}
	err := decoder.Decode(&user)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	createdUser, err := cfg.database.CreateUser(req.Context(), user.Email)

	if err != nil {
		respondWithError(w, 400, err.Error())
		return
	}

	respondWithJson(w, 201, returnBody{
		Id: createdUser.ID,
		CreatedAt: createdUser.CreatedAt,
		UpdatedAt: createdUser.UpdatedAt,
		Email: createdUser.Email,
	})
}