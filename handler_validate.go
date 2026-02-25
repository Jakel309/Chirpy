package main

import (
	"encoding/json"
	"net/http"
	"slices"
	"strings"
)

func handleChirp(w http.ResponseWriter, req *http.Request) {
	type chirp struct {
		Body string `json:"body"`
	}

	type returnBody struct {
		CleanedBody string `json:"cleaned_body"`
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

	respondWithJson(w, 200, returnBody{
		CleanedBody: final,
	})
}