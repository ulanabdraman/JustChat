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

	"JustChat/internal/middleware"

	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"os"
)

func main() {
	dbconn := db.InitDB()
	defer dbconn.Close()

	fmt.Println("Приложение запущено, база данных подключена.")

	router := gin.Default()

	userRepository := userRepo.NewUserRepo(dbconn)
	userUseCase := userUCpkg.NewUserUseCase(userRepository)

	authUC := authUCpkg.NewJWTUsecase("supersecret")
	authHandler := authHand.NewHandler(authUC, userUseCase)

	chatRepository := chatRepo.NewChatRepo(dbconn)
	chatUseCase := chatUC.NewChatUseCase(chatRepository)

	messageRepository := messageRepo.NewMessageRepo(dbconn)
	messageUseCase := messageUC.NewMessageUseCase(messageRepository)

	hub := websock.NewHub()

	api := router.Group("/api")

	api.POST("/login", authHandler.Login)
	api.POST("/register", authHandler.Register)

	// WebSocket (без middleware пока, если нужен — надо передавать токен)
	api.GET("/ws", func(c *gin.Context) {
		websock.ServeWS(hub, chatUseCase, messageUseCase, c.Writer, c.Request)
	})

	// 401 защита
	protected := api.Group("/")
	protected.Use(middleware.AuthMiddleware(authUC))
	chatHand.NewChatHandler(protected, chatUseCase)
	messageHand.NewMessageHandler(protected, messageUseCase)
	userHand.NewUserHandler(protected, userUseCase)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	go hub.Run()
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Не удалось запустить сервер: %v", err)
	}
}
