package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/oleshko-g/chirpy/internal/auth"
	"github.com/oleshko-g/chirpy/internal/database"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries      *database.Queries
	platform       string
	jwtSecret      string
}

type durationInSeconds time.Duration

func (d *durationInSeconds) UnmarshalJSON(data []byte) error {
	var input int
	if err := json.Unmarshal(data, &input); err != nil {
		return err
	}

	*d = durationInSeconds(time.Duration(input) * time.Second)

	return nil
}

func (d durationInSeconds) String() string {
	return time.Duration(d).String()
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

func validateChirp(chirpBody string) error {
	const chirpMaxLength int = 140

	if len(chirpBody) > chirpMaxLength {
		return fmt.Errorf("chirp is longer then %d characters", chirpMaxLength)
	}
	return nil
}

func createUser(w http.ResponseWriter, req *http.Request) {

	var reqBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
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

	hashedPassword, errHashPassword := auth.HashPassword(reqBody.Password)
	if errHashPassword != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	createdUser, errCreateUser := c.dbQueries.InsetUser(req.Context(), database.InsetUserParams{
		Email: reqBody.Email,
		HashedPassword: sql.NullString{
			String: hashedPassword,
			Valid:  true,
		},
	})

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

func loginUser(w http.ResponseWriter, r *http.Request) {

	var reqBody struct {
		Email     string             `json:"email"`
		Password  string             `json:"password"`
		ExpiresIn *durationInSeconds `json:"expires_in_seconds"`
	}

	errDecode := json.NewDecoder(r.Body).Decode(&reqBody)
	if errDecode != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(errDecode.Error()))
		return
	}

	fmt.Printf("%+v", reqBody)

	selectedUser, errSelectUserByEmail := c.dbQueries.SelectUserByEmail(r.Context(), reqBody.Email)

	if errSelectUserByEmail != nil {
		if errors.Is(errSelectUserByEmail, sql.ErrNoRows) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if errCheckPasswordHash := auth.CheckPasswordHash(selectedUser.HashedPassword.String, reqBody.Password); errCheckPasswordHash != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	errEncode := json.NewEncoder(w).Encode(
		struct {
			ID        uuid.UUID `json:"id"`
			CreatedAt time.Time `json:"created_at"`
			UpdatedAt time.Time `json:"updated_at"`
			Email     string    `json:"email"`
		}{
			ID:        selectedUser.ID,
			CreatedAt: selectedUser.CreatedAt,
			UpdatedAt: selectedUser.UpdatedAt,
			Email:     selectedUser.Email,
		},
	)
	if errEncode != nil {
		fmt.Fprintf(os.Stderr, "%s", errEncode)
		return
	}
}

func createChirp(w http.ResponseWriter, r *http.Request) {
	var reqBody struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}

	errDecode := json.NewDecoder(r.Body).Decode(&reqBody)
	if errDecode != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	errValidateChirp := validateChirp(reqBody.Body)
	if errValidateChirp != nil {
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		errEncode := json.NewEncoder(w).Encode(struct {
			Error string `json:"error"`
		}{Error: errValidateChirp.Error()})
		if errEncode != nil {
			fmt.Fprintf(os.Stderr, "%s", errEncode)
			return
		}
		return
	}

	createdChirp, errCreateChirp := c.dbQueries.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   cleanInput(reqBody.Body),
		UserID: reqBody.UserID,
	})

	if errCreateChirp != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	errEncode := json.NewEncoder(w).Encode(
		struct {
			ID        string `json:"id"`
			CreatedAt string `json:"created_at"`
			UpdatedAt string `json:"updated_at"`
			Body      string `json:"body"`
			UserID    string `json:"user_id"`
		}{
			ID:        createdChirp.ID.String(),
			CreatedAt: createdChirp.CreatedAt.String(),
			UpdatedAt: createdChirp.UpdatedAt.String(),
			Body:      createdChirp.Body,
			UserID:    createdChirp.UserID.String(),
		},
	)
	if errEncode != nil {
		fmt.Fprintf(os.Stderr, "%s", errEncode)
		return
	}
}

func getChirps(w http.ResponseWriter, r *http.Request) {

	selectedChirps, errSelectChirps := c.dbQueries.SelectChirps(r.Context())
	if errSelectChirps != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	chirps := make([]struct {
		ID        string `json:"id"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
		Body      string `json:"body"`
		UserID    string `json:"user_id"`
	}, len(selectedChirps))
	fmt.Printf("%+v", selectedChirps)
	for i, v := range selectedChirps {
		chirps[i].ID = v.ID.String()
		chirps[i].CreatedAt = v.CreatedAt.Format(time.RFC3339)
		chirps[i].UpdatedAt = v.UpdatedAt.Format(time.RFC3339)
		chirps[i].Body = v.Body
		chirps[i].UserID = v.UserID.String()
	}

	errEncode := json.NewEncoder(w).Encode(chirps)
	if errEncode != nil {
		fmt.Fprintf(os.Stderr, "%s", errEncode)
		return
	}
}

func getChirp(w http.ResponseWriter, r *http.Request) {
	chirp_uuid, errParse := uuid.Parse(r.PathValue("chirp_id"))
	if errParse != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	selectedChirp, errSelectChirp := c.dbQueries.SelectChirp(r.Context(), chirp_uuid)
	if errSelectChirp != nil {
		if errors.Is(errSelectChirp, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	errEncode := json.NewEncoder(w).Encode(struct {
		ID        string `json:"id"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
		Body      string `json:"body"`
		UserID    string `json:"user_id"`
	}{
		ID:        selectedChirp.ID.String(),
		CreatedAt: selectedChirp.CreatedAt.Format(time.RFC3339),
		UpdatedAt: selectedChirp.UpdatedAt.Format(time.RFC3339),
		Body:      selectedChirp.Body,
		UserID:    selectedChirp.UserID.String(),
	})
	if errEncode != nil {
		fmt.Fprintf(os.Stderr, "%s", errEncode)
		return
	}
}
