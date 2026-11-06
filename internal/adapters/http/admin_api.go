package http

import (
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/wabasi/wabasi/internal/application"
	"github.com/wabasi/wabasi/internal/domain"
	"github.com/wabasi/wabasi/internal/ports"
)

type AdminAPIHandler struct {
	admin         *application.AdminService
	messaging     *application.MessagingService
	observability *application.ObservabilityService
}

func NewAdminAPIHandler(admin *application.AdminService, messaging *application.MessagingService) *AdminAPIHandler {
	return &AdminAPIHandler{admin: admin, messaging: messaging}
}

func (h *AdminAPIHandler) SetObservabilityService(obs *application.ObservabilityService) {
	h.observability = obs
}

func (h *AdminAPIHandler) Register(app *fiber.App) {
	api := app.Group("/api/admin")

	// Auth middleware on all /api/admin/* routes
	api.Use(h.adminAuthMiddleware())

	// --- API Keys ---
	api.Post("/api-keys", h.handleCreateAPIKey)
	api.Get("/api-keys", h.handleListAPIKeys)
	api.Post("/api-keys/:id/revoke", h.handleRevokeAPIKey)

	// --- Businesses ---
	api.Get("/businesses", h.handleBusinessesJSON)
	api.Post("/businesses", h.handleCreateBusinessJSON)

	// --- WABAs ---
	api.Get("/wabas", h.handleWABAsJSON)
	api.Post("/wabas", h.handleCreateWABAJSON)

	// --- Phone Numbers ---
	api.Get("/phones", h.handlePhonesJSON)
	api.Post("/phones", h.handleProvisionPhoneJSON)

	// --- Applications ---
	api.Get("/applications", h.handleApplicationsJSON)
	api.Post("/applications", h.handleCreateApplicationJSON)

	// --- Credentials ---
	api.Get("/credentials", h.handleCredentialsJSON)
	api.Post("/credentials", h.handleIssueCredentialJSON)
	api.Post("/credentials/:id/revoke", h.handleRevokeCredentialJSON)

	// --- Webhooks ---
	api.Get("/webhooks", h.handleWebhooksJSON)
	api.Post("/webhooks", h.handleConfigureWebhookJSON)

	// --- Observability ---
	api.Get("/events", h.handleEventsJSON)
	api.Get("/events/:id", h.handleEventByIDJSON)
	api.Get("/traces/:trace_id", h.handleTraceJSON)
	api.Get("/metrics", h.handleMetricsJSON)
}

func (h *AdminAPIHandler) adminAuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if h.admin == nil {
			return c.Next()
		}

		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"error": fiber.Map{
					"code":    190,
					"message": "(#190) Missing authorization token",
					"type":    "OAuthException",
				},
			})
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == authHeader {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"error": fiber.Map{
					"code":    190,
					"message": "(#190) Invalid authorization header format",
					"type":    "OAuthException",
				},
			})
		}

		result, err := h.admin.AuthorizeAPIKey(c.Context(), token)
		if err != nil {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"error": fiber.Map{
					"code":    190,
					"message": "(#190) Invalid or revoked API key",
					"type":    "OAuthException",
				},
			})
		}

		c.Locals("api_key", result.APIKey)
		return c.Next()
	}
}

// --- Scope helpers ---

func requireScope(c *fiber.Ctx, scope domain.APIKeyScope) bool {
	key := c.Locals("api_key").(*domain.APIKey)
	if !key.HasScope(scope) {
		c.Status(http.StatusForbidden).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    190,
				"message": "(#190) Insufficient scope: requires " + string(scope),
				"type":    "OAuthException",
			},
		})
		return false
	}
	return true
}

// --- API Keys ---

func (h *AdminAPIHandler) handleCreateAPIKey(c *fiber.Ctx) error {
	if !requireScope(c, domain.ScopeAdminWrite) {
		return nil
	}

	var req struct {
		Name   string   `json:"name"`
		Scopes []string `json:"scopes"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if req.Name == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "name is required"})
	}

	scopes := make([]domain.APIKeyScope, len(req.Scopes))
	for i, s := range req.Scopes {
		sc := domain.APIKeyScope(s)
		if !sc.IsValid() {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid scope: " + s,
			})
		}
		scopes[i] = sc
	}

	_, plain, err := h.admin.CreateAPIKey(c.Context(), req.Name, scopes)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Re-fetch to get the created key (without the plain secret)
	keys, _ := h.admin.ListAPIKeys(c.Context())
	created := keys[0] // most recent

	return c.Status(http.StatusCreated).JSON(fiber.Map{
		"api_key": created,
		"secret":  plain,
	})
}

func (h *AdminAPIHandler) handleListAPIKeys(c *fiber.Ctx) error {
	if !requireScope(c, domain.ScopeAdminRead) {
		return nil
	}

	keys, err := h.admin.ListAPIKeys(c.Context())
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"api_keys": keys})
}

func (h *AdminAPIHandler) handleRevokeAPIKey(c *fiber.Ctx) error {
	if !requireScope(c, domain.ScopeAdminWrite) {
		return nil
	}

	id := c.Params("id")
	if err := h.admin.RevokeAPIKey(c.Context(), domain.APIKeyID(id)); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"ok": true})
}

// --- Businesses ---

func (h *AdminAPIHandler) handleBusinessesJSON(c *fiber.Ctx) error {
	if !requireScope(c, domain.ScopeAdminRead) {
		return nil
	}

	businesses, err := h.admin.ListBusinesses(c.Context())
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"businesses": businesses})
}

func (h *AdminAPIHandler) handleCreateBusinessJSON(c *fiber.Ctx) error {
	if !requireScope(c, domain.ScopeAdminWrite) {
		return nil
	}

	var req struct {
		Name        string `json:"name"`
		DisplayName string `json:"display_name"`
		Country     string `json:"country"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	b, err := h.admin.CreateBusiness(c.Context(), req.Name, req.DisplayName, req.Country)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(http.StatusCreated).JSON(fiber.Map{"business": b})
}

// --- WABAs ---

func (h *AdminAPIHandler) handleWABAsJSON(c *fiber.Ctx) error {
	if !requireScope(c, domain.ScopeAdminRead) {
		return nil
	}

	wabas, err := h.admin.ListWABAs(c.Context())
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"wabas": wabas})
}

func (h *AdminAPIHandler) handleCreateWABAJSON(c *fiber.Ctx) error {
	if !requireScope(c, domain.ScopeAdminWrite) {
		return nil
	}

	var req struct {
		BusinessID string `json:"business_id"`
		Name       string `json:"name"`
		Timezone   string `json:"timezone"`
		Currency   string `json:"currency"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	w, err := h.admin.CreateWABA(c.Context(), domain.BusinessID(req.BusinessID), req.Name, req.Timezone, req.Currency)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(http.StatusCreated).JSON(fiber.Map{"waba": w})
}

// --- Phone Numbers ---

func (h *AdminAPIHandler) handlePhonesJSON(c *fiber.Ctx) error {
	if !requireScope(c, domain.ScopeAdminRead) {
		return nil
	}

	phones, err := h.admin.ListPhoneNumbers(c.Context())
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"phone_numbers": phones})
}

func (h *AdminAPIHandler) handleProvisionPhoneJSON(c *fiber.Ctx) error {
	if !requireScope(c, domain.ScopeAdminWrite) {
		return nil
	}

	var req struct {
		WABAID      string `json:"waba_id"`
		Phone       string `json:"phone"`
		DisplayName string `json:"display_name"`
		Country     string `json:"country"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	p, err := h.admin.ProvisionPhoneNumber(c.Context(), domain.WABAID(req.WABAID), domain.PhoneNumber(req.Phone), req.DisplayName, req.Country)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(http.StatusCreated).JSON(fiber.Map{"phone_number": p})
}

// --- Applications ---

func (h *AdminAPIHandler) handleApplicationsJSON(c *fiber.Ctx) error {
	if !requireScope(c, domain.ScopeAdminRead) {
		return nil
	}

	apps, err := h.admin.ListApplications(c.Context())
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"applications": apps})
}

func (h *AdminAPIHandler) handleCreateApplicationJSON(c *fiber.Ctx) error {
	if !requireScope(c, domain.ScopeAdminWrite) {
		return nil
	}

	var req struct {
		BusinessID string `json:"business_id"`
		Name       string `json:"name"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	a, err := h.admin.CreateApplication(c.Context(), domain.BusinessID(req.BusinessID), req.Name)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(http.StatusCreated).JSON(fiber.Map{"application": a})
}

// --- Credentials ---

func (h *AdminAPIHandler) handleCredentialsJSON(c *fiber.Ctx) error {
	if !requireScope(c, domain.ScopeAdminRead) {
		return nil
	}

	apps, err := h.admin.ListApplications(c.Context())
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	type credJSON struct {
		ID            string `json:"id"`
		ApplicationID string `json:"application_id"`
		Status        string `json:"status"`
	}

	var allCreds []credJSON
	for _, a := range apps {
		creds, _ := h.admin.ListCredentials(c.Context(), a.ID)
		for _, cr := range creds {
			allCreds = append(allCreds, credJSON{
				ID:            string(cr.ID),
				ApplicationID: string(cr.ApplicationID),
				Status:        cr.Status,
			})
		}
	}
	return c.JSON(fiber.Map{"credentials": allCreds})
}

func (h *AdminAPIHandler) handleIssueCredentialJSON(c *fiber.Ctx) error {
	if !requireScope(c, domain.ScopeAdminWrite) {
		return nil
	}

	var req struct {
		ApplicationID string `json:"application_id"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	cred, plain, err := h.admin.IssueCredential(c.Context(), domain.ApplicationID(req.ApplicationID))
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(http.StatusCreated).JSON(fiber.Map{
		"credential_id": string(cred.ID),
		"token":         plain,
	})
}

func (h *AdminAPIHandler) handleRevokeCredentialJSON(c *fiber.Ctx) error {
	if !requireScope(c, domain.ScopeAdminWrite) {
		return nil
	}

	id := c.Params("id")
	if err := h.admin.RevokeCredential(c.Context(), domain.CredentialID(id)); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"ok": true})
}

// --- Webhooks ---

func (h *AdminAPIHandler) handleWebhooksJSON(c *fiber.Ctx) error {
	if !requireScope(c, domain.ScopeAdminRead) {
		return nil
	}

	apps, err := h.admin.ListApplications(c.Context())
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	type webhookJSON struct {
		ApplicationID string `json:"application_id"`
		CallbackURL   string `json:"callback_url"`
		VerifyToken   string `json:"verify_token"`
		Enabled       bool   `json:"enabled"`
	}

	var allWebhooks []webhookJSON
	for _, a := range apps {
		wh, _ := h.admin.GetWebhookConfig(c.Context(), a.ID)
		if wh != nil {
			allWebhooks = append(allWebhooks, webhookJSON{
				ApplicationID: string(wh.ApplicationID),
				CallbackURL:   wh.CallbackURL,
				VerifyToken:   wh.VerifyToken,
				Enabled:       wh.Enabled,
			})
		}
	}
	return c.JSON(fiber.Map{"webhook_configs": allWebhooks})
}

func (h *AdminAPIHandler) handleConfigureWebhookJSON(c *fiber.Ctx) error {
	if !requireScope(c, domain.ScopeAdminWrite) {
		return nil
	}

	var req struct {
		ApplicationID string `json:"application_id"`
		CallbackURL   string `json:"callback_url"`
		VerifyToken   string `json:"verify_token"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	wh, err := h.admin.ConfigureWebhook(c.Context(), domain.ApplicationID(req.ApplicationID), req.CallbackURL, req.VerifyToken)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(http.StatusCreated).JSON(fiber.Map{"webhook_config": wh})
}

// --- Observability ---

func (h *AdminAPIHandler) handleEventsJSON(c *fiber.Ctx) error {
	if !requireScope(c, domain.ScopeAdminObs) {
		return nil
	}

	if h.observability == nil {
		return c.Status(http.StatusServiceUnavailable).JSON(fiber.Map{"error": "observability not available"})
	}

	result, err := h.observability.ListEvents(c.Context(), ports.EventFilter{
		Limit: 50,
		Offset: 0,
	})
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"events": result.Events, "total": result.Total})
}

func (h *AdminAPIHandler) handleEventByIDJSON(c *fiber.Ctx) error {
	if !requireScope(c, domain.ScopeAdminObs) {
		return nil
	}

	if h.observability == nil {
		return c.Status(http.StatusServiceUnavailable).JSON(fiber.Map{"error": "observability not available"})
	}

	id := c.Params("id")
	event, err := h.observability.GetTrace(c.Context(), domain.TraceID(id))
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "trace not found"})
	}
	return c.JSON(fiber.Map{"events": event})
}

func (h *AdminAPIHandler) handleTraceJSON(c *fiber.Ctx) error {
	if !requireScope(c, domain.ScopeAdminObs) {
		return nil
	}

	if h.observability == nil {
		return c.Status(http.StatusServiceUnavailable).JSON(fiber.Map{"error": "observability not available"})
	}

	traceID := c.Params("trace_id")
	events, err := h.observability.GetTrace(c.Context(), domain.TraceID(traceID))
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "trace not found"})
	}
	return c.JSON(fiber.Map{"trace_id": traceID, "events": events})
}

func (h *AdminAPIHandler) handleMetricsJSON(c *fiber.Ctx) error {
	if !requireScope(c, domain.ScopeAdminObs) {
		return nil
	}

	if h.observability == nil {
		return c.Status(http.StatusServiceUnavailable).JSON(fiber.Map{"error": "observability not available"})
	}

	metrics, err := h.observability.GetMetricsSummary(c.Context(), 24*time.Hour)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(metrics)
}
