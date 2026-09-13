package domain

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

type APIKeyID string

func GenerateAPIKeyID() APIKeyID {
	return APIKeyID(GenerateMessageID())
}

type APIKeyScope string

const (
	ScopeAdminRead      APIKeyScope = "admin:read"
	ScopeAdminWrite     APIKeyScope = "admin:write"
	ScopeAdminMessaging APIKeyScope = "admin:messaging"
	ScopeAdminObs       APIKeyScope = "admin:obs"
)

var AllScopes = []APIKeyScope{
	ScopeAdminRead,
	ScopeAdminWrite,
	ScopeAdminMessaging,
	ScopeAdminObs,
}

func ParseScopes(s string) []APIKeyScope {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	scopes := make([]APIKeyScope, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			scopes = append(scopes, APIKeyScope(p))
		}
	}
	return scopes
}

func (s APIKeyScope) IsValid() bool {
	switch s {
	case ScopeAdminRead, ScopeAdminWrite, ScopeAdminMessaging, ScopeAdminObs:
		return true
	}
	return false
}

type APIKey struct {
	ID        APIKeyID      `json:"id"`
	Name      string        `json:"name"`
	KeyHash   string        `json:"-"`
	KeyPrefix string        `json:"key_prefix"`
	Scopes    []APIKeyScope `json:"scopes"`
	Status    string        `json:"status"`
	CreatedAt time.Time     `json:"created_at"`
	RevokedAt *time.Time    `json:"revoked_at,omitempty"`
}

func GenerateAPIKey() (plain string, keyPrefix string) {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	plain = hex.EncodeToString(b)
	keyPrefix = plain[:8]
	return
}

func NewAPIKey(name string, scopes []APIKeyScope) (*APIKey, string) {
	plain, prefix := GenerateAPIKey()
	ak := &APIKey{
		ID:        GenerateAPIKeyID(),
		Name:      name,
		KeyHash:   HashToken(plain),
		KeyPrefix: prefix,
		Scopes:    scopes,
		Status:    "active",
		CreatedAt: time.Now(),
	}
	return ak, plain
}

func (a *APIKey) HasScope(scope APIKeyScope) bool {
	for _, s := range a.Scopes {
		if s == scope {
			return true
		}
	}
	return false
}

func (a *APIKey) HasAnyScope(scopes ...APIKeyScope) bool {
	for _, required := range scopes {
		if a.HasScope(required) {
			return true
		}
	}
	return false
}

func (a *APIKey) CanRead() bool     { return a.HasScope(ScopeAdminRead) }
func (a *APIKey) CanWrite() bool    { return a.HasScope(ScopeAdminWrite) }
func (a *APIKey) CanMessage() bool  { return a.HasScope(ScopeAdminMessaging) }
func (a *APIKey) CanObserve() bool  { return a.HasScope(ScopeAdminObs) }

type AuthorizedAPIKeyResult struct {
	APIKey *APIKey
}

func (r *AuthorizedAPIKeyResult) String() string {
	return fmt.Sprintf("api_key=%s name=%s scopes=%v", r.APIKey.ID, r.APIKey.Name, r.APIKey.Scopes)
}
