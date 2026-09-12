package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/wakaba/wakaba/internal/adapters/http"
	sqliteAdapter "github.com/wakaba/wakaba/internal/adapters/sqlite"
	"github.com/wakaba/wakaba/internal/application"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	dbPath := "var/wakaba.db"
	if len(os.Args) > 1 {
		dbPath = os.Args[1]
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

	server := http.NewServer(messagingSvc, phoneReg, identityRepo, conversationRepo, messageRepo, logger)

	addr := ":8080"
	if port := os.Getenv("PORT"); port != "" {
		addr = ":" + port
	}

	logger.Info("starting wakaba", "addr", addr)
	fmt.Println("Wakaba is running at http://localhost" + addr)

	if err := server.Listen(addr); err != nil {
		logger.Error("server error", "error", err)
		os.Exit(1)
	}
}
