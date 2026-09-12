package domain

import "time"

type BusinessID string
type WABAID string
type BusinessPhoneNumberID string
type ApplicationID string
type CredentialID string

type Business struct {
	ID          BusinessID `json:"id"`
	Name        string     `json:"name"`
	DisplayName string     `json:"display_name,omitempty"`
	Country     string     `json:"country,omitempty"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type WABA struct {
	ID        BusinessID `json:"id"`
	BusinessID BusinessID `json:"business_id"`
	Name      string     `json:"name"`
	Timezone  string     `json:"timezone,omitempty"`
	Currency  string     `json:"currency,omitempty"`
	Status    string     `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
}

type BusinessPhoneNumber struct {
	ID            BusinessPhoneNumberID `json:"id"`
	WABAID        WABAID                `json:"waba_id"`
	Phone         PhoneNumber           `json:"phone"`
	DisplayName   string                `json:"display_name,omitempty"`
	Country       string                `json:"country,omitempty"`
	Status        string                `json:"status"`
	CreatedAt     time.Time             `json:"created_at"`
}

type Application struct {
	ID          ApplicationID `json:"id"`
	BusinessID  BusinessID    `json:"business_id"`
	Name        string        `json:"name"`
	Status      string        `json:"status"`
	AppSecret   string        `json:"app_secret,omitempty"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

type Credential struct {
	ID            CredentialID  `json:"id"`
	ApplicationID ApplicationID `json:"application_id"`
	TokenHash     string        `json:"token_hash,omitempty"`
	Status        string        `json:"status"`
	CreatedAt     time.Time     `json:"created_at"`
	RevokedAt     *time.Time    `json:"revoked_at,omitempty"`
}

type WebhookConfig struct {
	ID            ApplicationID `json:"id"`
	ApplicationID ApplicationID `json:"application_id"`
	CallbackURL   string        `json:"callback_url"`
	VerifyToken   string        `json:"verify_token"`
	Enabled       bool          `json:"enabled"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}
