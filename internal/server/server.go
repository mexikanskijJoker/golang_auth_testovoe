package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mexikanskijjoker/golang_auth_testovoe/internal/jwtmanager"
	"github.com/mexikanskijjoker/golang_auth_testovoe/internal/store"
)

type Server struct {
	*http.Server

	m     *jwtmanager.Manager
	store store.TokenStore
	// log   Logger
}

func New(m *jwtmanager.Manager, store store.TokenStore) *Server {
	s := &Server{
		m:     m,
		store: store,
	}

	r := mux.NewRouter()
	r.HandleFunc("/api/v1/login", s.handleLogin)
	// r.HandleFunc("/api/v1/refresh", s.handleRefresh)

	s.Server = &http.Server{
		Handler: r,
		Addr:    ":8000",
	}

	return s
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	guid := r.URL.Query().Get("guid")
	if guid == "" {
		http.Error(w, "invalid params", http.StatusForbidden)
		return
	}

	ip := r.Header.Get("X-Forwarded-For")
	access, refresh, err := s.m.GenerateJWT(guid, ip)
	if err != nil {
		http.Error(w, fmt.Sprintf("generate jwt: %v", err), http.StatusInternalServerError)
	}

	payload, err := json.Marshal(map[string]string{
		"access":  access,
		"refresh": refresh,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("marshal payload: %v", err), http.StatusInternalServerError)
	}

	w.Write(payload)
}

// func (s *Server) handleRefresh(w http.ResponseWriter, r *http.Request) {

// }
