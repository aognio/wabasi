package domain

import "time"

type MessageType string

const (
	MessageTypeText       MessageType = "text"
	MessageTypeImage      MessageType = "image"
	MessageTypeVideo      MessageType = "video"
	MessageTypeAudio      MessageType = "audio"
	MessageTypeDocument   MessageType = "document"
	MessageTypeLocation   MessageType = "location"
	MessageTypeContacts   MessageType = "contacts"
	MessageTypeTemplate   MessageType = "template"
	MessageTypeInteractive MessageType = "interactive"
	MessageTypeReaction   MessageType = "reaction"
)

type MessageStatus string

const (
	MessageStatusPending     MessageStatus = "pending"
	MessageStatusAccepted    MessageStatus = "accepted"
	MessageStatusSent        MessageStatus = "sent"
	MessageStatusDelivered   MessageStatus = "delivered"
	MessageStatusRead        MessageStatus = "read"
	MessageStatusFailed      MessageStatus = "failed"
)

type MessageDirection string

const (
	MessageDirectionInbound  MessageDirection = "inbound"
	MessageDirectionOutbound MessageDirection = "outbound"
)

type TextContent struct {
	Body       string `json:"body"`
	PreviewURL bool   `json:"preview_url,omitempty"`
}

type MediaContent struct {
	ID       string `json:"id"`
	MimeType string `json:"mime_type,omitempty"`
	Caption  string `json:"caption,omitempty"`
	Filename string `json:"filename,omitempty"`
}

type LocationContent struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Name      string  `json:"name,omitempty"`
	Address   string  `json:"address,omitempty"`
}

type ContactContent struct {
	Name       *ContactName       `json:"name,omitempty"`
	Phones     []ContactPhone     `json:"phones,omitempty"`
	Emails     []ContactEmail     `json:"emails,omitempty"`
	Addresses  []ContactAddress   `json:"addresses,omitempty"`
	Org        *ContactOrg        `json:"org,omitempty"`
	Birthday   string             `json:"birthday,omitempty"`
	Urls       []ContactURL       `json:"urls,omitempty"`
}

type ContactName struct {
	FirstName         string `json:"first_name,omitempty"`
	LastName          string `json:"last_name,omitempty"`
	MiddleName        string `json:"middle_name,omitempty"`
	NameSuffix        string `json:"name_suffix,omitempty"`
	NamePrefix        string `json:"name_prefix,omitempty"`
	FormattedName     string `json:"formatted_name,omitempty"`
}

type ContactPhone struct {
Phone string `json:"phone"`
Type  string `json:"type,omitempty"`
	WAID string `json:"wa_id,omitempty"`
}

type ContactEmail struct {
	Email string `json:"email"`
	Type  string `json:"type,omitempty"`
}

type ContactAddress struct {
	Street      string `json:"street,omitempty"`
	City        string `json:"city,omitempty"`
	State       string `json:"state,omitempty"`
	Zip         string `json:"zip,omitempty"`
	Country     string `json:"country,omitempty"`
	CountryCode string `json:"country_code,omitempty"`
	Type        string `json:"type,omitempty"`
}

type ContactOrg struct {
	Company    string `json:"company,omitempty"`
	Department string `json:"department,omitempty"`
	Title      string `json:"title,omitempty"`
}

type ContactURL struct {
	URL  string `json:"url"`
	Type string `json:"type,omitempty"`
}

type TemplateContent struct {
	Name       string            `json:"name"`
	Language   TemplateLanguage  `json:"language"`
	Components []TemplateComponent `json:"components,omitempty"`
}

type TemplateLanguage struct {
	Code string `json:"code"`
}

type TemplateComponent struct {
	Type       string              `json:"type"`
	SubType    string              `json:"sub_type,omitempty"`
	Index      int                 `json:"index,omitempty"`
	Parameters []TemplateParameter `json:"parameters,omitempty"`
}

type TemplateParameter struct {
	Type     string        `json:"type"`
	Text     string        `json:"text,omitempty"`
	Image    *MediaContent `json:"image,omitempty"`
	Video    *MediaContent `json:"video,omitempty"`
	Document *MediaContent `json:"document,omitempty"`
}

type InteractiveContent struct {
	Type   string              `json:"type"`
	Header *InteractiveHeader  `json:"header,omitempty"`
	Body   InteractiveBody     `json:"body"`
	Foot   *InteractiveFoot   `json:"foot,omitempty"`
	Action InteractiveAction  `json:"action"`
}

type InteractiveHeader struct {
	Type     string        `json:"type"`
	Text     string        `json:"text,omitempty"`
	Image    *MediaContent `json:"image,omitempty"`
	Video    *MediaContent `json:"video,omitempty"`
	Document *MediaContent `json:"document,omitempty"`
}

type InteractiveBody struct {
	Text string `json:"text"`
}

type InteractiveFoot struct {
	Text string `json:"text"`
}

type InteractiveAction struct {
	Buttons   []InteractiveButton   `json:"buttons,omitempty"`
	Sections  []InteractiveSection  `json:"sections,omitempty"`
	Button    string                `json:"button,omitempty"`
}

type InteractiveButton struct {
	Type  string `json:"type"`
	Text  string `json:"text,omitempty"`
	ID    string `json:"id,omitempty"`
}

type InteractiveSection struct {
	Title         string              `json:"title,omitempty"`
	Rows          []InteractiveRow    `json:"rows,omitempty"`
}

type InteractiveRow struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
}

type ReactionContent struct {
	MessageID string `json:"message_id"`
	Emoji     string `json:"emoji"`
}

type Message struct {
	ID             MessageID        `json:"id"`
	WAMessageID    string           `json:"wamessage_id,omitempty"`
	ConversationID string           `json:"conversation_id"`
	From           WhatsAppID       `json:"from"`
	To             WhatsAppID       `json:"to"`
	Type           MessageType      `json:"type"`
	Status         MessageStatus    `json:"status"`
	Direction      MessageDirection `json:"direction"`
	Text           *TextContent     `json:"text,omitempty"`
	Image          *MediaContent    `json:"image,omitempty"`
	Video          *MediaContent    `json:"video,omitempty"`
	Audio          *MediaContent    `json:"audio,omitempty"`
	Document       *MediaContent    `json:"document,omitempty"`
	Location       *LocationContent `json:"location,omitempty"`
	Contacts       []ContactContent `json:"contacts,omitempty"`
	Template       *TemplateContent `json:"template,omitempty"`
	Interactive    *InteractiveContent `json:"interactive,omitempty"`
	Reaction       *ReactionContent `json:"reaction,omitempty"`
	Timestamp      time.Time        `json:"timestamp"`
	SentAt         *time.Time       `json:"sent_at,omitempty"`
	DeliveredAt    *time.Time       `json:"delivered_at,omitempty"`
	ReadAt         *time.Time       `json:"read_at,omitempty"`
	MediaID        string           `json:"media_id,omitempty"`
	MediaMimeType  string           `json:"media_mime_type,omitempty"`
}

func NewInboundTextMessage(from, to WhatsAppID, body string, conversationID string) *Message {
	return &Message{
		ID:             GenerateMessageID(),
		ConversationID: conversationID,
		From:           from,
		To:             to,
		Type:           MessageTypeText,
		Status:         MessageStatusDelivered,
		Direction:      MessageDirectionInbound,
		Text:           &TextContent{Body: body},
		Timestamp:      time.Now(),
	}
}

func NewOutboundTextMessage(from, to WhatsAppID, body string, conversationID string) *Message {
	now := time.Now()
	return &Message{
		ID:             GenerateMessageID(),
		ConversationID: conversationID,
		From:           from,
		To:             to,
		Type:           MessageTypeText,
		Status:         MessageStatusPending,
		Direction:      MessageDirectionOutbound,
		Text:           &TextContent{Body: body},
		Timestamp:      now,
		SentAt:         &now,
	}
}
