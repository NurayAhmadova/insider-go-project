package repository

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"insider-go-project/internal/message-processor/storage/model"
)

type MessagesRepository struct {
	db      *sql.DB
	queries *Queries
}

func NewMessagesRepository(db *sql.DB) *MessagesRepository {
	return &MessagesRepository{
		db:      db,
		queries: New(db),
	}
}

func (m *MessagesRepository) Transactional(ctx context.Context, fn func(repo *MessagesRepository) error) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begintx: %w", err)
	}
	defer func(tx *sql.Tx) {
		err := tx.Rollback()
		if err != nil {
			return
		}
	}(tx)

	if err = fn(&MessagesRepository{
		db:      m.db,
		queries: m.queries.WithTx(tx),
	}); err != nil {
		return fmt.Errorf("begintx: %w", err)
	}

	return tx.Commit()
}

func (m *MessagesRepository) ListUnsentMessages(ctx context.Context, limit int32) ([]model.Message, error) {
	messages, err := m.queries.ListUnsentMessages(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("listing unsent messages: %w", err)
	}

	msgs := lo.Map(messages, func(msg Message, _ int) model.Message {
		return model.Message{
			ID:        msg.ID,
			Msisdn:    msg.Msisdn,
			Content:   msg.Content,
			Sent:      msg.Sent,
			CreatedAt: msg.CreatedAt,
			UpdatedAt: msg.UpdatedAt,
		}
	})

	return msgs, nil
}

func (m *MessagesRepository) ListSentMessages(ctx context.Context) ([]model.Message, error) {
	messages, err := m.queries.ListSentMessages(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing sent messages: %w", err)
	}

	msgs := lo.Map(messages, func(msg Message, _ int) model.Message {
		return model.Message{
			ID:        msg.ID,
			Msisdn:    msg.Msisdn,
			Content:   msg.Content,
			Sent:      msg.Sent,
			CreatedAt: msg.CreatedAt,
			UpdatedAt: msg.UpdatedAt,
		}
	})

	return msgs, nil
}

func (m *MessagesRepository) UpdateSentStatus(ctx context.Context, id uuid.UUID) error {
	err := m.queries.UpdateSentStatus(ctx, id)
	if err != nil {
		return fmt.Errorf("marking message as sent: %w", err)
	}

	return nil
}
