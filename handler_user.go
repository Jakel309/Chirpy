package main

import (
	"Chirpy/internal"
	"Chirpy/internal/database"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

func (cfg *apiConfig) handleCreateUser(w http.ResponseWriter, req *http.Request) {
	type recievedBody struct {
		Email string `json:"email"`
		Password string `json:"password"`
	}

	type returnBody struct {
		Id uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email string `json:"email"`
		IsChirpyRed bool `json:"is_chirpy_red"`
	}

	decoder := json.NewDecoder(req.Body)
	user := recievedBody{}
	err := decoder.Decode(&user)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	hash, err := internal.HashPassword(user.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}


	createdUser, err := cfg.database.CreateUser(req.Context(), database.CreateUserParams{Email: user.Email, HashedPassword: hash})

	if err != nil {
		respondWithError(w, 400, err.Error())
		return
	}

	respondWithJson(w, 201, returnBody{
		Id: createdUser.ID,
		CreatedAt: createdUser.CreatedAt,
		UpdatedAt: createdUser.UpdatedAt,
		Email: createdUser.Email,
		IsChirpyRed: createdUser.IsChirpyRed,
	})
}

func (cfg *apiConfig) handleUserLogin(w http.ResponseWriter, req *http.Request) {
	type recievedBody struct {
		Email string `json:"email"`
		Password string `json:"password"`
	}

	type returnBody struct {
		Id uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email string `json:"email"`
		Token string `json:"token"`
		RefreshToken string `json:"refresh_token"`
		IsChirpyRed bool `json:"is_chirpy_red"`
	}

	decoder := json.NewDecoder(req.Body)
	user := recievedBody{}
	err := decoder.Decode(&user)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	foundUser, err := cfg.database.GetUser(req.Context(), user.Email)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			respondWithError(w, 401, "Incorrect email or password")
			return
		} else {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	match, err := internal.CheckPasswordHash(user.Password, foundUser.HashedPassword)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if !match {
		respondWithError(w, 401, "Incorrect email or password")
		return
	}

	duration := 1 * time.Hour

	token, err := internal.MakeJWT(foundUser.ID, cfg.secret, duration)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}

	refreshToken := internal.MakeRefreshToken()

	_, err = cfg.database.CreateRefreshToken(req.Context(), database.CreateRefreshTokenParams{
		Token: refreshToken,
		UserID: foundUser.ID,
		ExpiresAt: time.Now().Add(60 * time.Hour * 24),
	})

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}

	respondWithJson(w, 200, returnBody{
		Id: foundUser.ID,
		CreatedAt: foundUser.CreatedAt,
		UpdatedAt: foundUser.UpdatedAt,
		Email: foundUser.Email,
		Token: token,
		RefreshToken: refreshToken,
		IsChirpyRed: foundUser.IsChirpyRed,
	})
}

func (cfg *apiConfig) handleUserUpdate(w http.ResponseWriter, req *http.Request) {
	type receivedBody struct {
		Password string `json:"password"`
		Email string `json:"email"`
	}

	type returnBody struct {
		Id uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email string `json:"email"`
		IsChirpyRed bool `json:"is_chirpy_red"`
	}

	decoder := json.NewDecoder(req.Body)
	user := receivedBody{}
	err := decoder.Decode(&user)
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
		respondWithError(w, 401, "Unauthorized")
		return
	}

	hash, err := internal.HashPassword(user.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	updatedUser, err := cfg.database.UpdateUser(req.Context(), database.UpdateUserParams{
		ID: userId,
		Email: user.Email,
		HashedPassword: hash,
	})

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJson(w, 200, returnBody{
		Id: userId,
		CreatedAt: updatedUser.CreatedAt,
		UpdatedAt: updatedUser.UpdatedAt,
		Email: updatedUser.Email,
		IsChirpyRed: updatedUser.IsChirpyRed,
	})
}

func (cfg *apiConfig) handleRefreshToken(w http.ResponseWriter, req *http.Request) {
	type ReturnBody struct {
		Token string `json:"token"`
	}

	refreshToken, err := internal.GetBearerToken(req.Header)

	if err != nil {
		respondWithError(w, 401, "Unauthorized")
		return
	}

	foundToken, err := cfg.database.GetRefreshToken(req.Context(), refreshToken)
	fmt.Printf(`Found Token: %v`, foundToken)

	if err != nil {
		respondWithError(w, 401, err.Error())
		return
	}

	if foundToken.RevokedAt.Valid {
		respondWithError(w, 401, "Unauthorized")
		return
	}

	duration := 1 * time.Hour

	token, err := internal.MakeJWT(foundToken.UserID, cfg.secret, duration)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJson(w, 200, ReturnBody{
		Token: token,
	})
}

func (cfg *apiConfig) handleRevokeToken(w http.ResponseWriter, req *http.Request) {
	refreshToken, err := internal.GetBearerToken(req.Header)

	if err != nil {
		respondWithError(w, 401, "Unauthorized")
		return
	}

	foundToken, err := cfg.database.GetRefreshToken(req.Context(), refreshToken)

	if err != nil || foundToken.RevokedAt.Valid {
		respondWithError(w, 401, err.Error())
		return
	}

	_, err = cfg.database.RevokeToken(req.Context(), refreshToken)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(204)
}

func (cfg *apiConfig) handleUpgradeUser(w http.ResponseWriter, req *http.Request) {
	apiKey, err := internal.GetAPIKey(req.Header)
	if err != nil {
		respondWithError(w, 401, "Unauthorized")
		return
	}

	if apiKey != cfg.polkaKey {
		respondWithError(w, 401, "Unauthorized")
		return
	}

	type data struct {
		UserId string `json:"user_id"`
	}
	type receivedBody struct {
		Event string `json:"event"`
		Data data `json:"data"`
	}

	decoder := json.NewDecoder(req.Body)
	info := receivedBody{}
	err = decoder.Decode(&info)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if info.Event != "user.upgraded" {
		w.WriteHeader(204)
		return
	}

	id, err := uuid.Parse(info.Data.UserId)
	
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	_, err = cfg.database.UpgradeUser(req.Context(), id)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			respondWithError(w, 404, "User not found")
			return
		} else {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	w.WriteHeader(204)

}