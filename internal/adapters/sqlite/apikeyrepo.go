package sqlite

import (
	"context"
	"fmt"
	"strings"

	"github.com/wakaba/wakaba/internal/domain"
	"github.com/wakaba/wakaba/internal/ports"
	"gorm.io/gorm"
)

type APIKeyRepository struct {
	db *gorm.DB
}

func NewAPIKeyRepository(db *gorm.DB) ports.APIKeyRepository {
	return &APIKeyRepository{db: db}
}

func (r *APIKeyRepository) Create(ctx context.Context, key *domain.APIKey) error {
	scopes := make([]string, len(key.Scopes))
	for i, s := range key.Scopes {
		scopes[i] = string(s)
	}
	rec := APIKeyRecord{
		ExternalID: string(key.ID),
		Name:       key.Name,
		KeyHash:    key.KeyHash,
		KeyPrefix:  key.KeyPrefix,
		Scopes:     strings.Join(scopes, ","),
		Status:     key.Status,
		RevokedAt:  key.RevokedAt,
	}
	return r.db.WithContext(ctx).Create(&rec).Error
}

func (r *APIKeyRepository) GetByID(ctx context.Context, id domain.APIKeyID) (*domain.APIKey, error) {
	var rec APIKeyRecord
	if err := r.db.WithContext(ctx).Where("external_id = ?", string(id)).First(&rec).Error; err != nil {
		return nil, fmt.Errorf("api key not found: %w", err)
	}
	return r.toDomain(&rec), nil
}

func (r *APIKeyRepository) GetByKeyHash(ctx context.Context, keyHash string) (*domain.APIKey, error) {
	var rec APIKeyRecord
	if err := r.db.WithContext(ctx).Where("key_hash = ? AND status = ?", keyHash, "active").First(&rec).Error; err != nil {
		return nil, fmt.Errorf("api key not found: %w", err)
	}
	return r.toDomain(&rec), nil
}

func (r *APIKeyRepository) List(ctx context.Context) ([]*domain.APIKey, error) {
	var recs []APIKeyRecord
	if err := r.db.WithContext(ctx).Order("created_at DESC").Find(&recs).Error; err != nil {
		return nil, fmt.Errorf("list api keys: %w", err)
	}
	keys := make([]*domain.APIKey, len(recs))
	for i, rec := range recs {
		keys[i] = r.toDomain(&rec)
	}
	return keys, nil
}

func (r *APIKeyRepository) Revoke(ctx context.Context, id domain.APIKeyID) error {
	result := r.db.WithContext(ctx).Model(&APIKeyRecord{}).
		Where("external_id = ? AND status = ?", string(id), "active").
		Updates(map[string]interface{}{
			"status": "revoked",
		})
	if result.RowsAffected == 0 {
		return fmt.Errorf("api key not found or already revoked")
	}
	return result.Error
}

func (r *APIKeyRepository) toDomain(rec *APIKeyRecord) *domain.APIKey {
	var scopes []domain.APIKeyScope
	if rec.Scopes != "" {
		parts := strings.Split(rec.Scopes, ",")
		scopes = make([]domain.APIKeyScope, len(parts))
		for i, p := range parts {
			scopes[i] = domain.APIKeyScope(strings.TrimSpace(p))
		}
	}
	return &domain.APIKey{
		ID:        domain.APIKeyID(rec.ExternalID),
		Name:      rec.Name,
		KeyHash:   rec.KeyHash,
		KeyPrefix: rec.KeyPrefix,
		Scopes:    scopes,
		Status:    rec.Status,
		CreatedAt: rec.CreatedAt,
		RevokedAt: rec.RevokedAt,
	}
}
