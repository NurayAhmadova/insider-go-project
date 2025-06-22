package httpapp

import (
	"insider-go-project/internal/message-processor/transport/httpapp/v1/message-processor/requests"
	"net/http"

	"github.com/labstack/echo/v4"
	messageprocessor "insider-go-project/internal/message-processor/services/message-processor"
	"insider-go-project/internal/message-processor/transport/httpapp/v1/message-processor/responses"
)

type Handler struct {
	service *messageprocessor.Service
}

func NewHandler(s *messageprocessor.Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) RegisterRoutes(e *echo.Echo) {
	e.POST("/scheduler", h.Scheduler)
	e.GET("/messages/sent", h.GetSentMessages)
}

func (h *Handler) GetSentMessages(c echo.Context) error {
	messages, err := h.service.ListSentMessages(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, responses.SimpleResponse{
			Message: "failed to list sent messages",
		})
	}
	return c.JSON(http.StatusOK, messages)
}

func (h *Handler) Scheduler(c echo.Context) error {
	req := new(requests.SchedulerRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, responses.SimpleResponse{
			Message: "invalid request payload",
		})
	}

	switch req.Action {
	case "start":
		if err := h.service.StartScheduler(c.Request().Context()); err != nil {
			return c.JSON(http.StatusInternalServerError, responses.SimpleResponse{
				Message: "failed to start auto-send",
			})
		}
		return c.JSON(http.StatusAccepted, responses.SimpleResponse{
			Message: "auto-send started",
		})

	case "stop":
		if err := h.service.StopScheduler(c.Request().Context()); err != nil {
			return c.JSON(http.StatusInternalServerError, responses.SimpleResponse{
				Message: "failed to stop auto-send",
			})
		}
		return c.JSON(http.StatusAccepted, responses.SimpleResponse{
			Message: "auto-send stopped",
		})

	default:
		return c.JSON(http.StatusBadRequest, responses.SimpleResponse{
			Message: "invalid action, must be 'start' or 'stop'",
		})
	}
}
