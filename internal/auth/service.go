package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
)

type Service struct {
	keyHash string
}

func New(apiKey string) *Service {
	sum := sha256.Sum256([]byte(apiKey))
	return &Service{keyHash: hex.EncodeToString(sum[:])}
}

func (s *Service) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("X-API-Key")
		if key == "" {
			key = strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		}
		if key == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		sum := sha256.Sum256([]byte(key))
		if hex.EncodeToString(sum[:]) != s.keyHash {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func HashKey(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func KeyPrefix(raw string) string {
	if len(raw) < 8 {
		return raw
	}
	return raw[:8]
}
