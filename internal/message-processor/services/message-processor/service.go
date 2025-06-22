package messageprocessor

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	messagescheduler "insider-go-project/internal/message-processor/services/message-scheduler"
	"insider-go-project/internal/message-processor/storage/model"
	"insider-go-project/internal/message-processor/storage/repository"
	"io"
	"log/slog"

	"net/http"
	"time"
)

type Service struct {
	repository  *repository.MessagesRepository
	scheduler   *messagescheduler.Scheduler
	redisClient *redis.Client
	httpClient  *http.Client
	log         *slog.Logger
	webhookURL  string
	authKey     string
	batchSize   int32
}

func NewService(repository *repository.MessagesRepository, log *slog.Logger, redisClient *redis.Client, webhookURL, authKey string, batchSize int32) *Service {
	return &Service{
		repository:  repository,
		scheduler:   messagescheduler.NewScheduler(),
		log:         log,
		httpClient:  &http.Client{Timeout: 10 * time.Second},
		redisClient: redisClient,
		webhookURL:  webhookURL,
		authKey:     authKey,
		batchSize:   batchSize,
	}
}

func (s *Service) ListSentMessages(ctx context.Context) ([]model.Message, error) {
	messages, err := s.repository.ListSentMessages(ctx)
	if err != nil {
		s.log.Error("listing sent messages", slog.String("error", err.Error()))

		return nil, fmt.Errorf("listing sent messages: %w", err)
	}

	return messages, nil
}

func (s *Service) StartScheduler(_ context.Context) error {
	s.log.Info("starting message scheduler")

	ch := s.scheduler.Start()
	go s.sendMessages(ch)

	return nil
}

func (s *Service) sendMessages(ch <-chan struct{}) {
	for range ch {
		ctx, cancel := context.WithTimeout(context.TODO(), 30*time.Second)

		err := s.repository.Transactional(ctx, func(repo *repository.MessagesRepository) error {
			messages, err := repo.ListUnsentMessages(ctx, s.batchSize)
			if err != nil {
				s.log.Error("listing unsent messages", slog.String("error", err.Error()))
				return fmt.Errorf("listing unsent messages: %w", err)
			}

			if len(messages) == 0 {
				s.log.Info("no unsent messages to process")

				return nil
			}

			for _, message := range messages {
				s.log.Info("sending message", slog.String("id", message.ID.String()), slog.String("to", message.Msisdn))

				messageID, err := s.sendMessage(ctx, message.Msisdn, message.Content)
				if err != nil {
					s.log.Error("sending message", slog.String("error", err.Error()))

					break
				}
				err = repo.UpdateSentStatus(ctx, message.ID)
				if err != nil {
					s.log.Warn("updating sent status", slog.String("error", err.Error()), slog.String("id", message.ID.String()))
					return nil // we need transaction to still commit even if we only send 1 message
				}

				if s.redisClient != nil {
					now := time.Now().UTC()
					err = s.redisClient.Set(ctx, "sent:"+messageID+":"+message.Msisdn, now.Format(time.RFC3339), 7*24*time.Hour).Err()
					if err != nil {
						s.log.Warn("setting redis sent key", slog.String("error", err.Error()), slog.String("message_id", messageID))
					}
				}
			}

			return nil
		})
		cancel()

		if err != nil {
			return
		}
	}

}

func (s *Service) StopScheduler(_ context.Context) error {
	s.log.Info("stopping message scheduler")

	s.scheduler.Stop()
	return nil
}

func (s *Service) sendMessage(ctx context.Context, to, content string) (string, error) {
	payload := map[string]string{
		"to":      to,
		"content": content,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.webhookURL, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-ins-auth-key", s.authKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			s.log.Error("closing http body", err)
		}
	}()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusAccepted {
		s.log.Error("unexpected status code", slog.Int("status", resp.StatusCode), slog.String("body", string(bodyBytes)))
		return "", fmt.Errorf("unexpected status: %s", resp.Status)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "application/json" || contentType == "application/json; charset=utf-8" {
		var respBody struct {
			Message   string `json:"message"`
			MessageID string `json:"messageId"`
		}
		if err := json.Unmarshal(bodyBytes, &respBody); err != nil {
			return "", fmt.Errorf("decoding response body: %w", err)
		}
		if respBody.MessageID == "" {
			return "", errors.New("empty messageId in response")
		}
		return respBody.MessageID, nil
	}

	s.log.Warn("non-json response body", slog.String("body", string(bodyBytes)))
	return "unknown-message-id", nil
}
