package main

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/nyingikachimbelengue/Chirpy-clone/internal/auth"
)

func (apiCfg *apiConfig) polkaWebhook(w http.ResponseWriter, r *http.Request) {
	type reqData struct {
		Event string `json:"event"`
		Data  struct {
			UserID string `json:"user_id"`
		} `json:"data"`
	}
	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if apiKey != apiCfg.polkaKey {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	var rData reqData
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&rData)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if rData.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	userID, err := uuid.Parse(rData.Data.UserID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	data, err := apiCfg.dbQueries.UpdateRedChirp(r.Context(), userID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	_ = data
	w.WriteHeader(http.StatusNoContent)
}