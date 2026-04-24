package handler

import (
	"net/http"
	"strings"

	"acl-challenge/internal/usecase"

	"github.com/gin-gonic/gin"
)

type createNotificationRequest struct {
	Recipient string `json:"recipient"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	Channel   string `json:"channel"`
}

type updateNotificationRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Channel string `json:"channel"`
}

func ListNotifications(c *gin.Context) {
	Success(c, http.StatusOK, []interface{}{})
}

func CreateNotification(c *gin.Context) {
	var req createNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logAndRespondWithError(c, usecase.ErrInvalidInput, "create notification request body invalid")
		return
	}

	if strings.TrimSpace(req.Recipient) == "" ||
		strings.TrimSpace(req.Title) == "" ||
		strings.TrimSpace(req.Content) == "" ||
		strings.TrimSpace(req.Channel) == "" {
		logAndRespondWithError(c, usecase.ErrInvalidInput, "create notification missing required fields")
		return
	}

	Success(c, http.StatusCreated, gin.H{"message": "stub: created"})
}

func GetNotification(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		logAndRespondWithError(c, usecase.ErrInvalidInput, "get notification id is required")
		return
	}

	Success(c, http.StatusOK, gin.H{"id": id})
}

func UpdateNotification(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		logAndRespondWithError(c, usecase.ErrInvalidInput, "update notification id is required")
		return
	}

	var req updateNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logAndRespondWithError(c, usecase.ErrInvalidInput, "update notification request body invalid")
		return
	}

	if strings.TrimSpace(req.Title) == "" &&
		strings.TrimSpace(req.Content) == "" &&
		strings.TrimSpace(req.Channel) == "" {
		logAndRespondWithError(c, usecase.ErrInvalidInput, "update notification has no fields to update")
		return
	}

	Success(c, http.StatusOK, gin.H{"message": "stub: updated"})
}

func DeleteNotification(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		logAndRespondWithError(c, usecase.ErrInvalidInput, "delete notification id is required")
		return
	}

	c.Status(http.StatusNoContent)
}
