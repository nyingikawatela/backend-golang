package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync/atomic"

	"time"

	"github.com/google/uuid"
	_ "github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/nyingikachimbelengue/Chirpy-clone/internal/auth"
	"github.com/nyingikachimbelengue/Chirpy-clone/internal/database"
	_ "golang.org/x/tools/go/cfg"
)

type apiConfig struct {
	dbQueries      *database.Queries
	fileserverHits atomic.Int32
	platform       string
	jwtKey         string
}

type error struct {
	Err string `json:"err"`
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}
func healthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	html := fmt.Sprintf(`
		<html>
			<body>
				<h1>Welcome, Chirpy Admin</h1>
				<p>Chirpy has been visited %d times!</p>
			</body>
		</html>
	`, cfg.fileserverHits.Load())
	w.Write([]byte(html))
}

func (cfg *apiConfig) resetHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("platform:", cfg.platform)
	if cfg.platform != "dev" {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	err := cfg.dbQueries.DeleteAllUsers(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func errorAux(msg string, w http.ResponseWriter) {
	res := error{
		Err: msg,
	}
	dat, _ := json.Marshal(res)
	w.WriteHeader(http.StatusBadRequest)
	w.Write(dat)
}

func (apiCfg *apiConfig) createUsers(w http.ResponseWriter, r *http.Request) {
	type Userdata struct {
		Email    string
		Password string
	}
	type resData struct {
		Id         string `json:"id"`
		Created_at string `json:"created_at"`
		Updated_at string `json:"updated_at"`
		Email      string `json:"email"`
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
	res := resData{
		Id:         data.ID.String(),
		Created_at: data.CreatedAt.String(),
		Updated_at: data.UpdatedAt.String(),
		Email:      data.Email,
	}
	dat, err := json.Marshal(res)
	if err != nil {
	}
	w.Write(dat)
}

func (apiCfg *apiConfig) chirpsHandler(w http.ResponseWriter, r *http.Request) {
	type s_requestData struct {
		Body   string `json:"body"`
		UserID string `json:"user_id"`
	}
	type resData struct {
		Id         string `json:"id"`
		Created_at string `json:"created_at"`
		Updated_at string `json:"updated_at"`
		Body       string `json:"body"`
		User_id    string `json:"user_id"`
	}
	bearer, err := auth.GetBearerToken(r.Header)
	if err != nil{
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Token invalido"))
		return
	}
	isValid, err := auth.ValidateJWT(bearer, apiCfg.jwtKey)
	if isValid == uuid.Nil || err != nil{
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
	fmt.Println(err)
	if err != nil {
		errorAux("error creating chirp", w)
		return
	}

	res := resData{
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
	type resChirp struct {
		Id         string `json:"id"`
		Created_at string `json:"created_at"`
		Updated_at string `json:"updated_at"`
		Body       string `json:"body"`
		User_id    string `json:"user_id"`
	}
	data, err := apiCfg.dbQueries.GetAllChirps(r.Context())
	if err != nil {
		errorAux("error fetching chirp", w)
		return
	}
	var Chirps []resChirp
	for _, chirp := range data {
		Chirps = append(Chirps, resChirp{
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
	type resChirp struct {
		Id         string `json:"id"`
		Created_at string `json:"created_at"`
		Updated_at string `json:"updated_at"`
		Body       string `json:"body"`
		User_id    string `json:"user_id"`
	}
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
	tmpChirp := resChirp{
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

func (apiCfg *apiConfig) loginUser(w http.ResponseWriter, r *http.Request) {
	type getData struct {
		Password         string `json:"password"`
		Email            string `json:"email"`
	}
	type resData struct {
		Id         string `json:"id"`
		Created_at string `json:"created_at"`
		Updated_at string `json:"updated_at"`
		Email      string `json:"email"`
		Token      string `json:"token"`
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
		fmt.Println(err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"Incorrect email or password"}`))
		return
	}
	verifyPassword, err := auth.CheckPasswordHash(gData.Password, dataUser.HashedPassword)
	if verifyPassword == false || err != nil {
		fmt.Println(err)
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
	refreshToken, err := apiCfg.dbQueries.CreateToken(r.Context(), database.CreateTokenParams{Token: randRToken, UserID: dataUser.ID, ExpiresAt: time.Now().UTC().Add(time.Hour * 24 * 60),})
	if err != nil{
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"err":"error creating refresh token"}`))
		return
	}
	res := resData{
		Id:         dataUser.ID.String(),
		Created_at: dataUser.CreatedAt.Format(time.RFC3339),
		Updated_at: dataUser.UpdatedAt.Format(time.RFC3339),
		Email:      dataUser.Email,
		Token:      tokenStr,
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

func (apiCfg *apiConfig) Refresh(w http.ResponseWriter, r *http.Request){
	type RES struct{
		Token string `json:"token"`
	}
	bearer, err := auth.GetBearerToken(r.Header)
	if err != nil || bearer == ""{
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
	if err != nil{
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
func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, _ := sql.Open("postgres", dbURL)

	apiCfg := apiConfig{
		dbQueries: database.New(db),
		platform:  os.Getenv("PLATFORM"),
		jwtKey:    os.Getenv("JWT"),
	}
	mux := http.NewServeMux()
	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	srcImage := http.Dir("./assets")
	serverImage := http.FileServer(srcImage)
	filterPath := http.StripPrefix("/app/assets", apiCfg.middlewareMetricsInc(serverImage))
	mux.Handle("/app/assets/logo.png", filterPath)
	mux.Handle("/app/", http.StripPrefix("/app/", apiCfg.middlewareMetricsInc(http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /admin/metrics", apiCfg.metricsHandler)
	mux.HandleFunc("POST /admin/reset", apiCfg.resetHandler)
	mux.HandleFunc("GET /api/healthz", healthz)
	mux.HandleFunc("POST /api/users", apiCfg.createUsers)
	mux.HandleFunc("POST /api/chirps", apiCfg.chirpsHandler)
	mux.HandleFunc("GET /api/chirps", apiCfg.getAllChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.getChirpById)
	mux.HandleFunc("POST /api/login", apiCfg.loginUser)
	mux.HandleFunc("POST /api/refresh", apiCfg.Refresh)
	mux.HandleFunc("POST /api/revoke", apiCfg.revokeHandler)
	fmt.Println("Server running on port 8080")
	server.ListenAndServe()
}
