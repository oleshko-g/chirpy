package main

import (
	"fmt"
	"log"
	"net/http"
)

const port string = "8080"
const rootPath = "./public"

func newServeMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("/app/",
		http.StripPrefix("/app",
			http.FileServer(http.Dir(rootPath))))

	mux.HandleFunc("/healthz", healthzHandler)
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
