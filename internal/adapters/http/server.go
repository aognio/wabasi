package http

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/wakaba/wakaba/internal/application"
	"github.com/wakaba/wakaba/internal/domain"
	"github.com/wakaba/wakaba/internal/ports"
)

type Server struct {
	app            *fiber.App
	messaging      *application.MessagingService
	phoneReg       *application.PhoneRegistry
	admin          *application.AdminService
	mediaStore     ports.MediaStore
	identities     ports.VirtualIdentityRepository
	conversations  ports.ConversationRepository
	messages       ports.MessageRepository
	logger         *slog.Logger
	sseClients     map[string][]chan string
	sseMu          sync.RWMutex
}

func NewServer(
	messaging *application.MessagingService,
	phoneReg *application.PhoneRegistry,
	identities ports.VirtualIdentityRepository,
	conversations ports.ConversationRepository,
	messages ports.MessageRepository,
	logger *slog.Logger,
) *Server {
	s := &Server{
		app:           fiber.New(),
		messaging:     messaging,
		phoneReg:      phoneReg,
		identities:    identities,
		conversations: conversations,
		messages:      messages,
		logger:         logger,
		sseClients:    make(map[string][]chan string),
	}
	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	// Webhook (no auth — Meta's server-to-server push)
	s.app.Get("/webhook", s.handleWebhookVerify)
	s.app.Post("/webhook", s.handleWebhookReceive)

	// Identity management
	s.app.Post("/api/identities", s.handleCreateIdentity)
	s.app.Get("/api/identities", s.handleListIdentities)

	// Virtual device routes
	s.app.Get("/device/:token", s.handleDevicePage)
	s.app.Get("/device/:token/identity", s.handleGetIdentity)
	s.app.Get("/device/:token/conversations", s.handleGetConversations)
	s.app.Get("/device/:token/conversations/:conv_id/messages", s.handleGetMessages)
	s.app.Post("/device/:token/conversations/:conv_id/messages", s.handleSendDeviceMessage)
	s.app.Post("/device/:token/conversations/:conv_id/read", s.handleMarkRead)
	s.app.Get("/device/:token/events", s.handleSSE)

	// Meta-compatible API — protected by bearer token auth
	meta := s.app.Group("/v18.0")
	meta.Use(s.authMiddleware())
	meta.Post("/:phone_number_id/messages", s.handleSendMessage)

	meta2 := s.app.Group("/:api_version")
	meta2.Use(s.authMiddleware())
	meta2.Post("/:phone_number_id/messages", s.handleSendMessage)

	// Media API (protected)
	media := s.app.Group("/media")
	media.Use(s.authMiddleware())
	media.Use(func(c *fiber.Ctx) error {
		c.Locals("media_store", s.mediaStore)
		return c.Next()
	})
	media.Post("/", s.handleUploadMedia)
	media.Get("/:media_id", s.handleGetMedia)
}

func (s *Server) SetAdminService(admin *application.AdminService) {
	s.admin = admin
}

func (s *Server) SetMediaStore(store ports.MediaStore) {
	s.mediaStore = store
}

func (s *Server) authMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if s.admin == nil {
			return c.Next()
		}

		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(http.StatusUnauthorized).JSON(domain.ErrMissingToken())
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == authHeader {
			return c.Status(http.StatusUnauthorized).JSON(domain.ErrMissingToken())
		}

		phoneNumberID := c.Params("phone_number_id")
		if phoneNumberID == "" {
			phoneNumberID = c.Query("phone_number_id")
		}
		if phoneNumberID == "" {
			phoneNumberID = c.Get("X-Phone-Number-Id")
		}

		// For Meta API, extract phone_number_id from URL path
		if phoneNumberID == "" {
			path := c.Path()
			parts := strings.Split(path, "/")
			for i, part := range parts {
				if part == "messages" && i > 0 {
					phoneNumberID = parts[i-1]
					break
				}
			}
		}

		// Normalize phone number if it looks like one
		if len(phoneNumberID) > 0 && phoneNumberID[0] >= '0' && phoneNumberID[0] <= '9' {
			phoneNumberID = string(domain.NormalizePhoneNumber(phoneNumberID))
		}

		result, err := s.admin.AuthorizeToken(c.Context(), token, phoneNumberID)
		if err != nil {
			s.logger.Error("auth failed", "error", err, "token_len", len(token), "phone_id", phoneNumberID)
			return c.Status(http.StatusUnauthorized).JSON(domain.ErrInvalidToken())
		}

		c.Locals("application", result.Application)
		c.Locals("waba", result.WABA)
		c.Locals("phone_number", result.PhoneNumber)
		c.Locals("webhook_config", result.WebhookConfig)

		return c.Next()
	}
}

func (s *Server) handleSendMessage(c *fiber.Ctx) error {
	phoneNumberID := c.Params("phone_number_id")

	var req struct {
		MessagingProduct string `json:"messaging_product"`
		To               string `json:"to"`
		Type             string `json:"type"`
		RecipientType    string `json:"recipient_type"`
		Text             *struct {
			PreviewURL bool   `json:"preview_url"`
			Body       string `json:"body"`
		} `json:"text,omitempty"`
		Image *struct {
			ID     string `json:"id"`
			Caption string `json:"caption,omitempty"`
		} `json:"image,omitempty"`
		Video *struct {
			ID     string `json:"id"`
			Caption string `json:"caption,omitempty"`
		} `json:"video,omitempty"`
		Audio *struct {
			ID string `json:"id"`
		} `json:"audio,omitempty"`
		Document *struct {
			ID       string `json:"id"`
			Caption  string `json:"caption,omitempty"`
			Filename string `json:"filename,omitempty"`
		} `json:"document,omitempty"`
		Location *struct {
			Latitude  float64 `json:"latitude"`
			Longitude float64 `json:"longitude"`
			Name      string  `json:"name,omitempty"`
			Address   string  `json:"address,omitempty"`
		} `json:"location,omitempty"`
		Contacts *struct {
			Contacts []struct {
				Name *struct {
					FirstName     string `json:"first_name,omitempty"`
					LastName      string `json:"last_name,omitempty"`
					FormattedName string `json:"formatted_name,omitempty"`
				} `json:"name,omitempty"`
				Phones []struct {
					Phone string `json:"phone"`
					Type  string `json:"type,omitempty"`
					WAID  string `json:"wa_id,omitempty"`
				} `json:"phones,omitempty"`
			} `json:"contacts"`
		} `json:"contacts,omitempty"`
		Template *struct {
			Name     string `json:"name"`
			Language struct {
				Code string `json:"code"`
			} `json:"language"`
			Components []struct {
				Type       string `json:"type"`
				SubType    string `json:"sub_type,omitempty"`
				Index      int    `json:"index,omitempty"`
				Parameters []struct {
					Type string `json:"type"`
					Text string `json:"text,omitempty"`
				} `json:"parameters,omitempty"`
			} `json:"components,omitempty"`
		} `json:"template,omitempty"`
		Interactive *struct {
			Type   string `json:"type"`
			Header *struct {
				Type string `json:"type"`
				Text string `json:"text,omitempty"`
			} `json:"header,omitempty"`
			Body struct {
				Text string `json:"text"`
			} `json:"body"`
			Foot *struct {
				Text string `json:"text"`
			} `json:"foot,omitempty"`
			Action struct {
				Buttons []struct {
					Type string `json:"type"`
					Text string `json:"text,omitempty"`
					ID   string `json:"id,omitempty"`
				} `json:"buttons,omitempty"`
				Sections []struct {
					Title string `json:"title,omitempty"`
					Rows  []struct {
						ID          string `json:"id"`
						Title       string `json:"title"`
						Description string `json:"description,omitempty"`
					} `json:"rows,omitempty"`
				} `json:"sections,omitempty"`
				Button string `json:"button,omitempty"`
			} `json:"action"`
		} `json:"interactive,omitempty"`
		Reaction *struct {
			MessageID string `json:"message_id"`
			Emoji     string `json:"emoji"`
		} `json:"reaction,omitempty"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(
			domain.NewMetaErrorResponse(100, "OAuthException", "Invalid parameter", "Could not parse request body"),
		)
	}

	if req.MessagingProduct != "whatsapp" {
		return c.Status(http.StatusBadRequest).JSON(
			domain.NewMetaErrorResponse(100, "OAuthException", "Invalid parameter", "messaging_product must be 'whatsapp'"),
		)
	}

	if req.To == "" {
		return c.Status(http.StatusBadRequest).JSON(
			domain.NewMetaErrorResponse(100, "OAuthException", "Invalid parameter", "to field is required"),
		)
	}

	msgType := domain.MessageType(req.Type)
	sendReq := &application.SendMessageRequest{
		From:          phoneNumberID,
		To:            req.To,
		Type:          msgType,
		PhoneNumberID: phoneNumberID,
	}

	switch msgType {
	case domain.MessageTypeText:
		if req.Text == nil || req.Text.Body == "" {
			return c.Status(http.StatusBadRequest).JSON(
				domain.NewMetaErrorResponse(100, "OAuthException", "Invalid parameter", "text.body is required"),
			)
		}
		sendReq.Body = req.Text.Body
		sendReq.PreviewURL = req.Text.PreviewURL

	case domain.MessageTypeImage:
		if req.Image == nil || req.Image.ID == "" {
			return c.Status(http.StatusBadRequest).JSON(
				domain.NewMetaErrorResponse(100, "OAuthException", "Invalid parameter", "image.id is required"),
			)
		}
		sendReq.MediaID = req.Image.ID
		sendReq.Caption = req.Image.Caption

	case domain.MessageTypeVideo:
		if req.Video == nil || req.Video.ID == "" {
			return c.Status(http.StatusBadRequest).JSON(
				domain.NewMetaErrorResponse(100, "OAuthException", "Invalid parameter", "video.id is required"),
			)
		}
		sendReq.MediaID = req.Video.ID
		sendReq.Caption = req.Video.Caption

	case domain.MessageTypeAudio:
		if req.Audio == nil || req.Audio.ID == "" {
			return c.Status(http.StatusBadRequest).JSON(
				domain.NewMetaErrorResponse(100, "OAuthException", "Invalid parameter", "audio.id is required"),
			)
		}
		sendReq.MediaID = req.Audio.ID

	case domain.MessageTypeDocument:
		if req.Document == nil || req.Document.ID == "" {
			return c.Status(http.StatusBadRequest).JSON(
				domain.NewMetaErrorResponse(100, "OAuthException", "Invalid parameter", "document.id is required"),
			)
		}
		sendReq.MediaID = req.Document.ID
		sendReq.Caption = req.Document.Caption
		sendReq.Filename = req.Document.Filename

	case domain.MessageTypeLocation:
		if req.Location == nil {
			return c.Status(http.StatusBadRequest).JSON(
				domain.NewMetaErrorResponse(100, "OAuthException", "Invalid parameter", "location is required"),
			)
		}
		sendReq.Location = &application.LocationRequest{
			Latitude:  req.Location.Latitude,
			Longitude: req.Location.Longitude,
			Name:      req.Location.Name,
			Address:   req.Location.Address,
		}

	case domain.MessageTypeContacts:
		if req.Contacts == nil || len(req.Contacts.Contacts) == 0 {
			return c.Status(http.StatusBadRequest).JSON(
				domain.NewMetaErrorResponse(100, "OAuthException", "Invalid parameter", "contacts is required"),
			)
		}
		for _, c := range req.Contacts.Contacts {
			cr := application.ContactRequest{}
			if c.Name != nil {
				cr.Name = &application.ContactNameRequest{
					FirstName:     c.Name.FirstName,
					LastName:      c.Name.LastName,
					FormattedName: c.Name.FormattedName,
				}
			}
			for _, p := range c.Phones {
				cr.Phones = append(cr.Phones, application.ContactPhoneRequest{
					Phone: p.Phone,
					Type:  p.Type,
					WAID:  p.WAID,
				})
			}
			sendReq.Contacts = append(sendReq.Contacts, cr)
		}

	case domain.MessageTypeTemplate:
		if req.Template == nil {
			return c.Status(http.StatusBadRequest).JSON(
				domain.NewMetaErrorResponse(100, "OAuthException", "Invalid parameter", "template is required"),
			)
		}
		sendReq.Template = &application.TemplateRequest{
			Name:     req.Template.Name,
			Language: req.Template.Language.Code,
		}
		for _, comp := range req.Template.Components {
			cr := application.TemplateComponentRequest{
				Type:    comp.Type,
				SubType: comp.SubType,
				Index:   comp.Index,
			}
			for _, p := range comp.Parameters {
				cr.Parameters = append(cr.Parameters, application.TemplateParameterRequest{
					Type: p.Type,
					Text: p.Text,
				})
			}
			sendReq.Template.Components = append(sendReq.Template.Components, cr)
		}

	case domain.MessageTypeInteractive:
		if req.Interactive == nil {
			return c.Status(http.StatusBadRequest).JSON(
				domain.NewMetaErrorResponse(100, "OAuthException", "Invalid parameter", "interactive is required"),
			)
		}
		sendReq.Interactive = &application.InteractiveRequest{
			Type: req.Interactive.Type,
			Body: req.Interactive.Body.Text,
		}
		if req.Interactive.Header != nil {
			sendReq.Interactive.Header = &application.InteractiveHeaderRequest{
				Type: req.Interactive.Header.Type,
				Text: req.Interactive.Header.Text,
			}
		}
		if req.Interactive.Foot != nil {
			sendReq.Interactive.Foot = req.Interactive.Foot.Text
		}
		for _, btn := range req.Interactive.Action.Buttons {
			sendReq.Interactive.Buttons = append(sendReq.Interactive.Buttons, application.InteractiveButtonRequest{
				Type: btn.Type,
				Text: btn.Text,
				ID:   btn.ID,
			})
		}
		for _, sec := range req.Interactive.Action.Sections {
			sr := application.InteractiveSectionRequest{Title: sec.Title}
			for _, row := range sec.Rows {
				sr.Rows = append(sr.Rows, application.InteractiveRowRequest{
					ID:          row.ID,
					Title:       row.Title,
					Description: row.Description,
				})
			}
			sendReq.Interactive.Sections = append(sendReq.Interactive.Sections, sr)
		}
		sendReq.Interactive.Button = req.Interactive.Action.Button

	case domain.MessageTypeReaction:
		if req.Reaction == nil {
			return c.Status(http.StatusBadRequest).JSON(
				domain.NewMetaErrorResponse(100, "OAuthException", "Invalid parameter", "reaction is required"),
			)
		}
		sendReq.ReactionMessageID = req.Reaction.MessageID
		sendReq.ReactionEmoji = req.Reaction.Emoji

	default:
		return c.Status(http.StatusBadRequest).JSON(domain.ErrUnsupportedMessageType())
	}

	resp, err := s.messaging.SendMessage(c.Context(), sendReq)
	if err != nil {
		s.logger.Error("failed to send message", "error", err)
		return c.Status(http.StatusInternalServerError).JSON(
			domain.NewMetaErrorResponse(100, "OAuthException", "Internal error", err.Error()),
		)
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"messaging_product": "whatsapp",
		"contacts": []fiber.Map{
			{"input": req.To, "wa_id": strings.TrimPrefix(req.To, "+")},
		},
		"messages": []fiber.Map{
			{"id": resp.WAMessageID},
		},
	})
}

func (s *Server) handleWebhookVerify(c *fiber.Ctx) error {
	mode := c.Query("hub.mode")
	token := c.Query("hub.verify_token")
	challenge := c.Query("hub.challenge")

	if mode != "subscribe" || token != "wakaba-verify-token" {
		return c.Status(http.StatusForbidden).SendString("Forbidden")
	}

	s.logger.Info("webhook verified", "challenge", challenge)
	return c.Status(http.StatusOK).SendString(challenge)
}

func (s *Server) handleWebhookReceive(c *fiber.Ctx) error {
	body := c.Body()

	var payload struct {
		Object string `json:"object"`
		Entry  []struct {
			ID      string `json:"id"`
			Changes []struct {
				Value struct {
					Metadata struct {
						PhoneNumberID string `json:"phone_number_id"`
					} `json:"metadata"`
					Messages []struct {
						From      string `json:"from"`
						ID        string `json:"id"`
						Timestamp string `json:"timestamp"`
						Type      string `json:"type"`
						Text      *struct {
							Body string `json:"body"`
						} `json:"text,omitempty"`
						Image *struct {
							ID     string `json:"id"`
							MimeType string `json:"mime_type,omitempty"`
							Caption  string `json:"caption,omitempty"`
						} `json:"image,omitempty"`
						Video *struct {
							ID     string `json:"id"`
							MimeType string `json:"mime_type,omitempty"`
							Caption  string `json:"caption,omitempty"`
						} `json:"video,omitempty"`
						Audio *struct {
							ID     string `json:"id"`
							MimeType string `json:"mime_type,omitempty"`
						} `json:"audio,omitempty"`
						Document *struct {
							ID       string `json:"id"`
							MimeType string `json:"mime_type,omitempty"`
							Caption  string `json:"caption,omitempty"`
							Filename string `json:"filename,omitempty"`
						} `json:"document,omitempty"`
						Location *struct {
							Latitude  float64 `json:"latitude"`
							Longitude float64 `json:"longitude"`
							Name      string  `json:"name,omitempty"`
							Address   string  `json:"address,omitempty"`
						} `json:"location,omitempty"`
						Contacts *struct {
							Contacts []struct {
								Name *struct {
									FirstName     string `json:"first_name,omitempty"`
									LastName      string `json:"last_name,omitempty"`
									FormattedName string `json:"formatted_name,omitempty"`
								} `json:"name,omitempty"`
								Phones []struct {
									Phone string `json:"phone"`
									Type  string `json:"type,omitempty"`
									WAID  string `json:"wa_id,omitempty"`
								} `json:"phones,omitempty"`
							} `json:"contacts"`
						} `json:"contacts,omitempty"`
						Interactive *struct {
							Type   string `json:"type"`
							ButtonID string `json:"button,omitempty"`
							 ListReply *struct {
								ID          string `json:"id"`
								Title       string `json:"title"`
								Description string `json:"description,omitempty"`
							} `json:"list_reply,omitempty"`
							ButtonReply *struct {
								ID   string `json:"id"`
								Text string `json:"title"`
							} `json:"button_reply,omitempty"`
						} `json:"interactive,omitempty"`
						Reaction *struct {
							MessageID string `json:"message_id"`
							Emoji     string `json:"emoji"`
						} `json:"reaction,omitempty"`
					} `json:"messages,omitempty"`
					Statuses []struct {
						ID          string `json:"id"`
						Status      string `json:"status"`
						Timestamp   string `json:"timestamp"`
						RecipientID string `json:"recipient_id"`
					} `json:"statuses,omitempty"`
				} `json:"value"`
				Field string `json:"field"`
			} `json:"changes"`
		} `json:"entry"`
	}

	if err := json.Unmarshal(body, &payload); err != nil {
		return c.Status(http.StatusBadRequest).SendString("Invalid JSON")
	}

	if payload.Object != "whatsapp_business_account" {
		return c.Status(http.StatusOK).SendString("OK")
	}

	for _, entry := range payload.Entry {
		for _, change := range entry.Changes {
			phoneID := change.Value.Metadata.PhoneNumberID
			for _, msg := range change.Value.Messages {
				s.logger.Info("inbound message", "from", msg.From, "to", phoneID, "type", msg.Type)

				recvReq := &application.ReceiveMessageRequest{
					From:          msg.From,
					To:            phoneID,
					Type:          domain.MessageType(msg.Type),
					PhoneNumberID: phoneID,
					WAMessageID:   msg.ID,
				}

				switch msg.Type {
				case "text":
					if msg.Text != nil {
						recvReq.Body = msg.Text.Body
					}
				case "image":
					if msg.Image != nil {
						recvReq.MediaID = msg.Image.ID
						recvReq.Caption = msg.Image.Caption
						recvReq.MediaMimeType = msg.Image.MimeType
					}
				case "video":
					if msg.Video != nil {
						recvReq.MediaID = msg.Video.ID
						recvReq.Caption = msg.Video.Caption
						recvReq.MediaMimeType = msg.Video.MimeType
					}
				case "audio":
					if msg.Audio != nil {
						recvReq.MediaID = msg.Audio.ID
						recvReq.MediaMimeType = msg.Audio.MimeType
					}
				case "document":
					if msg.Document != nil {
						recvReq.MediaID = msg.Document.ID
						recvReq.Caption = msg.Document.Caption
						recvReq.Filename = msg.Document.Filename
						recvReq.MediaMimeType = msg.Document.MimeType
					}
				case "location":
					if msg.Location != nil {
						recvReq.Location = &application.LocationRequest{
							Latitude:  msg.Location.Latitude,
							Longitude: msg.Location.Longitude,
							Name:      msg.Location.Name,
							Address:   msg.Location.Address,
						}
					}
				case "contacts":
					if msg.Contacts != nil {
						for _, c := range msg.Contacts.Contacts {
							cr := application.ContactRequest{}
							if c.Name != nil {
								cr.Name = &application.ContactNameRequest{
									FirstName:     c.Name.FirstName,
									LastName:      c.Name.LastName,
									FormattedName: c.Name.FormattedName,
								}
							}
							for _, p := range c.Phones {
								cr.Phones = append(cr.Phones, application.ContactPhoneRequest{
									Phone: p.Phone, Type: p.Type, WAID: p.WAID,
								})
							}
							recvReq.Contacts = append(recvReq.Contacts, cr)
						}
					}
				case "interactive":
					if msg.Interactive != nil {
						recvReq.InteractiveType = msg.Interactive.Type
						if msg.Interactive.ButtonID != "" {
							recvReq.InteractiveButtonID = msg.Interactive.ButtonID
						}
						if msg.Interactive.ListReply != nil {
							recvReq.InteractiveButtonID = msg.Interactive.ListReply.ID
						}
						if msg.Interactive.ButtonReply != nil {
							recvReq.InteractiveButtonID = msg.Interactive.ButtonReply.ID
						}
					}
				case "reaction":
					if msg.Reaction != nil {
						recvReq.ReactionMessageID = msg.Reaction.MessageID
						recvReq.ReactionEmoji = msg.Reaction.Emoji
					}
				}

				resp, err := s.messaging.ReceiveMessage(c.Context(), recvReq)
				if err != nil {
					s.logger.Error("failed to process inbound", "error", err)
				} else {
					s.notifySSE(msg.From, map[string]interface{}{
						"type":           "new_message",
						"message_id":     resp.MessageID,
						"conversation_id": "",
					})
				}
			}
			for _, status := range change.Value.Statuses {
				s.logger.Info("status update", "message_id", status.ID, "status", status.Status)
				var msgStatus domain.MessageStatus
				switch status.Status {
				case "sent":
					msgStatus = domain.MessageStatusSent
				case "delivered":
					msgStatus = domain.MessageStatusDelivered
				case "read":
					msgStatus = domain.MessageStatusRead
				case "failed":
					msgStatus = domain.MessageStatusFailed
				default:
					continue
				}
				_ = s.messaging.UpdateStatus(c.Context(), domain.MessageID(status.ID), msgStatus)
			}
		}
	}

	return c.Status(http.StatusOK).SendString("OK")
}

// --- Virtual Device Routes ---

func (s *Server) handleDevicePage(c *fiber.Ctx) error {
	token := c.Params("token")
	identity, err := s.getIdentityByToken(c.Context(), token)
	if err != nil || identity == nil {
		return c.Status(http.StatusNotFound).SendString("Identity not found")
	}

	return c.Type("html").SendString(fmt.Sprintf(`<!DOCTYPE html>
<html><head><title>Wakaba - %s</title>
<script src="https://cdn.tailwindcss.com"></script>
<script defer src="https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js"></script>
<style>[x-cloak]{display:none!important}</style>
</head>
<body class="bg-[#efeae2] h-screen flex flex-col" x-data="chat()" x-init="init()">
<div class="bg-[#008069] text-white px-4 py-3 flex items-center gap-3">
  <div class="w-10 h-10 rounded-full bg-[#d9fdd3] flex items-center justify-center text-[#008069] font-bold" x-text="identity?.display_name?.charAt(0)||'?'"></div>
  <div><div class="font-semibold" x-text="identity?.display_name"></div>
  <div class="text-xs text-green-100" x-text="identity?.phone"></div></div>
  <div class="ml-auto text-xs text-green-100">Wakaba Virtual Network</div>
</div>
<div class="flex-1 flex overflow-hidden">
  <div class="w-80 bg-white border-r border-gray-200 flex flex-col">
    <div class="p-3 border-b"><input type="text" placeholder="Search or start new chat" class="w-full px-3 py-2 bg-gray-100 rounded-lg text-sm focus:outline-none" x-model="search"></div>
    <div class="flex-1 overflow-y-auto">
      <template x-for="c in filtered" :key="c.id">
        <div @click="select(c)" class="flex items-center gap-3 px-4 py-3 cursor-pointer hover:bg-gray-50 border-b" :class="{'bg-gray-100':sel?.id===c.id}">
          <div class="w-12 h-12 rounded-full bg-gray-200 flex items-center justify-center font-semibold" x-text="c.name?.charAt(0)||'?'"></div>
          <div class="flex-1 min-w-0"><div class="flex justify-between"><span class="font-medium text-sm truncate" x-text="c.name||c.phone"></span>
          <span class="text-xs text-gray-400" x-text="fmt(c.last_message_at)"></span></div>
          <div class="flex items-center gap-1"><span class="text-sm text-gray-500 truncate" x-text="c.last_message||''"></span>
          <span x-show="c.unread_count>0" class="ml-auto bg-[#25d366] text-white text-xs rounded-full w-5 h-5 flex items-center justify-center" x-text="c.unread_count"></span></div></div>
        </div>
      </template>
      <div x-show="filtered.length===0" class="p-6 text-center text-gray-400 text-sm">No conversations yet.</div>
    </div>
  </div>
  <div class="flex-1 flex flex-col" x-show="sel">
    <div class="bg-gray-100 px-4 py-2 flex items-center gap-3 border-b">
      <div class="w-10 h-10 rounded-full bg-gray-200 flex items-center justify-center font-semibold" x-text="sel?.name?.charAt(0)||'?'"></div>
      <div><div class="font-medium text-sm" x-text="sel?.name||sel?.phone"></div><div class="text-xs text-gray-500" x-text="sel?.phone"></div></div>
    </div>
    <div class="flex-1 overflow-y-auto px-16 py-4 bg-[#efeae2]" x-ref="msgs">
      <template x-for="m in msgs" :key="m.id">
        <div class="flex mb-2" :class="m.direction==='outbound'?'justify-end':'justify-start'">
          <div class="max-w-[65%%] px-3 py-2 rounded-lg shadow-sm" :class="m.direction==='outbound'?'bg-[#d9fdd3] rounded-tr-none':'bg-white rounded-tl-none'">
            <div class="text-sm" x-text="m.text?.body||''"></div>
            <div class="flex items-center gap-1 justify-end mt-1">
              <span class="text-[10px] text-gray-500" x-text="fmt(m.timestamp)"></span>
              <template x-if="m.direction==='outbound'">
                <span class="text-[10px]" :class="m.status==='read'?'text-[#53bdeb]':'text-gray-400'" x-text="m.status==='read'||m.status==='delivered'?'✓✓':m.status==='sent'?'✓':m.status==='failed'?'⚠':'🕐'"></span>
              </template>
            </div>
          </div>
        </div>
      </template>
    </div>
    <div class="bg-gray-100 px-4 py-3 flex items-center gap-2">
      <input type="text" x-model="body" @keydown.enter="send()" placeholder="Type a message" class="flex-1 px-4 py-2 bg-white rounded-lg text-sm focus:outline-none">
      <button @click="send()" class="w-10 h-10 bg-[#008069] text-white rounded-full flex items-center justify-center hover:bg-[#006d5b]">➤</button>
    </div>
  </div>
  <div class="flex-1 flex items-center justify-center bg-gray-50" x-show="!sel">
    <div class="text-center text-gray-400"><p class="text-sm">Select a conversation</p></div>
  </div>
</div>
<script>
function chat(){return{
  identity:null,convs:[],sel:null,msgs:[],body:'',search:'',
  get filtered(){if(!this.search)return this.convs;const q=this.search.toLowerCase();return this.convs.filter(c=>(c.name||'').toLowerCase().includes(q)||(c.phone||'').includes(q))},
  async init(){const r=await fetch('/device/%s/identity');this.identity=await r.json();await this.loadConvs();this.connectSSE()},
  async loadConvs(){const r=await fetch('/device/%s/conversations');this.convs=await r.json()},
  async select(c){this.sel=c;const r=await fetch('/device/%s/conversations/'+c.id+'/messages');this.msgs=await r.json();await fetch('/device/%s/conversations/'+c.id+'/read',{method:'POST'});c.unread_count=0;this.$nextTick(()=>{const el=this.$refs.msgs;if(el)el.scrollTop=el.scrollHeight})},
  async send(){if(!this.body.trim()||!this.sel)return;const b=this.body.trim();this.body='';await fetch('/device/%s/conversations/'+this.sel.id+'/messages',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({body:b})});await this.loadConvs();if(this.sel){const r=await fetch('/device/%s/conversations/'+this.sel.id+'/messages');this.msgs=await r.json();this.$nextTick(()=>{const el=this.$refs.msgs;if(el)el.scrollTop=el.scrollHeight})}},
  connectSSE(){const es=new EventSource('/device/%s/events');es.onmessage=async(e)=>{const d=JSON.parse(e.data);if(d.type==='new_message'){await this.loadConvs();if(this.sel){const r=await fetch('/device/%s/conversations/'+this.sel.id+'/messages');this.msgs=await r.json();this.$nextTick(()=>{const el=this.$refs.msgs;if(el)el.scrollTop=el.scrollHeight})}}}},
  fmt(t){if(!t)return'';return new Date(t).toLocaleTimeString([],{hour:'2-digit',minute:'2-digit'})}
}}</script>
</body></html>`,
		identity.DisplayName,
		token, token, token, token, token, token, token, token))
}

func (s *Server) handleGetIdentity(c *fiber.Ctx) error {
	token := c.Params("token")
	identity, err := s.getIdentityByToken(c.Context(), token)
	if err != nil || identity == nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "not found"})
	}
	return c.JSON(fiber.Map{
		"id":          identity.ID,
		"phone":       identity.Phone,
		"display_name": identity.DisplayName,
	})
}

func (s *Server) handleGetConversations(c *fiber.Ctx) error {
	token := c.Params("token")
	identity, err := s.getIdentityByToken(c.Context(), token)
	if err != nil || identity == nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "not found"})
	}

	convs, err := s.conversations.GetConversationsByPhone(c.Context(), identity.Phone)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	type convResp struct {
		ID           string  `json:"id"`
		Phone        string  `json:"phone"`
		Name         string  `json:"name"`
		LastMessage  string  `json:"last_message"`
		LastMessageAt *string `json:"last_message_at"`
		UnreadCount  int     `json:"unread_count"`
	}

	result := make([]convResp, len(convs))
	for i, conv := range convs {
		name := ""
		ownerType, ownerID, _ := s.phoneReg.GetOwner(c.Context(), conv.PhoneNumber)
		if ownerType == "identity" {
			vID, _ := domain.VirtualIdentityID(ownerID), ""
			ident, _ := s.identities.GetIdentityByID(c.Context(), vID)
			if ident != nil {
				name = ident.DisplayName
			}
		} else if ownerType == "business_phone" {
			name = "Business"
		}

		var lastMsg *string
		if conv.LastMessageAt != nil {
			s := conv.LastMessageAt.Format("2006-01-02T15:04:05Z")
			lastMsg = &s
		}

		result[i] = convResp{
			ID:            string(conv.ID),
			Phone:         string(conv.PhoneNumber),
			Name:          name,
			LastMessageAt: lastMsg,
			UnreadCount:   conv.UnreadCount,
		}
	}

	return c.JSON(result)
}

func (s *Server) handleGetMessages(c *fiber.Ctx) error {
	convID := c.Params("conv_id")

	msgs, err := s.messages.GetMessagesByConversation(c.Context(), domain.ConversationID(convID), 50, nil)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	type msgResp struct {
		ID        string               `json:"id"`
		Body      string               `json:"body"`
		Direction domain.MessageDirection `json:"direction"`
		Status    domain.MessageStatus `json:"status"`
		Timestamp string               `json:"timestamp"`
	}

	result := make([]msgResp, len(msgs))
	for i, msg := range msgs {
		body := ""
		if msg.Text != nil {
			body = msg.Text.Body
		}
		result[i] = msgResp{
			ID:        string(msg.ID),
			Body:      body,
			Direction: msg.Direction,
			Status:    msg.Status,
			Timestamp: msg.Timestamp.Format("2006-01-02T15:04:05Z"),
		}
	}

	return c.JSON(result)
}

func (s *Server) handleSendDeviceMessage(c *fiber.Ctx) error {
	token := c.Params("token")
	convID := c.Params("conv_id")

	identity, err := s.getIdentityByToken(c.Context(), token)
	if err != nil || identity == nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "not found"})
	}

	var req struct {
		Body string `json:"body"`
	}
	if err := c.BodyParser(&req); err != nil || req.Body == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "body required"})
	}

	conv, err := s.conversations.GetConversationByID(c.Context(), domain.ConversationID(convID))
	if err != nil || conv == nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "conversation not found"})
	}

	_, err = s.messaging.ReceiveMessage(c.Context(), &application.ReceiveMessageRequest{
		From:          string(identity.Phone),
		To:            conv.BusinessPhoneID,
		Type:          domain.MessageTypeText,
		Body:          req.Body,
		PhoneNumberID: conv.BusinessPhoneID,
	})
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"ok": true})
}

func (s *Server) handleMarkRead(c *fiber.Ctx) error {
	convID := c.Params("conv_id")
	_ = s.conversations.ResetUnread(c.Context(), domain.ConversationID(convID))
	return c.JSON(fiber.Map{"ok": true})
}

func (s *Server) handleCreateIdentity(c *fiber.Ctx) error {
	var req struct {
		Phone       string `json:"phone"`
		DisplayName string `json:"display_name"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}

	phone := domain.NormalizePhoneNumber(req.Phone)

	available, err := s.phoneReg.IsAvailable(c.Context(), phone)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	if !available {
		return c.Status(http.StatusConflict).JSON(fiber.Map{"error": "phone number already registered"})
	}

	identity := domain.NewVirtualIdentity(phone, req.DisplayName)
	if err := s.identities.CreateIdentity(c.Context(), identity); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	if err := s.phoneReg.Reserve(c.Context(), phone, "identity", string(identity.ID)); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(http.StatusCreated).JSON(fiber.Map{
		"id":          identity.ID,
		"phone":       identity.Phone,
		"display_name": identity.DisplayName,
		"device_token": string(identity.ID),
	})
}

func (s *Server) handleListIdentities(c *fiber.Ctx) error {
	ids, err := s.identities.ListIdentities(c.Context())
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	type idResp struct {
		ID          string `json:"id"`
		Phone       string `json:"phone"`
		DisplayName string `json:"display_name"`
	}
	result := make([]idResp, len(ids))
	for i, id := range ids {
		result[i] = idResp{ID: string(id.ID), Phone: string(id.Phone), DisplayName: id.DisplayName}
	}
	return c.JSON(result)
}

func (s *Server) getIdentityByToken(ctx context.Context, token string) (*domain.VirtualIdentity, error) {
	return s.identities.GetIdentityByID(ctx, domain.VirtualIdentityID(token))
}

func (s *Server) notifySSE(phone string, data interface{}) {
	s.sseMu.RLock()
	defer s.sseMu.RUnlock()
	payload, _ := json.Marshal(data)
	for _, ch := range s.sseClients[phone] {
		select {
		case ch <- string(payload):
		default:
		}
	}
}

func (s *Server) handleSSE(c *fiber.Ctx) error {
	token := c.Params("token")
	identity, err := s.getIdentityByToken(c.Context(), token)
	if err != nil || identity == nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "not found"})
	}

	ch := make(chan string, 10)
	phone := string(identity.Phone)

	s.sseMu.Lock()
	s.sseClients[phone] = append(s.sseClients[phone], ch)
	s.sseMu.Unlock()

	defer func() {
		s.sseMu.Lock()
		clients := s.sseClients[phone]
		for i, client := range clients {
			if client == ch {
				s.sseClients[phone] = append(clients[:i], clients[i+1:]...)
				break
			}
		}
		s.sseMu.Unlock()
	}()

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("X-Accel-Buffering", "no")

	for {
		select {
		case <-c.Context().Done():
			return nil
		case msg := <-ch:
			if _, err := fmt.Fprintf(c.Context().Response.BodyWriter(), "data: %s\n\n", msg); err != nil {
				return nil
			}
		}
	}
}

func (s *Server) RegisterAdmin(admin *application.AdminService, messaging *application.MessagingService) {
	adminHandler := NewAdminHandler(admin, messaging)
	adminHandler.Register(s.app)
}

func (s *Server) handleUploadMedia(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "file is required"})
	}

	f, err := file.Open()
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed to open file"})
	}
	defer f.Close()

	data := make([]byte, file.Size)
	n, err := f.Read(data)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed to read file"})
	}
	data = data[:n]

	mediaStore := c.Locals("media_store").(ports.MediaStore)
	if mediaStore == nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "media store not available"})
	}

	mediaID, err := mediaStore.Store(c.Context(), file.Filename, data, file.Header.Get("Content-Type"))
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"id":          mediaID,
		"file_name":   file.Filename,
		"file_size":   n,
		"mime_type":   file.Header.Get("Content-Type"),
	})
}

func (s *Server) handleGetMedia(c *fiber.Ctx) error {
	mediaID := c.Params("media_id")

	mediaStore := c.Locals("media_store").(ports.MediaStore)
	if mediaStore == nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "media store not available"})
	}

	data, mimeType, err := mediaStore.Get(c.Context(), mediaID)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "media not found"})
	}

	c.Set("Content-Type", mimeType)
	return c.Send(data)
}

func (s *Server) Listen(addr string) error {
	return s.app.Listen(addr)
}

func (s *Server) Shutdown() error {
	return s.app.Shutdown()
}

func (s *Server) App() *fiber.App {
	return s.app
}
