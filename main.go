package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync/atomic"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/nyingikachimbelengue/Chirpy-clone/internal/database"

	_ "golang.org/x/tools/go/cfg"
)



type apiConfig struct {
	dbQueries *database.Queries
	fileserverHits atomic.Int32
	platform string
}
	type parameter struct{
		Body string `json:"body"`
	}
	type error struct{
		Err string `json:"err"`
	}
	type resp struct{
		Cleaned_body string `json:"cleaned_body"`
	}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request){
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}
func healthz(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request){
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

func (cfg *apiConfig) resetHandler(w http.ResponseWriter, r *http.Request){
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

func errorAux(msg string, w http.ResponseWriter){
	res := error{
		Err: msg, 
	}
	dat, _ := json.Marshal(res)
	w.WriteHeader(http.StatusBadRequest)
	w.Write(dat)
}

func cleanOutput(s string) string{
	div := strings.Split(s, " ")
	
	for i, t := range div{
		w := strings.ToLower(t)
		if w == "kerfuffle"{
			div[i] = strings.ReplaceAll(w, "kerfuffle", "****")
		}
		if w == "sharbert"{
			div[i] = strings.ReplaceAll(w, "sharbert", "****")
		}
		if w == "fornax"{
			div[i] = strings.ReplaceAll(w, "fornax", "****")
		}
		
	}
	res := strings.Join(div, " ")
	return res
}

func validate_chirpHandle(w http.ResponseWriter, r *http.Request){
	var data parameter;
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&data)
	if err != nil{
		errorAux("Something went wrong", w)
		return
	}
	if len(data.Body) > 140{
		errorAux("Chirp is too long", w)
		return 
	}
	retorno := cleanOutput(data.Body)
	fmt.Println(retorno)
	res := resp{
		Cleaned_body: retorno,
	}
	dat, _ := json.Marshal(res)
	w.WriteHeader(http.StatusOK)
	w.Write(dat)
}

func (apiCfg *apiConfig) createUsers(w http.ResponseWriter, r *http.Request){
	type Userdata struct{
		Email string
	}
	type resData struct{
		Id string `json:"id"`
		Created_at string `json:"created_at"`
		Updated_at string `json:"updated_at"`
		Email string `json:"email"`
	}
	var Udata Userdata
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&Udata)
	if err != nil{}
	data , err := apiCfg.dbQueries.CreateUser(r.Context(), Udata.Email)
	if err != nil{}
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	res := resData {
		Id:         data.ID.String(),
		Created_at: data.CreatedAt.String(),
		Updated_at: data.UpdatedAt.String(),
		Email:      data.Email,
	}
	dat , err := json.Marshal(res)
	if err != nil{}
	w.Write(dat)
}


func main(){
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, _ := sql.Open("postgres", dbURL)
	
	apiCfg := apiConfig{
		dbQueries : database.New(db),
		platform: os.Getenv("PLATFORM"),
	}
	mux := http.NewServeMux();
	server := http.Server{
		Addr: ":8080",
		Handler: mux,
	}
	
	srcImage := http.Dir("./assets")
	serverImage := http.FileServer(srcImage)
	filterPath := http.StripPrefix("/app/assets", apiCfg.middlewareMetricsInc(serverImage))	
	mux.Handle("/app/assets/logo.png", filterPath)
	mux.Handle("/app/", http.StripPrefix( "/app/", apiCfg.middlewareMetricsInc( http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /admin/metrics", apiCfg.metricsHandler)
	mux.HandleFunc("POST /admin/reset", apiCfg.resetHandler)
	mux.HandleFunc("GET /api/healthz", healthz)
	mux.HandleFunc("POST /api/validate_chirp", validate_chirpHandle)
	mux.HandleFunc("POST /api/users", apiCfg.createUsers)
	fmt.Println("Server running on port 8080")
	server.ListenAndServe();
}