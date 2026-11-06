package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	adapterHTTP "github.com/wabasi/wabasi/internal/adapters/http"
	sqliteAdapter "github.com/wabasi/wabasi/internal/adapters/sqlite"
	"github.com/wabasi/wabasi/internal/application"
	"log/slog"
	"os"
)

func setupTestServer(t *testing.T) (*adapterHTTP.Server, func()) {
	t.Helper()

	db, err := sqliteAdapter.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	messageRepo := sqliteAdapter.NewMessageRepository(db)
	conversationRepo := sqliteAdapter.NewConversationRepository(db)
	identityRepo := sqliteAdapter.NewVirtualIdentityRepository(db)
	phoneRepo := sqliteAdapter.NewPhoneNumberRepository(db)
	eventPub := sqliteAdapter.NewEventPublisher(db, logger)
	webhookDisp := sqliteAdapter.NewWebhookDispatcher(db, logger)

	messagingSvc := application.NewMessagingService(
		messageRepo, conversationRepo, identityRepo, phoneRepo, eventPub, webhookDisp, logger,
	)
	phoneReg := application.NewPhoneRegistry(phoneRepo)

	server := adapterHTTP.NewServer(messagingSvc, phoneReg, identityRepo, conversationRepo, messageRepo, logger)

	return server, func() {}
}

func setupTestWithCredential(t *testing.T) (*adapterHTTP.Server, string, func()) {
	t.Helper()

	db, err := sqliteAdapter.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	messageRepo := sqliteAdapter.NewMessageRepository(db)
	conversationRepo := sqliteAdapter.NewConversationRepository(db)
	identityRepo := sqliteAdapter.NewVirtualIdentityRepository(db)
	phoneRepo := sqliteAdapter.NewPhoneNumberRepository(db)
	eventPub := sqliteAdapter.NewEventPublisher(db, logger)
	webhookDisp := sqliteAdapter.NewWebhookDispatcher(db, logger)

	businessRepo := sqliteAdapter.NewBusinessRepository(db)
	wabaRepo := sqliteAdapter.NewWABARepository(db)
	businessPhoneRepo := sqliteAdapter.NewBusinessPhoneNumberRepository(db)
	appRepo := sqliteAdapter.NewApplicationRepository(db)
	credentialRepo := sqliteAdapter.NewCredentialRepository(db)
	webhookConfigRepo := sqliteAdapter.NewWebhookConfigRepository(db)
	apiKeyRepo := sqliteAdapter.NewAPIKeyRepository(db)

	messagingSvc := application.NewMessagingService(
		messageRepo, conversationRepo, identityRepo, phoneRepo, eventPub, webhookDisp, logger,
	)
	phoneReg := application.NewPhoneRegistry(phoneRepo)

	adminSvc := application.NewAdminService(
		businessRepo, wabaRepo, businessPhoneRepo, appRepo,
		credentialRepo, webhookConfigRepo, apiKeyRepo, phoneReg, eventPub, logger,
	)

	// Create a business with WABA and phone number
	biz, _ := adminSvc.CreateBusiness(context.Background(), "Test Biz", "Test", "PE")
	waba, _ := adminSvc.CreateWABA(context.Background(), biz.ID, "Test WABA", "America/Lima", "PEN")
	_, _ = adminSvc.ProvisionPhoneNumber(context.Background(), waba.ID, "+5115550100", "Test Phone", "PE")

	// Create an application and issue a credential
	app, _ := adminSvc.CreateApplication(context.Background(), biz.ID, "Test App")
	_, token, _ := adminSvc.IssueCredential(context.Background(), app.ID)

	server := adapterHTTP.NewServer(messagingSvc, phoneReg, identityRepo, conversationRepo, messageRepo, logger)
	server.SetAdminService(adminSvc)

	return server, token, func() {}
}

func TestCreateIdentity(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	body, _ := json.Marshal(map[string]string{
		"phone":        "+51987654321",
		"display_name": "Alice",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/identities", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := server.App().Test(req)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	if result["phone"] != "+51987654321" {
		t.Errorf("expected phone +51987654321, got %v", result["phone"])
	}
	if result["display_name"] != "Alice" {
		t.Errorf("expected display_name Alice, got %v", result["display_name"])
	}
}

func TestDuplicatePhoneRejection(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	body, _ := json.Marshal(map[string]string{
		"phone":        "+51987654321",
		"display_name": "Alice",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/identities", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := server.App().Test(req)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("first create: expected 201, got %d", resp.StatusCode)
	}

	req2 := httptest.NewRequest(http.MethodPost, "/api/identities", bytes.NewReader(body))
	req2.Header.Set("Content-Type", "application/json")
	resp2, _ := server.App().Test(req2)
	if resp2.StatusCode != http.StatusConflict {
		t.Fatalf("duplicate: expected 409, got %d", resp2.StatusCode)
	}
}

func TestSendMessageToIdentity(t *testing.T) {
	server, token, cleanup := setupTestWithCredential(t)
	defer cleanup()

	createBody, _ := json.Marshal(map[string]string{
		"phone":        "+51987654321",
		"display_name": "Alice",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/identities", bytes.NewReader(createBody))
	req.Header.Set("Content-Type", "application/json")
	server.App().Test(req)

	sendBody, _ := json.Marshal(map[string]interface{}{
		"messaging_product": "whatsapp",
		"recipient_type":    "individual",
		"to":                "+51987654321",
		"type":              "text",
		"text": map[string]interface{}{
			"preview_url": false,
			"body":        "Hello Alice!",
		},
	})

	req2 := httptest.NewRequest(http.MethodPost, "/v18.0/+5115550100/messages", bytes.NewReader(sendBody))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Authorization", "Bearer "+token)
	resp, err := server.App().Test(req2)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	if result["messaging_product"] != "whatsapp" {
		t.Errorf("expected messaging_product whatsapp, got %v", result["messaging_product"])
	}

	messages := result["messages"].([]interface{})
	if len(messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(messages))
	}

	msg := messages[0].(map[string]interface{})
	if msg["id"] == nil || msg["id"] == "" {
		t.Error("expected non-empty message id")
	}
}

func TestInboundWebhook(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	createBody, _ := json.Marshal(map[string]string{
		"phone":        "+51987654321",
		"display_name": "Alice",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/identities", bytes.NewReader(createBody))
	req.Header.Set("Content-Type", "application/json")
	server.App().Test(req)

	webhookPayload := map[string]interface{}{
		"object": "whatsapp_business_account",
		"entry": []map[string]interface{}{
			{
				"id": "waba-001",
				"changes": []map[string]interface{}{
					{
						"value": map[string]interface{}{
							"messaging_product": "whatsapp",
							"metadata": map[string]interface{}{
								"display_phone_number": "15550783881",
								"phone_number_id":      "biz_phone_001",
							},
							"contacts": []map[string]interface{}{
								{"profile": map[string]interface{}{"name": "Alice"}, "wa_id": "51987654321"},
							},
							"messages": []map[string]interface{}{
								{"from": "51987654321", "id": "wamid.inbound-test", "timestamp": "1757699100", "type": "text", "text": map[string]interface{}{"body": "Hello from Alice!"}},
							},
						},
						"field": "messages",
					},
				},
			},
		},
	}

	body, _ := json.Marshal(webhookPayload)
	req2 := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	req2.Header.Set("Content-Type", "application/json")
	resp, err := server.App().Test(req2)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestWebhookVerification(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/webhook?hub.mode=subscribe&hub.verify_token=wabasi-verify-token&hub.challenge=12345", nil)
	resp, err := server.App().Test(req)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	if buf.String() != "12345" {
		t.Errorf("expected challenge 12345, got %s", buf.String())
	}
}

func TestWebhookVerificationInvalidToken(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/webhook?hub.mode=subscribe&hub.verify_token=wrong-token&hub.challenge=12345", nil)
	resp, err := server.App().Test(req)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", resp.StatusCode)
	}
}

func TestMissingAuthorization(t *testing.T) {
	server, _, cleanup := setupTestWithCredential(t)
	defer cleanup()

	sendBody, _ := json.Marshal(map[string]interface{}{
		"messaging_product": "whatsapp",
		"to":                "+51987654321",
		"type":              "text",
		"text":              map[string]interface{}{"body": "test"},
	})

	req := httptest.NewRequest(http.MethodPost, "/v18.0/+5115550100/messages", bytes.NewReader(sendBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err := server.App().Test(req)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestEndToEndFlow(t *testing.T) {
	server, token, cleanup := setupTestWithCredential(t)
	defer cleanup()

	// 1. Create Alice
	createBody, _ := json.Marshal(map[string]string{"phone": "+51987654321", "display_name": "Alice"})
	req := httptest.NewRequest(http.MethodPost, "/api/identities", bytes.NewReader(createBody))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := server.App().Test(req)
	var alice map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&alice)

	// 2. Send message to Alice
	sendBody, _ := json.Marshal(map[string]interface{}{
		"messaging_product": "whatsapp",
		"to":                "+51987654321",
		"type":              "text",
		"text":              map[string]interface{}{"body": "Hello Alice!"},
	})
	req2 := httptest.NewRequest(http.MethodPost, "/v18.0/+5115550100/messages", bytes.NewReader(sendBody))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Authorization", "Bearer "+token)
	resp2, _ := server.App().Test(req2)
	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("send message: expected 200, got %d", resp2.StatusCode)
	}

	// 3. Alice sends inbound message
	webhookPayload := map[string]interface{}{
		"object": "whatsapp_business_account",
		"entry": []map[string]interface{}{
			{"id": "waba-001", "changes": []map[string]interface{}{
				{"value": map[string]interface{}{
					"messaging_product": "whatsapp",
					"metadata":          map[string]interface{}{"phone_number_id": "+5115550100"},
					"contacts":          []map[string]interface{}{{"profile": map[string]interface{}{"name": "Alice"}, "wa_id": "51987654321"}},
					"messages":          []map[string]interface{}{{"from": "51987654321", "id": "wamid.alice-reply", "timestamp": "1757699100", "type": "text", "text": map[string]interface{}{"body": "Hi there!"}}},
				}, "field": "messages"},
			}},
		},
	}
	body, _ := json.Marshal(webhookPayload)
	req3 := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	req3.Header.Set("Content-Type", "application/json")
	resp3, _ := server.App().Test(req3)
	if resp3.StatusCode != http.StatusOK {
		t.Fatalf("webhook: expected 200, got %d", resp3.StatusCode)
	}

	// 4. Check device conversations
	token2 := alice["device_token"].(string)
	req4 := httptest.NewRequest(http.MethodGet, "/device/"+token2+"/conversations", nil)
	resp4, _ := server.App().Test(req4)
	var convs []map[string]interface{}
	json.NewDecoder(resp4.Body).Decode(&convs)
	if len(convs) == 0 {
		t.Fatal("expected at least one conversation")
	}

	// 5. Check messages in conversation
	convID := convs[0]["id"].(string)
	req5 := httptest.NewRequest(http.MethodGet, "/device/"+token2+"/conversations/"+convID+"/messages", nil)
	resp5, _ := server.App().Test(req5)
	var msgs []map[string]interface{}
	json.NewDecoder(resp5.Body).Decode(&msgs)
	if len(msgs) < 2 {
		t.Fatalf("expected at least 2 messages, got %d", len(msgs))
	}
}
