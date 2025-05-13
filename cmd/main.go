package main

import (
	"JustChat/db"

	authHand "JustChat/internal/auth/delivery/http"
	authUCpkg "JustChat/internal/auth/usecase"

	chatHand "JustChat/internal/chat/delivery/http"
	chatRepo "JustChat/internal/chat/repository/postgres"
	chatUC "JustChat/internal/chat/usecase"

	messageHand "JustChat/internal/messages/delivery/http"
	messageRepo "JustChat/internal/messages/repository/postgres"
	messageUC "JustChat/internal/messages/usecase"

	"JustChat/internal/realtime/websock"

	userHand "JustChat/internal/users/delivery/http"
	userRepo "JustChat/internal/users/repository/postgres"
	userUCpkg "JustChat/internal/users/usecase"

	"JustChat/internal/middleware" // <- добавь сам middleware

	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"os"
)

func main() {
	// Инициализация БД
	dbconn := db.InitDB()
	defer dbconn.Close()

	fmt.Println("Приложение запущено, база данных подключена.")

	// Инициализация роутера
	router := gin.Default()

	// Инициализация зависимостей
	userRepository := userRepo.NewUserRepo(dbconn)
	userUseCase := userUCpkg.NewUserUseCase(userRepository)

	authUC := authUCpkg.NewJWTUsecase("supersecret")
	authHandler := authHand.NewHandler(authUC, userUseCase)

	chatRepository := chatRepo.NewChatRepo(dbconn)
	chatUseCase := chatUC.NewChatUseCase(chatRepository)

	messageRepository := messageRepo.NewMessageRepo(dbconn)
	messageUseCase := messageUC.NewMessageUseCase(messageRepository)

	hub := websock.NewHub()

	// API группы
	api := router.Group("/api")

	// Public routes
	api.POST("/login", authHandler.Login) // <-- login endpoint
	userHand.NewUserHandler(api, userUseCase)

	// WebSocket (без middleware пока, если нужен — надо передавать токен)
	api.GET("/ws", func(c *gin.Context) {
		websock.ServeWS(hub, chatUseCase, messageUseCase, c.Writer, c.Request)
	})

	// Protected routes
	protected := api.Group("/")
	protected.Use(middleware.AuthMiddleware(authUC)) // <-- middleware подключения
	chatHand.NewChatHandler(protected, chatUseCase)
	messageHand.NewMessageHandler(protected, messageUseCase)

	// Порт можно брать из env
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Запуск сервера
	go hub.Run()
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Не удалось запустить сервер: %v", err)
	}
}
