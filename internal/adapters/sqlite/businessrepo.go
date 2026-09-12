package sqlite

import (
	"context"
	"fmt"

	"github.com/wakaba/wakaba/internal/domain"
	"gorm.io/gorm"
)

type BusinessRepository struct {
	db *gorm.DB
}

func NewBusinessRepository(db *gorm.DB) *BusinessRepository {
	return &BusinessRepository{db: db}
}

func (r *BusinessRepository) Create(ctx context.Context, b *domain.Business) error {
	rec := BusinessRecord{
		ExternalID:  string(b.ID),
		Name:        b.Name,
		DisplayName: b.DisplayName,
		Country:     b.Country,
		Status:      b.Status,
	}
	return r.db.WithContext(ctx).Create(&rec).Error
}

func (r *BusinessRepository) GetByID(ctx context.Context, id domain.BusinessID) (*domain.Business, error) {
	var rec BusinessRecord
	if err := r.db.WithContext(ctx).Where("external_id = ?", string(id)).First(&rec).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("get business: %w", err)
	}
	return r.toDomain(&rec), nil
}

func (r *BusinessRepository) List(ctx context.Context) ([]*domain.Business, error) {
	var recs []BusinessRecord
	if err := r.db.WithContext(ctx).Order("created_at ASC").Find(&recs).Error; err != nil {
		return nil, fmt.Errorf("list businesses: %w", err)
	}
	result := make([]*domain.Business, len(recs))
	for i, rec := range recs {
		result[i] = r.toDomain(&rec)
	}
	return result, nil
}

func (r *BusinessRepository) toDomain(rec *BusinessRecord) *domain.Business {
	return &domain.Business{
		ID:          domain.BusinessID(rec.ExternalID),
		Name:        rec.Name,
		DisplayName: rec.DisplayName,
		Country:     rec.Country,
		Status:      rec.Status,
		CreatedAt:   rec.CreatedAt,
		UpdatedAt:   rec.UpdatedAt,
	}
}

type WABARepository struct {
	db *gorm.DB
}

func NewWABARepository(db *gorm.DB) *WABARepository {
	return &WABARepository{db: db}
}

func (r *WABARepository) Create(ctx context.Context, w *domain.WABA) error {
	rec := WABARecord{
		ExternalID: string(w.ID),
		BusinessID: string(w.BusinessID),
		Name:       w.Name,
		Timezone:   w.Timezone,
		Currency:   w.Currency,
		Status:     w.Status,
	}
	return r.db.WithContext(ctx).Create(&rec).Error
}

func (r *WABARepository) GetByID(ctx context.Context, id domain.WABAID) (*domain.WABA, error) {
	var rec WABARecord
	if err := r.db.WithContext(ctx).Where("external_id = ?", string(id)).First(&rec).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("get waba: %w", err)
	}
	return r.toDomain(&rec), nil
}

func (r *WABARepository) GetByBusiness(ctx context.Context, businessID domain.BusinessID) ([]*domain.WABA, error) {
	var recs []WABARecord
	if err := r.db.WithContext(ctx).Where("business_id = ?", string(businessID)).Order("created_at ASC").Find(&recs).Error; err != nil {
		return nil, fmt.Errorf("get wabas by business: %w", err)
	}
	result := make([]*domain.WABA, len(recs))
	for i, rec := range recs {
		result[i] = r.toDomain(&rec)
	}
	return result, nil
}

func (r *WABARepository) List(ctx context.Context) ([]*domain.WABA, error) {
	var recs []WABARecord
	if err := r.db.WithContext(ctx).Order("created_at ASC").Find(&recs).Error; err != nil {
		return nil, fmt.Errorf("list wabas: %w", err)
	}
	result := make([]*domain.WABA, len(recs))
	for i, rec := range recs {
		result[i] = r.toDomain(&rec)
	}
	return result, nil
}

func (r *WABARepository) toDomain(rec *WABARecord) *domain.WABA {
	return &domain.WABA{
		ID:         domain.WABAID(rec.ExternalID),
		BusinessID: domain.BusinessID(rec.BusinessID),
		Name:       rec.Name,
		Timezone:   rec.Timezone,
		Currency:   rec.Currency,
		Status:     rec.Status,
		CreatedAt:  rec.CreatedAt,
	}
}

type BusinessPhoneNumberRepository struct {
	db *gorm.DB
}

func NewBusinessPhoneNumberRepository(db *gorm.DB) *BusinessPhoneNumberRepository {
	return &BusinessPhoneNumberRepository{db: db}
}

func (r *BusinessPhoneNumberRepository) Create(ctx context.Context, p *domain.BusinessPhoneNumber) error {
	rec := BusinessPhoneNumberRecord{
		ExternalID:  string(p.ID),
		WABAID:      string(p.WABAID),
		Phone:       string(p.Phone),
		DisplayName: p.DisplayName,
		Country:     p.Country,
		Status:      p.Status,
	}
	return r.db.WithContext(ctx).Create(&rec).Error
}

func (r *BusinessPhoneNumberRepository) GetByID(ctx context.Context, id domain.BusinessPhoneNumberID) (*domain.BusinessPhoneNumber, error) {
	var rec BusinessPhoneNumberRecord
	if err := r.db.WithContext(ctx).Where("external_id = ?", string(id)).First(&rec).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("get business phone: %w", err)
	}
	return r.toDomain(&rec), nil
}

func (r *BusinessPhoneNumberRepository) GetByPhone(ctx context.Context, phone domain.PhoneNumber) (*domain.BusinessPhoneNumber, error) {
	var rec BusinessPhoneNumberRecord
	if err := r.db.WithContext(ctx).Where("phone = ?", string(phone)).First(&rec).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("get business phone by number: %w", err)
	}
	return r.toDomain(&rec), nil
}

func (r *BusinessPhoneNumberRepository) GetByWABA(ctx context.Context, wabaID domain.WABAID) ([]*domain.BusinessPhoneNumber, error) {
	var recs []BusinessPhoneNumberRecord
	if err := r.db.WithContext(ctx).Where("waba_id = ?", string(wabaID)).Order("created_at ASC").Find(&recs).Error; err != nil {
		return nil, fmt.Errorf("get phones by waba: %w", err)
	}
	result := make([]*domain.BusinessPhoneNumber, len(recs))
	for i, rec := range recs {
		result[i] = r.toDomain(&rec)
	}
	return result, nil
}

func (r *BusinessPhoneNumberRepository) List(ctx context.Context) ([]*domain.BusinessPhoneNumber, error) {
	var recs []BusinessPhoneNumberRecord
	if err := r.db.WithContext(ctx).Order("created_at ASC").Find(&recs).Error; err != nil {
		return nil, fmt.Errorf("list business phones: %w", err)
	}
	result := make([]*domain.BusinessPhoneNumber, len(recs))
	for i, rec := range recs {
		result[i] = r.toDomain(&rec)
	}
	return result, nil
}

func (r *BusinessPhoneNumberRepository) toDomain(rec *BusinessPhoneNumberRecord) *domain.BusinessPhoneNumber {
	return &domain.BusinessPhoneNumber{
		ID:          domain.BusinessPhoneNumberID(rec.ExternalID),
		WABAID:      domain.WABAID(rec.WABAID),
		Phone:       domain.PhoneNumber(rec.Phone),
		DisplayName: rec.DisplayName,
		Country:     rec.Country,
		Status:      rec.Status,
		CreatedAt:   rec.CreatedAt,
	}
}

type ApplicationRepo struct {
	db *gorm.DB
}

func NewApplicationRepository(db *gorm.DB) *ApplicationRepo {
	return &ApplicationRepo{db: db}
}

func (r *ApplicationRepo) Create(ctx context.Context, a *domain.Application) error {
	rec := ApplicationRecord{
		ExternalID: string(a.ID),
		BusinessID: string(a.BusinessID),
		Name:       a.Name,
		Status:     a.Status,
		AppSecret:  a.AppSecret,
	}
	return r.db.WithContext(ctx).Create(&rec).Error
}

func (r *ApplicationRepo) GetByID(ctx context.Context, id domain.ApplicationID) (*domain.Application, error) {
	var rec ApplicationRecord
	if err := r.db.WithContext(ctx).Where("external_id = ?", string(id)).First(&rec).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("get application: %w", err)
	}
	return r.appToDomain(&rec), nil
}

func (r *ApplicationRepo) GetByBusiness(ctx context.Context, businessID domain.BusinessID) ([]*domain.Application, error) {
	var recs []ApplicationRecord
	if err := r.db.WithContext(ctx).Where("business_id = ?", string(businessID)).Order("created_at ASC").Find(&recs).Error; err != nil {
		return nil, fmt.Errorf("get apps by business: %w", err)
	}
	result := make([]*domain.Application, len(recs))
	for i, rec := range recs {
		result[i] = r.appToDomain(&rec)
	}
	return result, nil
}

func (r *ApplicationRepo) List(ctx context.Context) ([]*domain.Application, error) {
	var recs []ApplicationRecord
	if err := r.db.WithContext(ctx).Order("created_at ASC").Find(&recs).Error; err != nil {
		return nil, fmt.Errorf("list applications: %w", err)
	}
	result := make([]*domain.Application, len(recs))
	for i, rec := range recs {
		result[i] = r.appToDomain(&rec)
	}
	return result, nil
}

func (r *ApplicationRepo) appToDomain(rec *ApplicationRecord) *domain.Application {
	return &domain.Application{
		ID:         domain.ApplicationID(rec.ExternalID),
		BusinessID: domain.BusinessID(rec.BusinessID),
		Name:       rec.Name,
		Status:     rec.Status,
		AppSecret:  rec.AppSecret,
		CreatedAt:  rec.CreatedAt,
		UpdatedAt:  rec.UpdatedAt,
	}
}

type CredentialRepo struct {
	db *gorm.DB
}

func NewCredentialRepository(db *gorm.DB) *CredentialRepo {
	return &CredentialRepo{db: db}
}

func (r *CredentialRepo) Create(ctx context.Context, c *domain.Credential) error {
	rec := CredentialRecord{
		ExternalID:    string(c.ID),
		ApplicationID: string(c.ApplicationID),
		TokenHash:     c.TokenHash,
		Status:        c.Status,
	}
	return r.db.WithContext(ctx).Create(&rec).Error
}

func (r *CredentialRepo) GetByID(ctx context.Context, id domain.CredentialID) (*domain.Credential, error) {
	var rec CredentialRecord
	if err := r.db.WithContext(ctx).Where("external_id = ?", string(id)).First(&rec).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("get credential: %w", err)
	}
	return r.credToDomain(&rec), nil
}

func (r *CredentialRepo) GetByTokenHash(ctx context.Context, tokenHash string) (*domain.Credential, error) {
	var rec CredentialRecord
	if err := r.db.WithContext(ctx).Where("token_hash = ? AND status = ?", tokenHash, "active").First(&rec).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("get credential by token: %w", err)
	}
	return r.credToDomain(&rec), nil
}

func (r *CredentialRepo) GetByApplication(ctx context.Context, appID domain.ApplicationID) ([]*domain.Credential, error) {
	var recs []CredentialRecord
	if err := r.db.WithContext(ctx).Where("application_id = ?", string(appID)).Order("created_at ASC").Find(&recs).Error; err != nil {
		return nil, fmt.Errorf("get credentials by app: %w", err)
	}
	result := make([]*domain.Credential, len(recs))
	for i, rec := range recs {
		result[i] = r.credToDomain(&rec)
	}
	return result, nil
}

func (r *CredentialRepo) Revoke(ctx context.Context, id domain.CredentialID) error {
	return r.db.WithContext(ctx).Model(&CredentialRecord{}).Where("external_id = ?", string(id)).Update("status", "revoked").Error
}

func (r *CredentialRepo) credToDomain(rec *CredentialRecord) *domain.Credential {
	return &domain.Credential{
		ID:            domain.CredentialID(rec.ExternalID),
		ApplicationID: domain.ApplicationID(rec.ApplicationID),
		TokenHash:     rec.TokenHash,
		Status:        rec.Status,
		CreatedAt:     rec.CreatedAt,
		RevokedAt:     rec.RevokedAt,
	}
}

type WebhookConfigRepo struct {
	db *gorm.DB
}

func NewWebhookConfigRepository(db *gorm.DB) *WebhookConfigRepo {
	return &WebhookConfigRepo{db: db}
}

func (r *WebhookConfigRepo) Create(ctx context.Context, w *domain.WebhookConfig) error {
	rec := WebhookConfigRecord{
		ApplicationID: string(w.ApplicationID),
		CallbackURL:   w.CallbackURL,
		VerifyToken:   w.VerifyToken,
		AppSecret:     w.AppSecret,
		Enabled:       w.Enabled,
	}
	return r.db.WithContext(ctx).Create(&rec).Error
}

func (r *WebhookConfigRepo) GetByApplication(ctx context.Context, appID domain.ApplicationID) (*domain.WebhookConfig, error) {
	var rec WebhookConfigRecord
	if err := r.db.WithContext(ctx).Where("application_id = ?", string(appID)).First(&rec).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("get webhook config: %w", err)
	}
	return r.whToDomain(&rec), nil
}

func (r *WebhookConfigRepo) Update(ctx context.Context, w *domain.WebhookConfig) error {
	return r.db.WithContext(ctx).Model(&WebhookConfigRecord{}).Where("application_id = ?", string(w.ApplicationID)).Updates(map[string]interface{}{
		"callback_url": w.CallbackURL,
		"verify_token": w.VerifyToken,
		"app_secret":   w.AppSecret,
		"enabled":      w.Enabled,
	}).Error
}

func (r *WebhookConfigRepo) whToDomain(rec *WebhookConfigRecord) *domain.WebhookConfig {
	return &domain.WebhookConfig{
		ID:            domain.ApplicationID(rec.ApplicationID),
		ApplicationID: domain.ApplicationID(rec.ApplicationID),
		CallbackURL:   rec.CallbackURL,
		VerifyToken:   rec.VerifyToken,
		AppSecret:     rec.AppSecret,
		Enabled:       rec.Enabled,
		CreatedAt:     rec.CreatedAt,
		UpdatedAt:     rec.UpdatedAt,
	}
}
