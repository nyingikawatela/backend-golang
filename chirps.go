package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/nyingikachimbelengue/Chirpy-clone/internal/auth"
	"github.com/nyingikachimbelengue/Chirpy-clone/internal/database"
	"github.com/nyingikachimbelengue/Chirpy-clone/internal/response"
	_ "github.com/nyingikachimbelengue/Chirpy-clone/internal/response"
)

func (apiCfg *apiConfig) chirpsHandler(w http.ResponseWriter, r *http.Request) {
	type s_requestData struct {
		Body   string `json:"body"`
		UserID string `json:"user_id"`
	}
	bearer, err := auth.GetBearerToken(r.Header)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Token invalido"))
		return
	}
	isValid, err := auth.ValidateJWT(bearer, apiCfg.jwtKey)
	if isValid == uuid.Nil || err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Token invalido"))
		return
	}
	var Rdata s_requestData
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&Rdata)
	if err != nil {
		errorAux("Something went wrong", w)
		return
	}

	if len(Rdata.Body) > 140 {
		errorAux("Chirp is too long", w)
		return
	}

	userID := isValid

	data, err := apiCfg.dbQueries.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   Rdata.Body,
		UserID: userID,
	})
	if err != nil {
		errorAux("error creating chirp", w)
		return
	}

	res := response.ChirpResponse{
		Id:         data.ID.String(),
		Created_at: data.CreatedAt.Format(time.RFC3339),
		Updated_at: data.UpdatedAt.Format(time.RFC3339),
		Body:       data.Body,
		User_id:    data.UserID.String(),
	}

	dat, err := json.Marshal(res)
	if err != nil {
		errorAux("error marshalling response", w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(dat)
}

func (apiCfg *apiConfig) getAllChirps(w http.ResponseWriter, r *http.Request) {
	s := r.URL.Query().Get("author_id")
	if s != "" {
		authorId, _ := uuid.Parse(s)
		data, err := apiCfg.dbQueries.GetChirpByAuthor(r.Context(), authorId)
		if err != nil {
			errorAux("error marshalling response", w)
			return
		}
		var Chirps []response.ChirpResponse
		for _, chirp := range data {
			Chirps = append(Chirps, response.ChirpResponse{
				Id:         chirp.ID.String(),
				Created_at: chirp.CreatedAt.Format(time.RFC3339),
				Updated_at: chirp.UpdatedAt.Format(time.RFC3339),
				Body:       chirp.Body,
				User_id:    chirp.UserID.String(),
			})
		}
		res, err := json.Marshal(Chirps)
		if err != nil {
			errorAux("error marshalling response", w)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(res)
		return
	}
	data, err := apiCfg.dbQueries.GetAllChirps(r.Context())

	if err != nil {
		errorAux("error fetching chirp", w)
		return
	}
	var Chirps []response.ChirpResponse
	for _, chirp := range data {
		Chirps = append(Chirps, response.ChirpResponse{
			Id:         chirp.ID.String(),
			Created_at: chirp.CreatedAt.Format(time.RFC3339),
			Updated_at: chirp.UpdatedAt.Format(time.RFC3339),
			Body:       chirp.Body,
			User_id:    chirp.UserID.String(),
		})
	}
	dat, err := json.Marshal(Chirps)
	if err != nil {
		errorAux("error marshalling response", w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(dat)
}

func (apiCfg *apiConfig) getChirpById(w http.ResponseWriter, r *http.Request) {
	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if chirpID == uuid.Nil || err != nil {
		errorAux("Nenhum chirp passado", w)
		return
	}
	data, err := apiCfg.dbQueries.GetChirpById(r.Context(), chirpID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"err":"chirp not found"}`))
		return
	}
	tmpChirp := response.ChirpResponse{
		Id:         data.ID.String(),
		Created_at: data.CreatedAt.Format(time.RFC3339),
		Updated_at: data.UpdatedAt.Format(time.RFC3339),
		Body:       data.Body,
		User_id:    data.UserID.String(),
	}
	res, err := json.Marshal(tmpChirp)
	if err != nil {
		errorAux("Error", w)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(res)
}

func (apiCfg *apiConfig) deleteChirpHandler(w http.ResponseWriter, r *http.Request) {
	bearer, err := auth.GetBearerToken(r.Header)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	userID, err := auth.ValidateJWT(bearer, apiCfg.jwtKey)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// buscar chirp
	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	chirp, err := apiCfg.dbQueries.GetChirpById(r.Context(), chirpID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// verificar se é o autor
	if chirp.UserID != userID {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// apagar
	err = apiCfg.dbQueries.DeleteChirp(r.Context(), chirpID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}