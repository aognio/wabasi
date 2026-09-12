package domain

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

type BusinessID string
type WABAID string
type BusinessPhoneNumberID string
type ApplicationID string
type CredentialID string

func GenerateBusinessID() BusinessID {
	return BusinessID(GenerateMessageID())
}

func GenerateWABAID() WABAID {
	return WABAID(GenerateMessageID())
}

func GenerateAppID() ApplicationID {
	return ApplicationID(GenerateMessageID())
}

func GenerateCredentialID() CredentialID {
	return CredentialID(GenerateMessageID())
}

type Business struct {
	ID          BusinessID `json:"id"`
	Name        string     `json:"name"`
	DisplayName string     `json:"display_name,omitempty"`
	Country     string     `json:"country,omitempty"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func NewBusiness(name, displayName, country string) *Business {
	now := time.Now()
	return &Business{
		ID:          GenerateBusinessID(),
		Name:        name,
		DisplayName: displayName,
		Country:     country,
		Status:      "active",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

type WABA struct {
	ID         WABAID     `json:"id"`
	BusinessID BusinessID `json:"business_id"`
	Name       string     `json:"name"`
	Timezone   string     `json:"timezone,omitempty"`
	Currency   string     `json:"currency,omitempty"`
	Status     string     `json:"status"`
	CreatedAt  time.Time  `json:"created_at"`
}

func NewWABA(businessID BusinessID, name, timezone, currency string) *WABA {
	return &WABA{
		ID:         GenerateWABAID(),
		BusinessID: businessID,
		Name:       name,
		Timezone:   timezone,
		Currency:   currency,
		Status:     "active",
		CreatedAt:  time.Now(),
	}
}

type BusinessPhoneNumber struct {
	ID          BusinessPhoneNumberID `json:"id"`
	WABAID      WABAID                `json:"waba_id"`
	Phone       PhoneNumber           `json:"phone"`
	DisplayName string                `json:"display_name,omitempty"`
	Country     string                `json:"country,omitempty"`
	Status      string                `json:"status"`
	CreatedAt   time.Time             `json:"created_at"`
}

func NewBusinessPhoneNumber(wabaID WABAID, phone PhoneNumber, displayName, country string) *BusinessPhoneNumber {
	return &BusinessPhoneNumber{
		ID:          BusinessPhoneNumberID(GenerateMessageID()),
		WABAID:      wabaID,
		Phone:       phone,
		DisplayName: displayName,
		Country:     country,
		Status:      "verified",
		CreatedAt:   time.Now(),
	}
}

type Application struct {
	ID         ApplicationID `json:"id"`
	BusinessID BusinessID    `json:"business_id"`
	Name       string        `json:"name"`
	Status     string        `json:"status"`
	AppSecret  string        `json:"app_secret,omitempty"`
	CreatedAt  time.Time     `json:"created_at"`
	UpdatedAt  time.Time     `json:"updated_at"`
}

func GenerateSecret() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func HashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

func NewApplication(businessID BusinessID, name string) *Application {
	now := time.Now()
	return &Application{
		ID:         GenerateAppID(),
		BusinessID: businessID,
		Name:       name,
		Status:     "active",
		AppSecret:  GenerateSecret(),
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

type Credential struct {
	ID            CredentialID  `json:"id"`
	ApplicationID ApplicationID `json:"application_id"`
	TokenHash     string        `json:"token_hash,omitempty"`
	Status        string        `json:"status"`
	CreatedAt     time.Time     `json:"created_at"`
	RevokedAt     *time.Time    `json:"revoked_at,omitempty"`
}

func NewCredential(applicationID ApplicationID, tokenHash string) *Credential {
	return &Credential{
		ID:            GenerateCredentialID(),
		ApplicationID: applicationID,
		TokenHash:     tokenHash,
		Status:        "active",
		CreatedAt:     time.Now(),
	}
}

type WebhookConfig struct {
	ID            ApplicationID `json:"id"`
	ApplicationID ApplicationID `json:"application_id"`
	CallbackURL   string        `json:"callback_url"`
	VerifyToken   string        `json:"verify_token"`
	AppSecret     string        `json:"app_secret,omitempty"`
	Enabled       bool          `json:"enabled"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

func NewWebhookConfig(appID ApplicationID, callbackURL, verifyToken, appSecret string) *WebhookConfig {
	now := time.Now()
	return &WebhookConfig{
		ID:            appID,
		ApplicationID: appID,
		CallbackURL:   callbackURL,
		VerifyToken:   verifyToken,
		AppSecret:     appSecret,
		Enabled:       true,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

type AuthorizedPhoneResult struct {
	Application   *Application
	WABA          *WABA
	PhoneNumber   *BusinessPhoneNumber
	WebhookConfig *WebhookConfig
}

func (a *AuthorizedPhoneResult) String() string {
	return fmt.Sprintf("app=%s waba=%s phone=%s", a.Application.ID, a.WABA.ID, a.PhoneNumber.ID)
}
