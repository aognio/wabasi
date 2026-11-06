package sqlite

import (
	"context"
	"fmt"

	"github.com/wabasi/wabasi/internal/domain"
	"gorm.io/gorm"
)

type PhoneNumberRepository struct {
	db *gorm.DB
}

func NewPhoneNumberRepository(db *gorm.DB) *PhoneNumberRepository {
	return &PhoneNumberRepository{db: db}
}

func (r *PhoneNumberRepository) Register(ctx context.Context, phone domain.PhoneNumber, ownerType string, ownerID string) error {
	record := PhoneNumberRecord{
		PhoneNumber: string(phone),
		OwnerType:   ownerType,
		OwnerID:     ownerID,
	}

	if err := r.db.WithContext(ctx).Create(&record).Error; err != nil {
		return fmt.Errorf("register phone: %w", err)
	}
	return nil
}

func (r *PhoneNumberRepository) IsAvailable(ctx context.Context, phone domain.PhoneNumber) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&PhoneNumberRecord{}).Where("phone_number = ?", string(phone)).Count(&count).Error; err != nil {
		return false, fmt.Errorf("check availability: %w", err)
	}
	return count == 0, nil
}

func (r *PhoneNumberRepository) GetOwner(ctx context.Context, phone domain.PhoneNumber) (string, string, error) {
	var record PhoneNumberRecord
	if err := r.db.WithContext(ctx).Where("phone_number = ?", string(phone)).First(&record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", "", nil
		}
		return "", "", fmt.Errorf("get owner: %w", err)
	}
	return record.OwnerType, record.OwnerID, nil
}
