package server

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mexikanskijjoker/golang_auth_testovoe/internal/jwtmanager"
	"github.com/mexikanskijjoker/golang_auth_testovoe/internal/store"
	"golang.org/x/crypto/bcrypt"
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
	email := "test@gmail.com"
	if guid == "" {
		http.Error(w, "invalid params", http.StatusForbidden)
		return
	}

	ip := r.Header.Get("X-Forwarded-For")

	if err := s.store.CreateUser(email, ip, guid); err != nil {
		http.Error(w, fmt.Sprintf("create user: %v", err), http.StatusInternalServerError)
		return
	}

	access, refresh, err := s.m.GenerateJWT(guid, ip)
	if err != nil {
		http.Error(w, fmt.Sprintf("generate jwt: %v", err), http.StatusInternalServerError)
		return
	}

	encodedRefreshToken, err := bcrypt.GenerateFromPassword([]byte(refresh), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, fmt.Sprintf("bcrypt.GenerateFromPassword: %v", err), http.StatusInternalServerError)
		return
	}

	encodedRefreshTokenStr := base64.StdEncoding.EncodeToString(encodedRefreshToken)
	if err := s.store.CreateRefreshToken(encodedRefreshTokenStr); err != nil {
		http.Error(w, fmt.Sprintf("create refresh token: %v", err), http.StatusInternalServerError)
		return
	}

	payload, err := json.Marshal(map[string]string{
		"access":  access,
		"refresh": base64.StdEncoding.EncodeToString(encodedRefreshToken),
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("marshal payload: %v", err), http.StatusInternalServerError)
	}

	w.Write(payload)
}

// func (s *Server) handleRefresh(w http.ResponseWriter, r *http.Request) {

// }
