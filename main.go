package main

import "net/http"

func main() {
	server := http.Server{
		Handler: http.NewServeMux(),
		Addr:    ":8080",
	}
	defer server.Close()

	for {
		server.ListenAndServe()
	}
}
