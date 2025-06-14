package main

import (
	"net/http"
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
	w.Header().Add("content-type", "text/plain; charset=utf-8")
	fileServerHits := strconv.Itoa(int(c.fileserverHits.Load()))
	responseData := []byte("Hits: " + fileServerHits)
	w.Write(responseData)
}

func healthzHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Add("content-type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}
