package handler

import (
	"net/http"
	"strings"

	"acl-challenge/pkg/response"
	"github.com/gin-gonic/gin"
)

type createNotificationRequest struct {
	UserID  string `json:"user_id"`
	Title   string `json:"title"`
	Content string `json:"content"`
	Channel string `json:"channel"`
}

type updateNotificationRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Channel string `json:"channel"`
}

func ListNotifications(c *gin.Context) {
	response.Success(c, http.StatusOK, []interface{}{})
}

func CreateNotification(c *gin.Context) {
	var req createNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	if strings.TrimSpace(req.UserID) == "" ||
		strings.TrimSpace(req.Title) == "" ||
		strings.TrimSpace(req.Content) == "" ||
		strings.TrimSpace(req.Channel) == "" {
		response.Error(c, http.StatusBadRequest, "user_id, title, content and channel are required")
		return
	}

	response.Success(c, http.StatusCreated, gin.H{"message": "stub: created"})
}

func GetNotification(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		response.Error(c, http.StatusBadRequest, "id is required")
		return
	}

	response.Success(c, http.StatusOK, gin.H{"id": id})
}

func UpdateNotification(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		response.Error(c, http.StatusBadRequest, "id is required")
		return
	}

	var req updateNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	if strings.TrimSpace(req.Title) == "" &&
		strings.TrimSpace(req.Content) == "" &&
		strings.TrimSpace(req.Channel) == "" {
		response.Error(c, http.StatusBadRequest, "at least one field is required")
		return
	}

	response.Success(c, http.StatusOK, gin.H{"message": "stub: updated"})
}

func DeleteNotification(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		response.Error(c, http.StatusBadRequest, "id is required")
		return
	}

	c.Status(http.StatusNoContent)
}
