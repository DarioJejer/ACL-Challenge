package main

import (
	"acl-challenge/internal/api/handler"
	"acl-challenge/internal/api/router"
	"acl-challenge/internal/domain/entity"
	notificationinfra "acl-challenge/internal/infrastructure/notification"
	"acl-challenge/internal/infrastructure/persistence"
	"acl-challenge/internal/usecase"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	initLogger(getEnv("APP_ENV"))

	db, err := persistence.NewDB(persistence.Config{
		Host:     getEnv("DB_HOST"),
		Port:     getEnv("DB_PORT"),
		User:     getEnv("DB_USER"),
		Password: getEnv("DB_PASSWORD"),
		Name:     getEnv("DB_NAME"),
		SSLMode:  getEnv("DB_SSLMODE"),
	})
	if err != nil {
		log.Fatalf("database startup failed: %v", err)
	}

	if parseBool(getEnv("RUN_MIGRATIONS")) {
		sqlDB, err := db.DB()
		if err != nil {
			log.Fatalf("database startup failed: cannot access sql db: %v", err)
		}

		migrationsPath, err := filepath.Abs("migrations")
		if err != nil {
			log.Fatalf("database startup failed: cannot resolve migrations path: %v", err)
		}

		if err := persistence.RunMigrations(sqlDB, migrationsPath); err != nil {
			log.Fatalf("database startup failed: migrations failed: %v", err)
		}
	}

	senderRegistry := notificationinfra.SenderRegistry{
		entity.ChannelEmail:            &notificationinfra.EmailSender{},
		entity.ChannelSMS:              &notificationinfra.SMSSender{},
		entity.ChannelPushNotification: &notificationinfra.PushSender{},
	}

	userRepo := persistence.NewUserRepository(db)
	notifRepo := persistence.NewNotificationRepository(db)

	userUC := usecase.NewUserUseCase(userRepo)
	notifUC := usecase.NewNotificationUseCase(userRepo, notifRepo, senderRegistry)

	userHandler := handler.NewUserHandler(userUC)
	notifHandler := handler.NewNotificationHandler(notifUC)

	r := router.NewRouter(router.Dependencies{
		UserHandler:         userHandler,
		NotificationHandler: notifHandler,
	})
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	if err := r.Run("0.0.0.0:" + port); err != nil {
		log.Fatal(err)
	}
}

func initLogger(appEnv string) {
	var logger *slog.Logger
	if strings.EqualFold(strings.TrimSpace(appEnv), "production") {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	} else {
		logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	}
	slog.SetDefault(logger)
}

func getEnv(key string) string {
	return strings.TrimSpace(os.Getenv(key))
}

func parseBool(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "y", "on":
		return true
	default:
		return false
	}
}
