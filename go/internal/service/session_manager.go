package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"

	"gpt-team-api/internal/apperr"
)

type SessionManager struct {
	secret []byte
	ttl    time.Duration
}

type sessionPayload struct {
	UserID    uint64 `json:"uid"`
	ExpiresAt int64  `json:"exp"`
}

func NewSessionManager(secret string, ttl time.Duration) *SessionManager {
	return &SessionManager{
		secret: []byte(secret),
		ttl:    ttl,
	}
}

func (m *SessionManager) Issue(userID uint64) (string, error) {
	payloadBytes, err := json.Marshal(sessionPayload{
		UserID:    userID,
		ExpiresAt: time.Now().UTC().Add(m.ttl).Unix(),
	})
	if err != nil {
		return "", apperr.Internal("session_encode_failed", "failed to create session", err)
	}

	encodedPayload := base64.RawURLEncoding.EncodeToString(payloadBytes)
	signature := m.sign(encodedPayload)
	return encodedPayload + "." + signature, nil
}

func (m *SessionManager) Verify(token string) (uint64, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return 0, apperr.Unauthorized("invalid_session", "login required")
	}

	if subtle.ConstantTimeCompare([]byte(parts[1]), []byte(m.sign(parts[0]))) != 1 {
		return 0, apperr.Unauthorized("invalid_session", "login required")
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return 0, apperr.Unauthorized("invalid_session", "login required")
	}

	var payload sessionPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return 0, apperr.Unauthorized("invalid_session", "login required")
	}

	if payload.UserID == 0 || payload.ExpiresAt <= time.Now().UTC().Unix() {
		return 0, apperr.Unauthorized("session_expired", "login required")
	}

	return payload.UserID, nil
}

func (m *SessionManager) sign(payload string) string {
	mac := hmac.New(sha256.New, m.secret)
	_, _ = mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
