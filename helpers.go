package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func respondWithError( w http.ResponseWriter, code int, msg string) {
	type error struct {
		Error string `json:"error"`
	}

	errBody := error {
		Error: msg,
	}

	respondWithJson(w, code, errBody)
}

func respondWithJson(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	dat, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(code)
	w.Write(dat)
}