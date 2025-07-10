package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/oleshko-g/chirpy/internal/database"
)

const port string = "8080"
const rootPath = "./public"

var c *apiConfig = &apiConfig{
	fileserverHits: atomic.Int32{},
}

func init() {
	godotenv.Load()

	c.platform = os.Getenv("PLATFORM")
	c.jwtSecret = os.Getenv("JWT_SECRET")

	dbConn, errDBConn := openPostgresDB(os.Getenv("DB_URL"))
	if errDBConn != nil {
		fmt.Fprintln(os.Stderr, errDBConn)
		os.Exit(1)
	}

	c.dbQueries = database.New(dbConn)
}

func newServeMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("/app/",
		c.incFileSrvHits(
			http.StripPrefix("/app",
				http.FileServer(
					http.Dir(rootPath),
				),
			),
		),
	)

	mux.HandleFunc("GET /api/healthz", healthzHandler)
	mux.HandleFunc("POST /api/users", createUser)
	mux.HandleFunc("PUT /api/users", authenticateUserMiddleware(putUserHandler))
	mux.HandleFunc("POST /api/login", loginUser)
	mux.HandleFunc("GET /admin/metrics", c.showFileSrvHits)
	mux.HandleFunc("POST /admin/reset", c.resetServer)
	mux.HandleFunc("POST /api/chirps", createChirp)
	mux.HandleFunc("DELETE /api/chirps/{chirp_id}", authenticateUserMiddleware(deleteChirp))
	mux.HandleFunc("GET /api/chirps", getChirps)
	mux.HandleFunc("GET /api/chirps/{chirp_id}", getChirp)
	mux.HandleFunc("POST /api/refresh", refreshAccessTokenHandler)
	mux.HandleFunc("POST /api/revoke", UpdateRefreshToken)
	mux.HandleFunc("POST /api/polka/webhooks", setUserIsChirpyRed)

	return mux
}

func openPostgresDB(dbURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return db, nil
}

func main() {

	server := &http.Server{
		Handler: newServeMux(),
		Addr:    ":" + port,
	}

	fmt.Printf("Serving on the port %s\n", server.Addr)
	log.Fatal(server.ListenAndServe())
}
