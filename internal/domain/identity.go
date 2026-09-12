package domain

import "time"

type VirtualIdentityID string

type VirtualIdentity struct {
	ID          VirtualIdentityID `json:"id"`
	Phone       PhoneNumber       `json:"phone"`
	DisplayName string            `json:"display_name"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

func NewVirtualIdentity(phone PhoneNumber, displayName string) *VirtualIdentity {
	now := time.Now()
	return &VirtualIdentity{
		ID:          VirtualIdentityID(GenerateMessageID()),
		Phone:       phone,
		DisplayName: displayName,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}
