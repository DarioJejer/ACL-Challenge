package integration_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"acl-challenge/internal/api/dtos/response"
	"acl-challenge/internal/api/handler"
	"acl-challenge/internal/api/router"
	"acl-challenge/internal/domain/entity"
	notificationinfra "acl-challenge/internal/infrastructure/notification"
	"acl-challenge/internal/infrastructure/persistence"
	"acl-challenge/internal/usecase"
	"acl-challenge/tests/testhelper"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// notificationSuccessEnvelope matches handler.Success JSON for notification responses.
type notificationSuccessEnvelope struct {
	Success bool                     `json:"success"`
	Data    response.NotificationDTO `json:"data"`
}

func TestUserEndpointsIntegration(t *testing.T) {
	db := testhelper.SetupTestDB(t)
	r := newTestRouter(db)

	t.Run("PUT /api/v1/users/:id - 200 updated", func(t *testing.T) {
		testhelper.TruncateAll(t, db)
		user := seedUser(t, db, uuid.New())

		rec := performJSONRequest(t, r, http.MethodPut, "/api/v1/users/"+user.ID.String(), map[string]any{
			"email": "updated@example.com",
		})

		require.Equal(t, http.StatusOK, rec.Code)

		var reloaded persistence.UserModel
		require.NoError(t, db.First(&reloaded, "id = ?", user.ID).Error)
		require.Equal(t, "updated@example.com", reloaded.Email)
	})

	t.Run("PUT /api/v1/users/:id - 404 not found", func(t *testing.T) {
		testhelper.TruncateAll(t, db)
		rec := performJSONRequest(t, r, http.MethodPut, "/api/v1/users/"+uuid.NewString(), map[string]any{
			"email": "updated@example.com",
		})
		require.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("PUT /api/v1/users/:id - 400 bad request", func(t *testing.T) {
		testhelper.TruncateAll(t, db)
		rec := performJSONRequest(t, r, http.MethodPut, "/api/v1/users/"+uuid.NewString(), map[string]any{})
		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("DELETE /api/v1/users/:id - 204 deleted", func(t *testing.T) {
		testhelper.TruncateAll(t, db)
		user := seedUser(t, db, uuid.New())
		rec := performJSONRequest(t, r, http.MethodDelete, "/api/v1/users/"+user.ID.String(), nil)
		require.Equal(t, http.StatusNoContent, rec.Code)

		var count int64
		require.NoError(t, db.Model(&persistence.UserModel{}).Where("id = ?", user.ID).Count(&count).Error)
		require.Equal(t, int64(0), count)
	})

	t.Run("DELETE /api/v1/users/:id - 404 not found", func(t *testing.T) {
		testhelper.TruncateAll(t, db)
		rec := performJSONRequest(t, r, http.MethodDelete, "/api/v1/users/"+uuid.NewString(), nil)
		require.Equal(t, http.StatusNotFound, rec.Code)
	})
}

func TestNotificationEndpointsIntegration(t *testing.T) {
	db := testhelper.SetupTestDB(t)
	r := newTestRouter(db)

	t.Run("GET /api/v1/notifications - 200 list empty", func(t *testing.T) {
		testhelper.TruncateAll(t, db)
		rec := performJSONRequest(t, r, http.MethodGet, "/api/v1/notifications/", nil)
		require.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("GET /api/v1/notifications - 200 list non-empty", func(t *testing.T) {
		testhelper.TruncateAll(t, db)
		fixedUserID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
		seedUser(t, db, fixedUserID)
		seedNotification(t, db, fixedUserID)

		rec := performJSONRequest(t, r, http.MethodGet, "/api/v1/notifications/", nil)
		require.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("POST /api/v1/notifications - 201 created", func(t *testing.T) {
		testhelper.TruncateAll(t, db)
		user := seedUser(t, db, uuid.New())
		rec := performJSONRequest(t, r, http.MethodPost, "/api/v1/notifications/", map[string]any{
			"recipient": user.ID.String(),
			"title":     "Test",
			"content":   "Hello world",
			"channel":   string(entity.ChannelEmail),
		})
		require.Equal(t, http.StatusCreated, rec.Code)

		var env notificationSuccessEnvelope
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &env))
		require.True(t, env.Success)
		require.Equal(t, user.ID.String(), env.Data.Recipient)
		require.Equal(t, "Test", env.Data.Title)
		require.Equal(t, "Hello world", env.Data.Content)
		require.Equal(t, string(entity.ChannelEmail), env.Data.Channel)
	})

	t.Run("POST /api/v1/notifications - 400 bad request", func(t *testing.T) {
		testhelper.TruncateAll(t, db)
		rec := performJSONRequest(t, r, http.MethodPost, "/api/v1/notifications/", map[string]any{
			"title": "missing fields",
		})
		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("GET /api/v1/notifications/:id - 200 found", func(t *testing.T) {
		testhelper.TruncateAll(t, db)
		user := seedUser(t, db, uuid.New())
		n := seedNotification(t, db, user.ID)

		rec := performJSONRequest(t, r, http.MethodGet, "/api/v1/notifications/"+n.ID.String(), nil)
		require.Equal(t, http.StatusOK, rec.Code)

		var env notificationSuccessEnvelope
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &env))
		require.True(t, env.Success)
		require.Equal(t, user.ID.String(), env.Data.Recipient)
		require.Equal(t, n.ID.String(), env.Data.ID)
	})

	t.Run("GET /api/v1/notifications/:id - 404 not found", func(t *testing.T) {
		testhelper.TruncateAll(t, db)
		rec := performJSONRequest(t, r, http.MethodGet, "/api/v1/notifications/"+uuid.NewString(), nil)
		require.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("PUT /api/v1/notifications/:id - 200 updated", func(t *testing.T) {
		testhelper.TruncateAll(t, db)
		user := seedUser(t, db, uuid.New())
		n := seedNotification(t, db, user.ID)
		rec := performJSONRequest(t, r, http.MethodPut, "/api/v1/notifications/"+n.ID.String(), map[string]any{
			"title": "Updated title",
		})
		require.Equal(t, http.StatusOK, rec.Code)

		var env notificationSuccessEnvelope
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &env))
		require.True(t, env.Success)
		require.Equal(t, n.ID.String(), env.Data.ID)
		require.Equal(t, user.ID.String(), env.Data.Recipient)
		require.Equal(t, "Updated title", env.Data.Title)
		require.Equal(t, n.Channel, env.Data.Channel)
	})

	t.Run("PUT /api/v1/notifications/:id - 404 not found", func(t *testing.T) {
		testhelper.TruncateAll(t, db)
		rec := performJSONRequest(t, r, http.MethodPut, "/api/v1/notifications/"+uuid.NewString(), map[string]any{
			"title": "Updated title",
		})
		require.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("DELETE /api/v1/notifications/:id - 204 deleted", func(t *testing.T) {
		testhelper.TruncateAll(t, db)
		user := seedUser(t, db, uuid.New())
		n := seedNotification(t, db, user.ID)

		rec := performJSONRequest(t, r, http.MethodDelete, "/api/v1/notifications/"+n.ID.String(), nil)
		require.Equal(t, http.StatusNoContent, rec.Code)

		var notifCount int64
		require.NoError(t, db.Model(&persistence.NotificationModel{}).Count(&notifCount).Error)
		require.Equal(t, int64(0), notifCount)
	})

	t.Run("DELETE /api/v1/notifications/:id - 404 not found", func(t *testing.T) {
		testhelper.TruncateAll(t, db)
		rec := performJSONRequest(t, r, http.MethodDelete, "/api/v1/notifications/"+uuid.NewString(), nil)
		require.Equal(t, http.StatusNotFound, rec.Code)
	})
}

func newTestRouter(db *gorm.DB) http.Handler {
	userRepo := persistence.NewUserRepository(db)
	notifRepo := persistence.NewNotificationRepository(db)

	senderRegistry := notificationinfra.SenderRegistry{
		entity.ChannelEmail:            &notificationinfra.EmailSender{},
		entity.ChannelSMS:              &notificationinfra.SMSSender{},
		entity.ChannelPushNotification: &notificationinfra.PushSender{},
	}

	userUC := usecase.NewUserUseCase(userRepo)
	notifUC := usecase.NewNotificationUseCase(userRepo, notifRepo, senderRegistry)

	authHandler := handler.NewAuthHandler(userUC)
	userHandler := handler.NewUserHandler(userUC)
	notifHandler := handler.NewNotificationHandler(notifUC)

	return router.NewRouter(router.Dependencies{
		AuthHandler:         authHandler,
		UserHandler:         userHandler,
		NotificationHandler: notifHandler,
	})
}

func performJSONRequest(t *testing.T, h http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()

	var payload []byte
	var err error
	if body != nil {
		payload, err = json.Marshal(body)
		require.NoError(t, err)
	}

	req := httptest.NewRequest(method, path, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func seedUser(t *testing.T, db *gorm.DB, id uuid.UUID) persistence.UserModel {
	t.Helper()
	user := persistence.UserModel{
		ID:           id,
		Email:        "seed-" + id.String() + "@example.com",
		PasswordHash: "hash",
	}
	require.NoError(t, db.Create(&user).Error)
	return user
}

func seedNotification(t *testing.T, db *gorm.DB, userID uuid.UUID) persistence.NotificationModel {
	t.Helper()
	n := persistence.NotificationModel{
		ID:        uuid.New(),
		Recipient: userID,
		Title:     "Seed title",
		Content:   "Seed content",
		Channel:   string(entity.ChannelEmail),
	}
	require.NoError(t, db.Create(&n).Error)
	return n
}
