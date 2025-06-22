package model

import (
	"github.com/google/uuid"
	"time"
)

type Message struct {
	ID        uuid.UUID
	Msisdn    string
	Content   string
	Sent      bool
	CreatedAt time.Time
	UpdatedAt time.Time
}
