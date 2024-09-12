package server

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/smtp"

	"github.com/gorilla/mux"
	"github.com/mexikanskijjoker/golang_auth_testovoe/internal/jwtmanager"
	"github.com/mexikanskijjoker/golang_auth_testovoe/internal/store"
	"golang.org/x/crypto/bcrypt"
)

type Server struct {
	*http.Server
	m     *jwtmanager.Manager
	store store.TokenStore
}

func New(m *jwtmanager.Manager, store store.TokenStore) *Server {
	s := &Server{
		m:     m,
		store: store,
	}

	r := mux.NewRouter()
	r.HandleFunc("/api/v1/login", s.handleLogin).Methods("GET")
	r.HandleFunc("/api/v1/refresh", s.handleRefresh).Methods("POST")

	s.Server = &http.Server{
		Handler: r,
		Addr:    ":8080",
	}

	return s
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	guid := r.URL.Query().Get("guid")
	email := "test@gmail.com"

	ip, err := s.getIP(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("get IP: %v", err), http.StatusInternalServerError)
		return
	}

	if err := s.store.CreateUser(email, ip, guid); err != nil {
		http.Error(w, fmt.Sprintf("create user: %v", err), http.StatusInternalServerError)
		return
	}

	access, refresh, err := s.m.GenerateJWT(guid, ip)
	if err != nil {
		http.Error(w, fmt.Sprintf("generate jwt: %v", err), http.StatusInternalServerError)
		return
	}

	encodedRefreshTokenStr, err := s.encodeRefreshToken(refresh)
	if err != nil {
		http.Error(w, fmt.Sprintf("encode refresh token: %v", err), http.StatusInternalServerError)
		return
	}

	if err := s.store.CreateRefreshToken(encodedRefreshTokenStr); err != nil {
		http.Error(w, fmt.Sprintf("create refresh token: %v", err), http.StatusInternalServerError)
		return
	}

	payload, err := json.Marshal(map[string]string{
		"access":  access,
		"refresh": encodedRefreshTokenStr,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("marshal payload: %v", err), http.StatusInternalServerError)
		return
	}

	w.Write(payload)
}

func (s *Server) handleRefresh(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Access  string
		Refresh string
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("decode request: %v", err), http.StatusBadRequest)
		return
	}

	claims, err := s.m.ParseJWT(req.Access)
	if err != nil {
		http.Error(w, fmt.Sprintf("parse access token: %v", err), http.StatusUnauthorized)
		return
	}

	ip, err := s.getIP(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("get IP: %v", err), http.StatusInternalServerError)
		return
	}

	if claims["ip"] != ip {
		s.sendWarningEmail("test@gmail.com", ip)
	}

	if err := s.verifyRefreshToken(req.Refresh); err != nil {
		http.Error(w, fmt.Sprintf("invalid refresh token: %v", err), http.StatusUnauthorized)
		return
	}

	guid := claims["guid"].(string)
	access, refresh, err := s.m.GenerateJWT(guid, ip)
	if err != nil {
		http.Error(w, fmt.Sprintf("generate new jwt: %v", err), http.StatusInternalServerError)
		return
	}

	encodedRefreshTokenStr, err := s.encodeRefreshToken(refresh)
	if err != nil {
		http.Error(w, fmt.Sprintf("encode refresh token: %v", err), http.StatusInternalServerError)
		return
	}

	if err := s.store.CreateRefreshToken(encodedRefreshTokenStr); err != nil {
		http.Error(w, fmt.Sprintf("create refresh token: %v", err), http.StatusInternalServerError)
		return
	}

	payload, err := json.Marshal(map[string]string{
		"access":  access,
		"refresh": encodedRefreshTokenStr,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("marshal payload: %v", err), http.StatusInternalServerError)
		return
	}

	w.Write(payload)
}

func (s *Server) getIP(r *http.Request) (string, error) {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "", err
	}
	return ip, nil
}

func (s *Server) encodeRefreshToken(refresh string) (string, error) {
	encodedRefreshToken, err := bcrypt.GenerateFromPassword([]byte(refresh[:72]), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(encodedRefreshToken), nil
}

func (s *Server) verifyRefreshToken(refresh string) error {
	decodedRefreshToken, err := base64.StdEncoding.DecodeString(refresh)
	if err != nil {
		return err
	}
	return bcrypt.CompareHashAndPassword(decodedRefreshToken, []byte(refresh[:72]))
}

func (s *Server) sendWarningEmail(email, newIP string) {
	from := "your-email@example.com"
	password := "your-email-password"
	to := email
	smtpHost := "smtp.example.com"
	smtpPort := "587"

	message := []byte(fmt.Sprintf("Subject: IP Address Change Warning\n\nYour IP address has changed to %s.", newIP))

	auth := smtp.PlainAuth("", from, password, smtpHost)
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{to}, message)
	if err != nil {
		fmt.Printf("send warning email: %v\n", err)
	}
}
