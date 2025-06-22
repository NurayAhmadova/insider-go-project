package responses

import (
	"time"

	"github.com/google/uuid"
)

type SentMessage struct {
	ID      uuid.UUID `json:"id"`
	Msisdn  string    `json:"msisdn"`
	Content string    `json:"content"`
	SentAt  time.Time `json:"sent_at"`
}

type SentMessagesResponse struct {
	Status   string        `json:"status"`
	Messages []SentMessage `json:"messages"`
	Error    string        `json:"error,omitempty"`
}
