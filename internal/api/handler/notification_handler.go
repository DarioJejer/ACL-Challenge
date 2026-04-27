package handler

import (
	"net/http"
	"strings"

	"acl-challenge/internal/api/dtos/request"
	"acl-challenge/internal/usecase"

	"github.com/gin-gonic/gin"
)

type NotificationHandler struct {
	notifUC *usecase.NotificationUseCase
}

func NewNotificationHandler(notifUC *usecase.NotificationUseCase) *NotificationHandler {
	return &NotificationHandler{notifUC: notifUC}
}

func (h *NotificationHandler) ListNotifications(c *gin.Context) {
	userID := strings.TrimSpace(c.Param("userID"))
	// TODO(auth): replace with JWT context
	if userID == "" {
		userID = "00000000-0000-0000-0000-000000000001"
	}

	notifications, err := h.notifUC.ListNotifications(c.Request.Context(), userID)
	if err != nil {
		logAndRespondWithError(c, err, "list notifications failed")
		return
	}

	Success(c, http.StatusOK, toNotificationDTOList(notifications))
}

func (h *NotificationHandler) CreateNotification(c *gin.Context) {
	var req request.ResquestNotificationDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		logAndRespondWithError(c, usecase.ErrInvalidInput, "create notification request body invalid")
		return
	}

	if strings.TrimSpace(req.Recipient) == "" ||
		strings.TrimSpace(req.Title) == "" ||
		strings.TrimSpace(req.Content) == "" ||
		strings.TrimSpace(string(req.Channel)) == "" {
		logAndRespondWithError(c, usecase.ErrInvalidInput, "create notification missing required fields")
		return
	}

	notification, err := h.notifUC.CreateNotification(c.Request.Context(), req)
	if err != nil {
		logAndRespondWithError(c, err, "create notification failed")
		return
	}

	Success(c, http.StatusCreated, toNotificationDTO(notification))
}

func (h *NotificationHandler) GetNotification(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		logAndRespondWithError(c, usecase.ErrInvalidInput, "get notification id is required")
		return
	}

	notification, err := h.notifUC.GetNotification(c.Request.Context(), id)
	if err != nil {
		logAndRespondWithError(c, err, "get notification failed")
		return
	}

	Success(c, http.StatusOK, toNotificationDTO(notification))
}

func (h *NotificationHandler) UpdateNotification(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		logAndRespondWithError(c, usecase.ErrInvalidInput, "update notification id is required")
		return
	}

	var req request.ResquestNotificationDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		logAndRespondWithError(c, usecase.ErrInvalidInput, "update notification request body invalid")
		return
	}

	if strings.TrimSpace(req.Title) == "" &&
		strings.TrimSpace(req.Content) == "" &&
		strings.TrimSpace(string(req.Channel)) == "" &&
		strings.TrimSpace(req.Recipient) == "" {
		logAndRespondWithError(c, usecase.ErrInvalidInput, "update notification has no fields to update")
		return
	}

	notification, err := h.notifUC.UpdateNotification(c.Request.Context(), id, req)
	if err != nil {
		logAndRespondWithError(c, err, "update notification failed")
		return
	}

	Success(c, http.StatusOK, toNotificationDTO(notification))
}

func (h *NotificationHandler) DeleteNotification(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		logAndRespondWithError(c, usecase.ErrInvalidInput, "delete notification id is required")
		return
	}

	if err := h.notifUC.DeleteNotification(c.Request.Context(), id); err != nil {
		logAndRespondWithError(c, err, "delete notification failed")
		return
	}

	c.Status(http.StatusNoContent)
}
