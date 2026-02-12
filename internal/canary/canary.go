package canary

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

type Service struct {
	secret []byte
}

func New(secret string) *Service {
	return &Service{secret: []byte(secret)}
}

func (s *Service) Token(sessionID, path string, n int) string {
	id := fmt.Sprintf("%s:%s:%d", sessionID, path, n)
	mac := hmac.New(sha256.New, s.secret)
	_, _ = mac.Write([]byte(id))
	sig := hex.EncodeToString(mac.Sum(nil))
	return fmt.Sprintf("swc_%s_%s", short(id), sig[:16])
}

func (s *Service) Verify(token string, sessionID, path string, n int) bool {
	expected := s.Token(sessionID, path, n)
	return hmac.Equal([]byte(expected), []byte(token))
}

func Inject(template, token string) string {
	return strings.ReplaceAll(template, "__TOKEN__", token)
}

func short(v string) string {
	h := sha256.Sum256([]byte(v))
	return hex.EncodeToString(h[:4])
}
