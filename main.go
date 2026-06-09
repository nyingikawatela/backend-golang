package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync/atomic"

	_ "time"

	_ "github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	_ "github.com/nyingikachimbelengue/Chirpy-clone/internal/auth"
	"github.com/nyingikachimbelengue/Chirpy-clone/internal/database"
	_ "golang.org/x/tools/go/cfg"
)

type apiConfig struct {
	dbQueries      *database.Queries
	fileserverHits atomic.Int32
	platform       string
	jwtKey         string
	polkaKey       string
}

type errorResponse struct {
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
	res := errorResponse{
		Err: msg,
	}
	dat, _ := json.Marshal(res)
	w.WriteHeader(http.StatusBadRequest)
	w.Write(dat)
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, _ := sql.Open("postgres", dbURL)

	apiCfg := apiConfig{
		dbQueries: database.New(db),
		platform:  os.Getenv("PLATFORM"),
		jwtKey:    os.Getenv("JWT"),
		polkaKey:  os.Getenv("POLKA_KEY"),
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
	mux.HandleFunc("PUT /api/users", apiCfg.updateUserHandler)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", apiCfg.deleteChirpHandler)
	mux.HandleFunc("POST /api/polka/webhooks", apiCfg.polkaWebhook)
	fmt.Println("Server running on port 8080")
	server.ListenAndServe()
}