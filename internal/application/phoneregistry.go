package application

import (
	"context"
	"fmt"

	"github.com/wabasi/wabasi/internal/domain"
	"github.com/wabasi/wabasi/internal/ports"
)

type PhoneRegistry struct {
	phones ports.PhoneNumberRepository
}

func NewPhoneRegistry(phones ports.PhoneNumberRepository) *PhoneRegistry {
	return &PhoneRegistry{phones: phones}
}

func (r *PhoneRegistry) Reserve(ctx context.Context, phone domain.PhoneNumber, ownerType string, ownerID string) error {
	available, err := r.phones.IsAvailable(ctx, phone)
	if err != nil {
		return fmt.Errorf("check availability: %w", err)
	}
	if !available {
		return fmt.Errorf("phone number %s is already registered", phone)
	}
	return r.phones.Register(ctx, phone, ownerType, ownerID)
}

func (r *PhoneRegistry) IsAvailable(ctx context.Context, phone domain.PhoneNumber) (bool, error) {
	return r.phones.IsAvailable(ctx, phone)
}

func (r *PhoneRegistry) GetOwner(ctx context.Context, phone domain.PhoneNumber) (string, string, error) {
	return r.phones.GetOwner(ctx, phone)
}
