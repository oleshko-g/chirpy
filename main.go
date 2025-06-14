package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

const port string = "8080"
const rootPath = "./public"

var c *apiConfig = &apiConfig{
	fileserverHits: atomic.Int32{},
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

	mux.HandleFunc("/healthz", healthzHandler)
	mux.HandleFunc("/metrics", c.showFileSrvHits)
	mux.HandleFunc("/reset", c.resetFileSrvHits)
	return mux
}

func main() {

	server := &http.Server{
		Handler: newServeMux(),
		Addr:    ":" + port,
	}

	fmt.Printf("Serving on the port %s", server.Addr)
	log.Fatal(server.ListenAndServe())
}
