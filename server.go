package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (c *apiConfig) incFileSrvHits(h http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		c.fileserverHits.Add(1)
		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(handler)
}

func (c *apiConfig) resetFileSrvHits(w http.ResponseWriter, _ *http.Request) {
	c.fileserverHits.Store(0)
	w.Header().Add("content-type", "text/plain; charset=utf-8")
	fileServerHits := strconv.Itoa(int(c.fileserverHits.Load()))
	responseData := []byte("Hits have been reset to: " + fileServerHits)
	w.Write(responseData)
}

func (c *apiConfig) showFileSrvHits(w http.ResponseWriter, _ *http.Request) {
	w.Header().Add("content-type", "text/html")
	responseData := []byte(
		fmt.Sprintf(
			`<html>
				<body>
					<h1>Welcome, Chirpy Admin</h1>
					<p>Chirpy has been visited %d times!</p>
				</body>
			</html>`, int(c.fileserverHits.Load()),
		),
	)

	w.Write(responseData)
}

func healthzHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Add("content-type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func validateChirp(w http.ResponseWriter, req *http.Request) {
	const chirpMaxLength int = 140

	type ValidateChirpRequest struct {
		Body string `json:"body"`
	}

	type ValidateChirpResponse struct {
		Valid bool `json:"valid"`
	}

	type ValidateChirpErr struct {
		Error string `json:"error"`
	}

	decoder := json.NewDecoder(req.Body)
	defer req.Body.Close()
	var reqBody ValidateChirpRequest
	err := decoder.Decode(&reqBody)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		errEncode := json.NewEncoder(w).Encode(ValidateChirpErr{
			Error: err.Error(),
		})
		if errEncode != nil {
			fmt.Fprintf(os.Stderr, "%s", errEncode)
			return
		}
		return
	}
	log.Printf("%+v", reqBody)

	if len(reqBody.Body) > chirpMaxLength {
		w.WriteHeader(http.StatusBadRequest)
		errEncode := json.NewEncoder(w).Encode(ValidateChirpErr{
			Error: fmt.Sprintf("Chirp is longer then %d characters", chirpMaxLength),
		})
		if errEncode != nil {
			fmt.Fprintf(os.Stderr, "%s", errEncode)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		return
	}

	w.Header().Set("content-type", "application/json")
	errEncode := json.NewEncoder(w).Encode(ValidateChirpResponse{Valid: true})
	if errEncode != nil {
		fmt.Fprintf(os.Stderr, "%s", errEncode)
		return
	}
}
