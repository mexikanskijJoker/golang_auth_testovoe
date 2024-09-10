package jwtmanager

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	contextKeyIP        = "ip"
	contextKeyGUID      = "guid"
	contextKeyExp       = "exp"
	accessTokenTimeout  = time.Minute * 30
	refreshTokenTimeout = time.Hour * 24
)

type Manager struct {
	secret []byte
}

func New(secret []byte) *Manager {
	return &Manager{secret: secret}
}

func (m *Manager) GenerateJWT(guid, ip string) (access, refresh string, err error) {
	uuid := uuid.New().String()
	accessPayload := jwt.MapClaims{
		uuid:           uuid,
		contextKeyIP:   ip,
		contextKeyGUID: guid,
		contextKeyExp:  time.Now().Add(accessTokenTimeout).Unix(),
	}
	refreshPayload := jwt.MapClaims{
		uuid:           uuid,
		contextKeyIP:   ip,
		contextKeyGUID: guid,
		contextKeyExp:  time.Now().Add(refreshTokenTimeout).Unix(),
	}

	accessT := jwt.NewWithClaims(jwt.SigningMethodHS512, accessPayload)
	refreshT := jwt.NewWithClaims(jwt.SigningMethodHS512, refreshPayload)

	access, err = accessT.SignedString(m.secret)
	if err != nil {
		return "", "", err
	}

	refresh, err = refreshT.SignedString(m.secret)
	if err != nil {
		return "", "", err
	}

	return
}
