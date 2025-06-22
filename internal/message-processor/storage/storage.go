package storage

import (
	"context"
	"github.com/google/uuid"
	"insider-go-project/internal/message-processor/storage/model"
)

//go:generate mockgen -source=storage.go -destination=mock_storage.go -package=storage

type MessagesRepository interface {
	ListUnsentMessages(ctx context.Context, limit int32) ([]model.Message, error)
	ListSentMessages(ctx context.Context) ([]model.Message, error)
	UpdateSentStatus(ctx context.Context, id uuid.UUID) error
}
