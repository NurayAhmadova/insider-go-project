package services

import (
	"context"
	"insider-go-project/internal/message-processor/storage/model"
)

//go:generate mockgen -source=services.go -destination=mock_services.go -package=services

type (
	MessageProcessorService interface {
		ListSentMessages(ctx context.Context) ([]model.Message, error)
		StartScheduler(_ context.Context) error
		StopScheduler(_ context.Context) error
	}

	MessageSchedulerService interface {
		Start() <-chan struct{}
		Stop()
	}
)
