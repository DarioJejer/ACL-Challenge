package handler

import (
	"net/http"
	"strings"

	"acl-challenge/internal/api/dtos/request"
	"acl-challenge/internal/api/middleware"
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
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		logAndRespondWithError(c, usecase.ErrUnauthorized, "list notifications missing user context")
		return
	}

	notifications, err := h.notifUC.ListNotifications(c.Request.Context(), userID)
	if err != nil {
		logAndRespondWithError(c, err, "list notifications failed")
		return
	}

	Success(c, http.StatusOK, toNotificationDTOList(notifications))
}

func (h *NotificationHandler) CreateNotification(c *gin.Context) {
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		logAndRespondWithError(c, usecase.ErrUnauthorized, "create notification missing user context")
		return
	}

	var req request.ResquestNotificationDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		logAndRespondWithError(c, usecase.ErrInvalidInput, "create notification request body invalid")
		return
	}

	if strings.TrimSpace(req.Title) == "" ||
		strings.TrimSpace(req.Content) == "" ||
		strings.TrimSpace(string(req.Channel)) == "" {
		logAndRespondWithError(c, usecase.ErrInvalidInput, "create notification missing required fields")
		return
	}

	notification, err := h.notifUC.CreateNotification(c.Request.Context(), userID, req)
	if err != nil {
		logAndRespondWithError(c, err, "create notification failed")
		return
	}

	Success(c, http.StatusCreated, toNotificationDTO(notification))
}

func (h *NotificationHandler) GetNotification(c *gin.Context) {
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		logAndRespondWithError(c, usecase.ErrUnauthorized, "get notification missing user context")
		return
	}

	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		logAndRespondWithError(c, usecase.ErrInvalidInput, "get notification id is required")
		return
	}

	notification, err := h.notifUC.GetNotification(c.Request.Context(), id, userID)
	if err != nil {
		logAndRespondWithError(c, err, "get notification failed")
		return
	}

	Success(c, http.StatusOK, toNotificationDTO(notification))
}

func (h *NotificationHandler) UpdateNotification(c *gin.Context) {
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		logAndRespondWithError(c, usecase.ErrUnauthorized, "update notification missing user context")
		return
	}

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
		strings.TrimSpace(string(req.Channel)) == "" {
		logAndRespondWithError(c, usecase.ErrInvalidInput, "update notification has no fields to update")
		return
	}

	notification, err := h.notifUC.UpdateNotification(c.Request.Context(), id, userID, req)
	if err != nil {
		logAndRespondWithError(c, err, "update notification failed")
		return
	}

	Success(c, http.StatusOK, toNotificationDTO(notification))
}

func (h *NotificationHandler) DeleteNotification(c *gin.Context) {
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		logAndRespondWithError(c, usecase.ErrUnauthorized, "delete notification missing user context")
		return
	}

	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		logAndRespondWithError(c, usecase.ErrInvalidInput, "delete notification id is required")
		return
	}

	if err := h.notifUC.DeleteNotification(c.Request.Context(), id, userID); err != nil {
		logAndRespondWithError(c, err, "delete notification failed")
		return
	}

	c.Status(http.StatusNoContent)
}
