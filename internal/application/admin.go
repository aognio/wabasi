package application

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/wabasi/wabasi/internal/domain"
	"github.com/wabasi/wabasi/internal/ports"
)

type AdminService struct {
	businesses   ports.BusinessRepository
	wabas        ports.WABARepository
	phones       ports.BusinessPhoneNumberRepository
	apps         ports.ApplicationRepository
	credentials  ports.CredentialRepository
	webhooks     ports.WebhookConfigRepository
	apiKeys      ports.APIKeyRepository
	phoneReg     *PhoneRegistry
	events       ports.EventPublisher
	logger       *slog.Logger
}

func NewAdminService(
	businesses ports.BusinessRepository,
	wabas ports.WABARepository,
	phones ports.BusinessPhoneNumberRepository,
	apps ports.ApplicationRepository,
	credentials ports.CredentialRepository,
	webhooks ports.WebhookConfigRepository,
	apiKeys ports.APIKeyRepository,
	phoneReg *PhoneRegistry,
	events ports.EventPublisher,
	logger *slog.Logger,
) *AdminService {
	return &AdminService{
		businesses:  businesses,
		wabas:       wabas,
		phones:      phones,
		apps:        apps,
		credentials: credentials,
		webhooks:    webhooks,
		apiKeys:     apiKeys,
		phoneReg:    phoneReg,
		events:      events,
		logger:      logger,
	}
}

// --- Business ---

func (s *AdminService) CreateBusiness(ctx context.Context, name, displayName, country string) (*domain.Business, error) {
	b := domain.NewBusiness(name, displayName, country)
	if err := s.businesses.Create(ctx, b); err != nil {
		return nil, fmt.Errorf("create business: %w", err)
	}
	s.events.Publish(ctx, domain.NewEvent(domain.EventBusinessCreated).
		WithSource(domain.SourceAdmin).
		WithActorID(string(b.ID)).
		WithPayload(b).
		Build())
	return b, nil
}

func (s *AdminService) GetBusiness(ctx context.Context, id domain.BusinessID) (*domain.Business, error) {
	return s.businesses.GetByID(ctx, id)
}

func (s *AdminService) ListBusinesses(ctx context.Context) ([]*domain.Business, error) {
	return s.businesses.List(ctx)
}

// --- WABA ---

func (s *AdminService) CreateWABA(ctx context.Context, businessID domain.BusinessID, name, timezone, currency string) (*domain.WABA, error) {
	b, err := s.businesses.GetByID(ctx, businessID)
	if err != nil || b == nil {
		return nil, fmt.Errorf("business not found")
	}
	w := domain.NewWABA(businessID, name, timezone, currency)
	if err := s.wabas.Create(ctx, w); err != nil {
		return nil, fmt.Errorf("create waba: %w", err)
	}
	s.events.Publish(ctx, domain.NewEvent(domain.EventWABACreated).
		WithSource(domain.SourceAdmin).
		WithActorID(string(w.ID)).
		WithPayload(w).
		Build())
	return w, nil
}

func (s *AdminService) GetWABA(ctx context.Context, id domain.WABAID) (*domain.WABA, error) {
	return s.wabas.GetByID(ctx, id)
}

func (s *AdminService) ListWABAs(ctx context.Context) ([]*domain.WABA, error) {
	return s.wabas.List(ctx)
}

func (s *AdminService) ListWABAsByBusiness(ctx context.Context, businessID domain.BusinessID) ([]*domain.WABA, error) {
	return s.wabas.GetByBusiness(ctx, businessID)
}

// --- Business Phone Numbers ---

func (s *AdminService) ProvisionPhoneNumber(ctx context.Context, wabaID domain.WABAID, phone domain.PhoneNumber, displayName, country string) (*domain.BusinessPhoneNumber, error) {
	w, err := s.wabas.GetByID(ctx, wabaID)
	if err != nil || w == nil {
		return nil, fmt.Errorf("waba not found")
	}

	available, err := s.phoneReg.IsAvailable(ctx, phone)
	if err != nil {
		return nil, fmt.Errorf("check availability: %w", err)
	}
	if !available {
		return nil, fmt.Errorf("phone number %s is already registered", phone)
	}

	p := domain.NewBusinessPhoneNumber(wabaID, phone, displayName, country)
	if err := s.phones.Create(ctx, p); err != nil {
		return nil, fmt.Errorf("provision phone: %w", err)
	}
	if err := s.phoneReg.Reserve(ctx, phone, "business_phone", string(p.ID)); err != nil {
		return nil, fmt.Errorf("reserve phone: %w", err)
	}
	s.events.Publish(ctx, domain.NewEvent(domain.EventBusinessPhoneProvisioned).
		WithSource(domain.SourceAdmin).
		WithActorID(string(p.ID)).
		WithPayload(p).
		Build())
	return p, nil
}

func (s *AdminService) GetPhoneNumber(ctx context.Context, id domain.BusinessPhoneNumberID) (*domain.BusinessPhoneNumber, error) {
	return s.phones.GetByID(ctx, id)
}

func (s *AdminService) ListPhoneNumbers(ctx context.Context) ([]*domain.BusinessPhoneNumber, error) {
	return s.phones.List(ctx)
}

// --- Applications ---

func (s *AdminService) CreateApplication(ctx context.Context, businessID domain.BusinessID, name string) (*domain.Application, error) {
	b, err := s.businesses.GetByID(ctx, businessID)
	if err != nil || b == nil {
		return nil, fmt.Errorf("business not found")
	}
	a := domain.NewApplication(businessID, name)
	if err := s.apps.Create(ctx, a); err != nil {
		return nil, fmt.Errorf("create application: %w", err)
	}
	s.events.Publish(ctx, domain.NewEvent(domain.EventApplicationCreated).
		WithSource(domain.SourceAdmin).
		WithActorID(string(a.ID)).
		WithApplicationID(string(a.ID)).
		WithPayload(a).
		Build())
	return a, nil
}

func (s *AdminService) GetApplication(ctx context.Context, id domain.ApplicationID) (*domain.Application, error) {
	return s.apps.GetByID(ctx, id)
}

func (s *AdminService) ListApplications(ctx context.Context) ([]*domain.Application, error) {
	return s.apps.List(ctx)
}

// --- Credentials ---

func (s *AdminService) IssueCredential(ctx context.Context, appID domain.ApplicationID) (*domain.Credential, string, error) {
	a, err := s.apps.GetByID(ctx, appID)
	if err != nil || a == nil {
		return nil, "", fmt.Errorf("application not found")
	}

	token := domain.GenerateMessageID()
	tokenHash := domain.HashToken(string(token))

	cred := domain.NewCredential(appID, tokenHash)
	if err := s.credentials.Create(ctx, cred); err != nil {
		return nil, "", fmt.Errorf("create credential: %w", err)
	}
	s.events.Publish(ctx, domain.NewEvent(domain.EventCredentialIssued).
		WithSource(domain.SourceAdmin).
		WithActorID(string(cred.ID)).
		WithApplicationID(string(appID)).
		WithPayload(cred).
		Build())
	return cred, string(token), nil
}

func (s *AdminService) GetCredential(ctx context.Context, id domain.CredentialID) (*domain.Credential, error) {
	return s.credentials.GetByID(ctx, id)
}

func (s *AdminService) ListCredentials(ctx context.Context, appID domain.ApplicationID) ([]*domain.Credential, error) {
	return s.credentials.GetByApplication(ctx, appID)
}

func (s *AdminService) RevokeCredential(ctx context.Context, id domain.CredentialID) error {
	if err := s.credentials.Revoke(ctx, id); err != nil {
		return fmt.Errorf("revoke credential: %w", err)
	}
	s.events.Publish(context.Background(), domain.NewEvent(domain.EventCredentialRevoked).
		WithSource(domain.SourceAdmin).
		WithActorID(string(id)).
		WithPayload(map[string]string{"id": string(id)}).
		Build())
	return nil
}

// --- Webhook Config ---

func (s *AdminService) ConfigureWebhook(ctx context.Context, appID domain.ApplicationID, callbackURL, verifyToken string) (*domain.WebhookConfig, error) {
	a, err := s.apps.GetByID(ctx, appID)
	if err != nil || a == nil {
		return nil, fmt.Errorf("application not found")
	}

	existing, _ := s.webhooks.GetByApplication(ctx, appID)
	if existing != nil {
		existing.CallbackURL = callbackURL
		existing.VerifyToken = verifyToken
		existing.AppSecret = a.AppSecret
		existing.Enabled = true
		if err := s.webhooks.Update(ctx, existing); err != nil {
			return nil, fmt.Errorf("update webhook: %w", err)
		}
		s.events.Publish(ctx, domain.NewEvent(domain.EventWebhookConfigured).
			WithSource(domain.SourceAdmin).
			WithApplicationID(string(appID)).
			WithPayload(existing).
			Build())
		return existing, nil
	}

	wh := domain.NewWebhookConfig(appID, callbackURL, verifyToken, a.AppSecret)
	if err := s.webhooks.Create(ctx, wh); err != nil {
		return nil, fmt.Errorf("create webhook: %w", err)
	}
	s.events.Publish(ctx, domain.NewEvent(domain.EventWebhookConfigured).
		WithSource(domain.SourceAdmin).
		WithApplicationID(string(appID)).
		WithPayload(wh).
		Build())
	return wh, nil
}

func (s *AdminService) GetWebhookConfig(ctx context.Context, appID domain.ApplicationID) (*domain.WebhookConfig, error) {
	return s.webhooks.GetByApplication(ctx, appID)
}

// --- Authorization ---

func (s *AdminService) AuthorizeToken(ctx context.Context, token string, phoneNumberID string) (*domain.AuthorizedPhoneResult, error) {
	tokenHash := domain.HashToken(token)

	cred, err := s.credentials.GetByTokenHash(ctx, tokenHash)
	if err != nil || cred == nil {
		return nil, fmt.Errorf("invalid token")
	}

	a, err := s.apps.GetByID(ctx, cred.ApplicationID)
	if err != nil || a == nil || a.Status != "active" {
		return nil, fmt.Errorf("application not found or inactive")
	}

	businessPhones, err := s.phones.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list phones: %w", err)
	}

	for _, bp := range businessPhones {
		if string(bp.ID) == phoneNumberID || string(bp.Phone) == phoneNumberID {
			waba, err := s.wabas.GetByID(ctx, bp.WABAID)
			if err != nil || waba == nil {
				continue
			}

			wh, _ := s.webhooks.GetByApplication(ctx, a.ID)

			return &domain.AuthorizedPhoneResult{
				Application:   a,
				WABA:          waba,
				PhoneNumber:   bp,
				WebhookConfig: wh,
			}, nil
		}
	}

	return nil, fmt.Errorf("phone_number_id not authorized for this application")
}

// --- API Keys ---

func (s *AdminService) CreateAPIKey(ctx context.Context, name string, scopes []domain.APIKeyScope) (*domain.APIKey, string, error) {
	ak, plain := domain.NewAPIKey(name, scopes)
	if err := s.apiKeys.Create(ctx, ak); err != nil {
		return nil, "", fmt.Errorf("create api key: %w", err)
	}
	s.events.Publish(ctx, domain.NewEvent(domain.EventAPIKeyCreated).
		WithSource(domain.SourceAdmin).
		WithActorID(string(ak.ID)).
		WithPayload(map[string]interface{}{
			"id":     string(ak.ID),
			"name":   ak.Name,
			"scopes": ak.Scopes,
		}).
		Build())
	return ak, plain, nil
}

func (s *AdminService) ListAPIKeys(ctx context.Context) ([]*domain.APIKey, error) {
	return s.apiKeys.List(ctx)
}

func (s *AdminService) RevokeAPIKey(ctx context.Context, id domain.APIKeyID) error {
	if err := s.apiKeys.Revoke(ctx, id); err != nil {
		return fmt.Errorf("revoke api key: %w", err)
	}
	s.events.Publish(context.Background(), domain.NewEvent(domain.EventAPIKeyRevoked).
		WithSource(domain.SourceAdmin).
		WithActorID(string(id)).
		WithPayload(map[string]string{"id": string(id)}).
		Build())
	return nil
}

func (s *AdminService) AuthorizeAPIKey(ctx context.Context, token string) (*domain.AuthorizedAPIKeyResult, error) {
	tokenHash := domain.HashToken(token)

	key, err := s.apiKeys.GetByKeyHash(ctx, tokenHash)
	if err != nil || key == nil {
		return nil, fmt.Errorf("invalid api key")
	}
	if key.Status != "active" {
		return nil, fmt.Errorf("api key revoked")
	}

	return &domain.AuthorizedAPIKeyResult{APIKey: key}, nil
}
