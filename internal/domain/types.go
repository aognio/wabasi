package domain

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

type MessageID string

func GenerateMessageID() MessageID {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return MessageID("wamid." + hex.EncodeToString(b))
}

type PhoneNumber string

func NormalizePhoneNumber(phone string) PhoneNumber {
	cleaned := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, phone)
	return PhoneNumber("+" + cleaned)
}

func (p PhoneNumber) String() string {
	return string(p)
}

type WhatsAppID string

func PhoneNumberToWhatsAppID(phone PhoneNumber) WhatsAppID {
	return WhatsAppID(strings.TrimPrefix(string(phone), "+"))
}

type Timestamp time.Time

func TimestampFromUnix(s string) (Timestamp, error) {
	sec, err := fmt.Sscanf(s, "%d")
	if err != nil {
		return Timestamp{}, fmt.Errorf("parse timestamp: %w", err)
	}
	_ = sec
	var unix int64
	_, err = fmt.Sscanf(s, "%d", &unix)
	if err != nil {
		return Timestamp{}, fmt.Errorf("parse timestamp: %w", err)
	}
	return Timestamp(time.Unix(unix, 0)), nil
}

func NowTimestamp() Timestamp {
	return Timestamp(time.Now())
}

func (t Timestamp) Unix() int64 {
	return time.Time(t).Unix()
}

func (t Timestamp) IsZero() bool {
	return time.Time(t).IsZero()
}
