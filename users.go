package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/nyingikachimbelengue/Chirpy-clone/internal/auth"
	"github.com/nyingikachimbelengue/Chirpy-clone/internal/database"
	"github.com/nyingikachimbelengue/Chirpy-clone/internal/response"
)

func (apiCfg *apiConfig) createUsers(w http.ResponseWriter, r *http.Request) {
	type Userdata struct {
		Email    string
		Password string
	}
	var Udata Userdata
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&Udata)
	if err != nil {
	}
	password := auth.HashPassword(Udata.Password)
	data, err := apiCfg.dbQueries.CreateUser(r.Context(), database.CreateUserParams{Email: Udata.Email, HashedPassword: password})
	if err != nil {
	}
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	res := response.UserResponse{
		Id:            data.ID.String(),
		Created_at:    data.CreatedAt.String(),
		Updated_at:    data.UpdatedAt.String(),
		Email:         data.Email,
		Is_chirpy_red: data.IsChirpyRed,
	}
	dat, err := json.Marshal(res)
	if err != nil {
	}
	w.Write(dat)
}

func (apiCfg *apiConfig) updateUserHandler(w http.ResponseWriter, r *http.Request) {
	type reqData struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
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
	var rData reqData
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&rData)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	hashedPassword := auth.HashPassword(rData.Password)

	data, err := apiCfg.dbQueries.UpdateUser(r.Context(), database.UpdateUserParams{
		Email:          rData.Email,
		HashedPassword: hashedPassword,
		ID:             userID,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	res := response.UserResponse{
		Id:            data.ID.String(),
		Created_at:    data.CreatedAt.Format(time.RFC3339),
		Updated_at:    data.UpdatedAt.Format(time.RFC3339),
		Email:         data.Email,
		Is_chirpy_red: data.IsChirpyRed,
	}

	dat, err := json.Marshal(res)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(dat)
}