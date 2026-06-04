package main

import (
	"fmt"
	"net/http"
	"sync/atomic"

	_ "golang.org/x/tools/go/cfg"
)



type apiConfig struct {
	fileserverHits atomic.Int32
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
	fmt.Fprintf(w, html)
}

func (cfg *apiConfig) resetHandler(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	cfg.fileserverHits.Store(0)
}



func main(){
	mux := http.NewServeMux();
	server := http.Server{
		Addr: ":8080",
		Handler: mux,
	}
	apiCfg := apiConfig{}
	//rota index.html
	mux.Handle("/app/", http.StripPrefix( "/app/", apiCfg.middlewareMetricsInc( http.FileServer(http.Dir(".")))))

	//servir imagens
	// 1 - Definir diretorio do arquivo
	srcImage := http.Dir("./assets")
	// 2 - Criar o servidor de arquivo
	serverImage := http.FileServer(srcImage)

	filterPath := http.StripPrefix("/app/assets", apiCfg.middlewareMetricsInc(serverImage))
	mux.Handle("/app/assets/logo.png", filterPath)


	mux.HandleFunc("GET /admin/metrics", apiCfg.metricsHandler)
	mux.HandleFunc("POST /admin/reset", apiCfg.resetHandler)
	mux.HandleFunc("GET /api/healthz", healthz)
	fmt.Println("Server running on port 8080")
	server.ListenAndServe();
}