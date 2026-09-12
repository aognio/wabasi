package domain

import "fmt"

type MetaError struct {
	Code    int              `json:"code"`
	Message string           `json:"message"`
	Type    string           `json:"type"`
	ErrorData *MetaErrorData `json:"error_data,omitempty"`
	FBTraceID string         `json:"fbtrace_id,omitempty"`
}

type MetaErrorData struct {
	MessagingProduct string `json:"messaging_product"`
	Details          string `json:"details"`
}

type MetaErrorResponse struct {
	Error MetaError `json:"error"`
}

func NewMetaErrorResponse(code int, errType, message, details string) *MetaErrorResponse {
	return &MetaErrorResponse{
		Error: MetaError{
			Code:    code,
			Message: fmt.Sprintf("(#%d) %s", code, message),
			Type:    errType,
			ErrorData: &MetaErrorData{
				MessagingProduct: "whatsapp",
				Details:          details,
			},
		},
	}
}

var (
	ErrInvalidToken = func() *MetaErrorResponse {
		return NewMetaErrorResponse(190, "OAuthException", "Invalid OAuth access token", "The access token is invalid or has expired.")
	}

	ErrInvalidParameter = func(details string) *MetaErrorResponse {
		return NewMetaErrorResponse(100, "OAuthException", "Invalid parameter", details)
	}

	ErrUnsupportedMessageType = func() *MetaErrorResponse {
		return NewMetaErrorResponse(131051, "WhatsAppBusinessApiError", "Unsupported message type", "This message type is not supported.")
	}

	ErrPhoneNumberNotFound = func() *MetaErrorResponse {
		return NewMetaErrorResponse(100, "OAuthException", "Invalid parameter", "Phone number does not exist or is not a valid WhatsApp number.")
	}

	ErrRateLimit = func() *MetaErrorResponse {
		return NewMetaErrorResponse(80007, "OAuthException", "Rate limit exceeded", "API rate limit has been reached.")
	}

	ErrUnknownPhoneID = func() *MetaErrorResponse {
		return NewMetaErrorResponse(100, "OAuthException", "Invalid parameter", "Unknown phone_number_id.")
	}

	ErrMissingToken = func() *MetaErrorResponse {
		return NewMetaErrorResponse(190, "OAuthException", "Invalid OAuth access token", "Authorization header is missing or invalid.")
	}
)
