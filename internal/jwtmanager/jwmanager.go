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
	accessPayload := m.createPayload(uuid, guid, ip, accessTokenTimeout)
	refreshPayload := m.createPayload(uuid, guid, ip, refreshTokenTimeout)

	access, err = m.signToken(accessPayload)
	if err != nil {
		return "", "", err
	}

	refresh, err = m.signToken(refreshPayload)
	if err != nil {
		return "", "", err
	}

	return
}

func (m *Manager) createPayload(uuid, guid, ip string, timeout time.Duration) jwt.MapClaims {
	return jwt.MapClaims{
		"uuid":         uuid,
		contextKeyIP:   ip,
		contextKeyGUID: guid,
		contextKeyExp:  time.Now().Add(timeout).Unix(),
	}
}

func (m *Manager) signToken(payload jwt.MapClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, payload)
	return token.SignedString(m.secret)
}

func (m *Manager) ParseJWT(tokenStr string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrInvalidKeyType
		}
		return m.secret, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, jwt.ErrInvalidKeyType
	}

	return claims, nil
}
