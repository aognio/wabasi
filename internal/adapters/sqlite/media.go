package sqlite

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"gorm.io/gorm"
)

type MediaRecord struct {
	gorm.Model
	ExternalID string `gorm:"uniqueIndex"`
	FileName   string
	MimeType   string `gorm:"size:128"`
	FileSize   int64
	FilePath   string
}

func (MediaRecord) TableName() string {
	return "media"
}

type MediaStore struct {
	db   *gorm.DB
	base string
}

func NewMediaStore(db *gorm.DB, basePath string) *MediaStore {
	os.MkdirAll(filepath.Join(basePath, "media"), 0755)
	return &MediaStore{db: db, base: basePath}
}

func (s *MediaStore) Store(ctx context.Context, filename string, data []byte, mimeType string) (string, error) {
	id := fmt.Sprintf("%s_%s", generateRandomID(), sanitizeFilename(filename))
	relPath := filepath.Join("media", id)
	absPath := filepath.Join(s.base, relPath)

	if err := os.WriteFile(absPath, data, 0644); err != nil {
		return "", fmt.Errorf("write media file: %w", err)
	}

	rec := MediaRecord{
		ExternalID: id,
		FileName:   filename,
		MimeType:   mimeType,
		FileSize:   int64(len(data)),
		FilePath:   relPath,
	}
	if err := s.db.WithContext(ctx).Create(&rec).Error; err != nil {
		os.Remove(absPath)
		return "", fmt.Errorf("create media record: %w", err)
	}

	return id, nil
}

func (s *MediaStore) Get(ctx context.Context, id string) ([]byte, string, error) {
	var rec MediaRecord
	if err := s.db.WithContext(ctx).Where("external_id = ?", id).First(&rec).Error; err != nil {
		return nil, "", fmt.Errorf("get media: %w", err)
	}

	data, err := os.ReadFile(filepath.Join(s.base, rec.FilePath))
	if err != nil {
		return nil, "", fmt.Errorf("read media file: %w", err)
	}

	return data, rec.MimeType, nil
}

func (s *MediaStore) GetURL(ctx context.Context, id string) (string, error) {
	return fmt.Sprintf("/media/%s", id), nil
}

func generateRandomID() string {
	b := make([]byte, 8)
	for i := range b {
		b[i] = "0123456789abcdef"[int(b[i])%16]
	}
	return string(b)
}

func sanitizeFilename(name string) string {
	cleaned := ""
	for _, c := range name {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '.' || c == '-' || c == '_' {
			cleaned += string(c)
		}
	}
	if cleaned == "" {
		cleaned = "file"
	}
	return cleaned
}
