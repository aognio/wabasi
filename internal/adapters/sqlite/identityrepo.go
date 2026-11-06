package sqlite

import (
	"context"
	"fmt"

	"github.com/wabasi/wabasi/internal/domain"
	"gorm.io/gorm"
)

type VirtualIdentityRepository struct {
	db *gorm.DB
}

func NewVirtualIdentityRepository(db *gorm.DB) *VirtualIdentityRepository {
	return &VirtualIdentityRepository{db: db}
}

func (r *VirtualIdentityRepository) CreateIdentity(ctx context.Context, identity *domain.VirtualIdentity) error {
	record := VirtualIdentityRecord{
		ExternalID:  string(identity.ID),
		Phone:       string(identity.Phone),
		DisplayName: identity.DisplayName,
	}

	if err := r.db.WithContext(ctx).Create(&record).Error; err != nil {
		return fmt.Errorf("create identity: %w", err)
	}
	return nil
}

func (r *VirtualIdentityRepository) GetIdentityByID(ctx context.Context, id domain.VirtualIdentityID) (*domain.VirtualIdentity, error) {
	var record VirtualIdentityRecord
	if err := r.db.WithContext(ctx).Where("external_id = ?", string(id)).First(&record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("get identity: %w", err)
	}
	return r.toDomain(&record), nil
}

func (r *VirtualIdentityRepository) GetIdentityByPhone(ctx context.Context, phone domain.PhoneNumber) (*domain.VirtualIdentity, error) {
	var record VirtualIdentityRecord
	if err := r.db.WithContext(ctx).Where("phone = ?", string(phone)).First(&record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("get identity by phone: %w", err)
	}
	return r.toDomain(&record), nil
}

func (r *VirtualIdentityRepository) ListIdentities(ctx context.Context) ([]*domain.VirtualIdentity, error) {
	var records []VirtualIdentityRecord
	if err := r.db.WithContext(ctx).Order("created_at ASC").Find(&records).Error; err != nil {
		return nil, fmt.Errorf("list identities: %w", err)
	}

	result := make([]*domain.VirtualIdentity, len(records))
	for i, rec := range records {
		result[i] = r.toDomain(&rec)
	}
	return result, nil
}

func (r *VirtualIdentityRepository) toDomain(rec *VirtualIdentityRecord) *domain.VirtualIdentity {
	return &domain.VirtualIdentity{
		ID:          domain.VirtualIdentityID(rec.ExternalID),
		Phone:       domain.PhoneNumber(rec.Phone),
		DisplayName: rec.DisplayName,
		CreatedAt:   rec.CreatedAt,
		UpdatedAt:   rec.UpdatedAt,
	}
}
