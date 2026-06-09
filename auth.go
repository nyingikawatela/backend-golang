package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/nyingikachimbelengue/Chirpy-clone/internal/auth"
	"github.com/nyingikachimbelengue/Chirpy-clone/internal/database"
)

func (apiCfg *apiConfig) loginUser(w http.ResponseWriter, r *http.Request) {
	type getData struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	type resData struct {
		Id            string `json:"id"`
		Created_at    string `json:"created_at"`
		Updated_at    string `json:"updated_at"`
		Email         string `json:"email"`
		Is_chirpy_red bool   `json:"is_chirpy_red"`
		Token         string `json:"token"`
		Refresh_token string `json:"refresh_token"`
	}
	var gData getData
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&gData)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"err":"Something went wrong"}`))
		return
	}
	dataUser, err := apiCfg.dbQueries.GetUserById(r.Context(), gData.Email)

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"Incorrect email or password"}`))
		return
	}
	verifyPassword, err := auth.CheckPasswordHash(gData.Password, dataUser.HashedPassword)
	if verifyPassword == false || err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"Incorrect email or password"}`))
		return
	}
	tokenStr, err := auth.MakeJWT(dataUser.ID, apiCfg.jwtKey, time.Duration(3600)*time.Second)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"err":"error creating token"}`))
		return
	}
	randRToken := auth.MakeRefreshToken()
	refreshToken, err := apiCfg.dbQueries.CreateToken(r.Context(), database.CreateTokenParams{Token: randRToken, UserID: dataUser.ID, ExpiresAt: time.Now().UTC().Add(time.Hour * 24 * 60)})
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"err":"error creating refresh token"}`))
		return
	}
	res := resData{
		Id:            dataUser.ID.String(),
		Created_at:    dataUser.CreatedAt.Format(time.RFC3339),
		Updated_at:    dataUser.UpdatedAt.Format(time.RFC3339),
		Email:         dataUser.Email,
		Is_chirpy_red: dataUser.IsChirpyRed,
		Token:         tokenStr,
		Refresh_token: refreshToken.Token,
	}
	finallyData, err := json.Marshal(res)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"err":"error marshalling response"}`))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(finallyData)
}

func (apiCfg *apiConfig) Refresh(w http.ResponseWriter, r *http.Request) {
	type RES struct {
		Token string `json:"token"`
	}
	bearer, err := auth.GetBearerToken(r.Header)
	if err != nil || bearer == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Token invalido"))
		return
	}
	getToken, err := apiCfg.dbQueries.GetUserFromRefreshToken(r.Context(), bearer)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if getToken.ExpiresAt.Before(time.Now().UTC()) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if getToken.RevokedAt.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	tokenStr, err := auth.MakeJWT(getToken.UserID, apiCfg.jwtKey, time.Hour)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	res := RES{
		Token: tokenStr,
	}
	data, err := json.Marshal(res)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func (apiCfg *apiConfig) revokeHandler(w http.ResponseWriter, r *http.Request) {
	bearer, err := auth.GetBearerToken(r.Header)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	err = apiCfg.dbQueries.RevokeRefreshToken(r.Context(), bearer)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}