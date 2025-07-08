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

func authenticateUserMiddleware(handlerWithUser func(w http.ResponseWriter, r *http.Request, userID uuid.UUID)) (handler func(w http.ResponseWriter, r *http.Request)) {
	return func(w http.ResponseWriter, r *http.Request) {
		b, errGetBearerToken := auth.GetBearerToken(&r.Header)
		if errGetBearerToken != nil {
			w.WriteHeader(http.StatusUnauthorized)
		}
		userID, errValidateUserJWT := auth.ValidateUserJWT(b, c.jwtSecret)
		if errValidateUserJWT != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		handlerWithUser(w, r, userID)
	}
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

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt string    `json:"created_at"`
		UpdatedAt string    `json:"updated_at"`
		Email     string    `json:"email"`
	}{
		ID:        createdUser.ID,
		CreatedAt: createdUser.CreatedAt.Format(time.RFC3339),
		UpdatedAt: createdUser.UpdatedAt.Format(time.RFC3339),
		Email:     createdUser.Email,
	})
}

func loginUser(w http.ResponseWriter, r *http.Request) {

	var reqBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	errDecode := json.NewDecoder(r.Body).Decode(&reqBody)
	if errDecode != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(errDecode.Error()))
		return
	}

	fmt.Printf("Loggin in the user with the credentials:\n%+v", reqBody)
	fmt.Println()

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

	selectedUserJWT, errSignUserJWT := auth.SignUserJWT(
		selectedUser.ID,
		c.jwtSecret,
		time.Duration(1)*time.Hour,
	)
	if errSignUserJWT != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	refreshToken, _ := auth.MakeRefreshToken()
	errInsertRefreshToken := c.dbQueries.InsertRefreshToken(r.Context(),
		database.InsertRefreshTokenParams{
			Token:     refreshToken,
			UserID:    selectedUser.ID,
			ExpiresAt: time.Now().UTC().AddDate(0, 0, 60),
		},
	)
	if errInsertRefreshToken != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	errEncode := json.NewEncoder(w).Encode(
		struct {
			ID           uuid.UUID `json:"id"`
			CreatedAt    string    `json:"created_at"`
			UpdatedAt    string    `json:"updated_at"`
			Email        string    `json:"email"`
			Token        string    `json:"token"`
			RefreshToken string    `json:"refresh_token"`
		}{
			ID:           selectedUser.ID,
			CreatedAt:    selectedUser.CreatedAt.Format(time.RFC3339),
			UpdatedAt:    selectedUser.UpdatedAt.Format(time.RFC3339),
			Email:        selectedUser.Email,
			Token:        selectedUserJWT,
			RefreshToken: refreshToken,
		},
	)
	if errEncode != nil {
		fmt.Fprintf(os.Stderr, "%s", errEncode)
		return
	}
}

func createChirp(w http.ResponseWriter, r *http.Request) {
	bearerToken, errGetBearerToken := auth.GetBearerToken(&r.Header)
	if errGetBearerToken != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userID, errValitateUserJWT := auth.ValidateUserJWT(bearerToken, c.jwtSecret)
	if errValitateUserJWT != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var reqBody struct {
		Body string `json:"body"`
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
		UserID: userID,
	})

	if errCreateChirp != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	errEncode := json.NewEncoder(w).Encode(
		struct {
			ID        uuid.UUID `json:"id"`
			CreatedAt string    `json:"created_at"`
			UpdatedAt string    `json:"updated_at"`
			Body      string    `json:"body"`
			UserID    string    `json:"user_id"`
		}{
			ID:        createdChirp.ID,
			CreatedAt: createdChirp.CreatedAt.Format(time.RFC3339),
			UpdatedAt: createdChirp.UpdatedAt.Format(time.RFC3339),
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
		ID        uuid.UUID `json:"id"`
		CreatedAt string    `json:"created_at"`
		UpdatedAt string    `json:"updated_at"`
		Body      string    `json:"body"`
		UserID    string    `json:"user_id"`
	}, len(selectedChirps))
	fmt.Printf("%+v", selectedChirps)
	for i, v := range selectedChirps {
		chirps[i].ID = v.ID
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

	if selectedChirp.DeletedAt.Valid {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	errEncode := json.NewEncoder(w).Encode(struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt string    `json:"created_at"`
		UpdatedAt string    `json:"updated_at"`
		Body      string    `json:"body"`
		UserID    uuid.UUID `json:"user_id"`
	}{
		ID:        selectedChirp.ID,
		CreatedAt: selectedChirp.CreatedAt.Format(time.RFC3339),
		UpdatedAt: selectedChirp.UpdatedAt.Format(time.RFC3339),
		Body:      selectedChirp.Body,
		UserID:    selectedChirp.UserID,
	})
	if errEncode != nil {
		fmt.Fprintf(os.Stderr, "%s", errEncode)
		return
	}
}

func refreshAccessTokenHandler(w http.ResponseWriter, r *http.Request) {
	b, errGetBearerToken := auth.GetBearerToken(&r.Header)
	if errGetBearerToken != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	selectedRefreshToken, errSelectRefreshToken := c.dbQueries.SelectRefreshToken(r.Context(), b)

	if errors.Is(errSelectRefreshToken, sql.ErrNoRows) ||
		selectedRefreshToken.ExpiresAt.Before(time.Now()) ||
		selectedRefreshToken.RevokedAt.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if errSelectRefreshToken != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	token, errSignUserJWT := auth.SignUserJWT(selectedRefreshToken.UserID, c.jwtSecret, time.Duration(1)*time.Hour)
	if errSignUserJWT != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	errEncode := json.NewEncoder(w).Encode(struct {
		Token string `json:"token"`
	}{
		Token: token,
	})
	if errEncode != nil {
		fmt.Fprintf(os.Stderr, "%s", errEncode)
		return
	}

}

func UpdateRefreshToken(w http.ResponseWriter, r *http.Request) {
	b, errGetBearerToken := auth.GetBearerToken(&r.Header)
	if errGetBearerToken != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	errUpdateRefreshToken := c.dbQueries.UpdateRefreshToken(r.Context(),
		database.UpdateRefreshTokenParams{
			Token:     b,
			RevokedAt: sql.NullTime{Time: time.Now().UTC(), Valid: true},
		},
	)
	if errUpdateRefreshToken != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func putUserHandler(w http.ResponseWriter, r *http.Request, userID uuid.UUID) {
	var reqBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	errDecode := json.NewDecoder(r.Body).Decode(&reqBody)
	if errDecode != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	hashedPassword, errHashPassword := auth.HashPassword(reqBody.Password)
	if errHashPassword != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	errUpdateUser := c.dbQueries.UpdateUser(r.Context(), database.UpdateUserParams{
		ID:    userID,
		Email: reqBody.Email,
		HashedPassword: sql.NullString{
			String: hashedPassword,
			Valid:  true,
		},
	})

	if errUpdateUser != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(struct {
		Email string `json:"email"`
	}{
		Email: reqBody.Email,
	})
}

func deleteChirp(w http.ResponseWriter, r *http.Request, userID uuid.UUID) {
	chirp_uuid, errParse := uuid.Parse(r.PathValue("chirp_id"))
	if errParse != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	selectedChirp, errSelectChirp := c.dbQueries.SelectChirp(r.Context(), chirp_uuid)
	if errSelectChirp != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	fmt.Printf("%+v\n", selectedChirp)

	if selectedChirp.UserID != userID {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	if selectedChirp.DeletedAt.Valid {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	errUpdateChirp := c.dbQueries.UpdateChirp(r.Context(), database.UpdateChirpParams{
		ID:     chirp_uuid,
		UserID: userID,
		DeletedAt: sql.NullTime{
			Time:  time.Now().UTC(),
			Valid: true,
		},
	})
	if errUpdateChirp != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)

}
