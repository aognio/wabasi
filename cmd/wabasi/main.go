package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/wabasi/wabasi/internal/adapters/http"
	wabasiMCP "github.com/wabasi/wabasi/internal/adapters/mcp"
	sqliteAdapter "github.com/wabasi/wabasi/internal/adapters/sqlite"
	"github.com/wabasi/wabasi/internal/adapters/webhook"
	"github.com/wabasi/wabasi/internal/application"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	dbPath := "var/wabasi.db"
	mcpMode := false

	for _, arg := range os.Args[1:] {
		if arg == "--mcp" {
			mcpMode = true
		} else if arg[0] != '-' {
			dbPath = arg
		}
	}

	db, err := sqliteAdapter.Open(dbPath)
	if err != nil {
		logger.Error("failed to open database", "error", err)
		os.Exit(1)
	}

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

	mediaStore := sqliteAdapter.NewMediaStore(db, "var")

	obsRepo := sqliteAdapter.NewEventRepository(db)

	messagingSvc := application.NewMessagingService(
		messageRepo,
		conversationRepo,
		identityRepo,
		phoneRepo,
		eventPub,
		webhookDisp,
		logger,
	)

	phoneReg := application.NewPhoneRegistry(phoneRepo)

	adminSvc := application.NewAdminService(
		businessRepo,
		wabaRepo,
		businessPhoneRepo,
		appRepo,
		credentialRepo,
		webhookConfigRepo,
		apiKeyRepo,
		phoneReg,
		eventPub,
		logger,
	)

	statusSim := application.NewStatusSimulator(
		messageRepo,
		conversationRepo,
		eventPub,
		webhookDisp,
		logger,
	)

	obsSvc := application.NewObservabilityService(
		obsRepo,
		messageRepo,
		conversationRepo,
		identityRepo,
	)

	deliveryWorker := webhook.NewDeliveryWorker(db, logger)
	deliveryWorker.Start(context.Background())
	statusSim.Start(context.Background())

	if mcpMode {
		mcpServer := wabasiMCP.NewMCPServer(
			obsSvc,
			messageRepo,
			conversationRepo,
			identityRepo,
			obsRepo,
		)
		if err := mcpServer.Serve_stdio(); err != nil {
			logger.Error("mcp server error", "error", err)
			os.Exit(1)
		}
		return
	}

	server := http.NewServer(messagingSvc, phoneReg, identityRepo, conversationRepo, messageRepo, logger)
	server.SetAdminService(adminSvc)
	server.SetMediaStore(mediaStore)
	server.SetObservabilityService(obsSvc)
	server.RegisterAdmin(adminSvc, messagingSvc)

	addr := ":8080"
	if port := os.Getenv("PORT"); port != "" {
		addr = ":" + port
	}

	logger.Info("starting wabasi", "addr", addr)
	fmt.Println("Wabasi is running at http://localhost" + addr)

	if err := server.Listen(addr); err != nil {
		logger.Error("server error", "error", err)
		os.Exit(1)
	}
}
