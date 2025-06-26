package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/oleshko-g/chirpy/internal/database"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries      *database.Queries
	platform       string
}

func (c *apiConfig) incFileSrvHits(h http.Handler) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		c.fileserverHits.Add(1)
		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(handler)
}

func (c *apiConfig) resetServer(w http.ResponseWriter, req *http.Request) {
	if c.platform != "dev" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	c.fileserverHits.Store(0)
	w.Header().Add("content-type", "text/plain; charset=utf-8")
	fileServerHits := strconv.Itoa(int(c.fileserverHits.Load()))

	errResetUsers := c.dbQueries.ResetUsers(req.Context())
	if errResetUsers != nil {
		fmt.Fprintf(os.Stderr, "%s", errResetUsers)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Server Hits have been reset to: " + fileServerHits + ".\n" + "All data has been reset."))
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

func cleanInput(s string) string {
	const profaneStub string = "****"
	profanities := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}

	fields := strings.Fields(s)
	for i, field := range fields {
		field = strings.ToLower(field)
		if _, ok := profanities[field]; ok {
			fields[i] = profaneStub
		}
	}

	return strings.Join(fields, " ")
}

func validateChirp(w http.ResponseWriter, req *http.Request) {
	const chirpMaxLength int = 140

	type ValidateChirpRequest struct {
		Body string `json:"body"`
	}

	type ValidateChirpResponse struct {
		CleanedBody string `json:"cleaned_body"`
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
	errEncode := json.NewEncoder(w).Encode(ValidateChirpResponse{CleanedBody: cleanInput(reqBody.Body)})
	if errEncode != nil {
		fmt.Fprintf(os.Stderr, "%s", errEncode)
		return
	}
}

func createUser(w http.ResponseWriter, req *http.Request) {

	var reqBody struct {
		Email string `json:"email"`
	}

	var resBody struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}

	decoder := json.NewDecoder(req.Body)
	errDecode := decoder.Decode(&reqBody)
	if errDecode != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	createdUser, errCreateUser := c.dbQueries.CreateUser(req.Context(), reqBody.Email)
	if errCreateUser != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resBody.ID = createdUser.ID
	resBody.CreatedAt = createdUser.CreatedAt
	resBody.UpdatedAt = createdUser.UpdatedAt
	resBody.Email = createdUser.Email

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resBody)
}
