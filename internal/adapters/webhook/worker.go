package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"gorm.io/gorm"
)

type WebhookDelivery struct {
	ID            string
	ApplicationID string
	CallbackURL   string
	EventType     string
	Payload       string
	Status        string
	Attempts      int
	MaxAttempts   int
	NextRetryAt   *time.Time
	LastStatusCode int
	LastError     string
	Signature     string
}

type DeliveryWorker struct {
	db        *gorm.DB
	client    *http.Client
	logger    *slog.Logger
	interval  time.Duration
	running   bool
}

func NewDeliveryWorker(db *gorm.DB, logger *slog.Logger) *DeliveryWorker {
	return &DeliveryWorker{
		db: db,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger:   logger,
		interval: 2 * time.Second,
	}
}

func (w *DeliveryWorker) Start(ctx context.Context) {
	w.running = true
	w.logger.Info("webhook delivery worker started")
	go w.run(ctx)
}

func (w *DeliveryWorker) Stop() {
	w.running = false
	w.logger.Info("webhook delivery worker stopped")
}

func (w *DeliveryWorker) run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if !w.running {
				return
			}
			w.processPending(ctx)
		}
	}
}

func (w *DeliveryWorker) processPending(ctx context.Context) {
	var deliveries []WebhookDelivery
	err := w.db.WithContext(ctx).
		Table("webhook_deliveries").
		Where("status IN ? AND (next_retry_at IS NULL OR next_retry_at <= ?)",
			[]string{"pending", "retrying"}, time.Now()).
		Order("created_at ASC").
		Limit(10).
		Find(&deliveries).Error
	if err != nil {
		w.logger.Error("failed to fetch pending webhooks", "error", err)
		return
	}

	for i := range deliveries {
		w.deliver(ctx, &deliveries[i])
	}
}

func (w *DeliveryWorker) deliver(ctx context.Context, d *WebhookDelivery) {
	if d.CallbackURL == "" {
		w.logger.Warn("webhook has no callback URL, marking failed", "id", d.ID)
		w.markFailed(d, 0, "no callback URL")
		return
	}

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(d.Payload), &payload); err != nil {
		w.logger.Error("failed to unmarshal webhook payload", "error", err, "id", d.ID)
		w.markFailed(d, 0, "invalid payload")
		return
	}

	body, err := json.Marshal(payload)
	if err != nil {
		w.logger.Error("failed to marshal webhook payload", "error", err, "id", d.ID)
		w.markFailed(d, 0, "marshal error")
		return
	}

	signature := computeSignature(body, d.Signature)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, d.CallbackURL, bytes.NewReader(body))
	if err != nil {
		w.logger.Error("failed to create webhook request", "error", err, "id", d.ID)
		w.markFailed(d, 0, "request creation error")
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Hub-Signature-256", "sha256="+signature)
	req.Header.Set("User-Agent", "Wabasi-Webhook/1.0")

	resp, err := w.client.Do(req)
	if err != nil {
		w.logger.Error("webhook delivery failed", "error", err, "id", d.ID, "url", d.CallbackURL)
		w.scheduleRetry(d, 0, err.Error())
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		w.logger.Info("webhook delivered",
			"id", d.ID,
			"url", d.CallbackURL,
			"status", resp.StatusCode,
			"attempts", d.Attempts+1,
		)
		d.Status = "delivered"
		d.Attempts++
		d.LastStatusCode = resp.StatusCode
		w.updateDelivery(d)
		return
	}

	w.logger.Warn("webhook returned error",
		"id", d.ID,
		"url", d.CallbackURL,
		"status", resp.StatusCode,
		"body", string(respBody),
	)
	w.scheduleRetry(d, resp.StatusCode, fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(respBody)))
}

func (w *DeliveryWorker) scheduleRetry(d *WebhookDelivery, statusCode int, errMsg string) {
	d.Attempts++
	d.LastStatusCode = statusCode
	d.LastError = errMsg

	if d.Attempts >= d.MaxAttempts {
		d.Status = "failed"
		w.logger.Error("webhook delivery permanently failed",
			"id", d.ID,
			"attempts", d.Attempts,
			"url", d.CallbackURL,
		)
	} else {
		d.Status = "retrying"
		backoff := time.Duration(d.Attempts*d.Attempts) * time.Second
		if backoff > 5*time.Minute {
			backoff = 5 * time.Minute
		}
		next := time.Now().Add(backoff)
		d.NextRetryAt = &next
	}

	w.updateDelivery(d)
}

func (w *DeliveryWorker) markFailed(d *WebhookDelivery, statusCode int, errMsg string) {
	d.Status = "failed"
	d.Attempts++
	d.LastStatusCode = statusCode
	d.LastError = errMsg
	w.updateDelivery(d)
}

func (w *DeliveryWorker) updateDelivery(d *WebhookDelivery) {
	w.db.Table("webhook_deliveries").Where("id = ?", d.ID).Updates(map[string]interface{}{
		"status":           d.Status,
		"attempts":         d.Attempts,
		"last_status_code": d.LastStatusCode,
		"last_error":       d.LastError,
		"next_retry_at":    d.NextRetryAt,
	})
}

func computeSignature(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}
