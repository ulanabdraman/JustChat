package main

import (
	"JustChat/db"
	"JustChat/internal/realtime/websock/transport"
	"context"
	"github.com/gin-contrib/cors"
	"time"

	authHand "JustChat/internal/auth/delivery/http"
	authUCpkg "JustChat/internal/auth/usecase"

	chatmembersHand "JustChat/internal/chatmembers/delivery/http"
	chatmembersRepo "JustChat/internal/chatmembers/repository/postgres"
	chatmembersUC "JustChat/internal/chatmembers/usecase"

	chatHand "JustChat/internal/chat/delivery/http"
	chatRepo "JustChat/internal/chat/repository/postgres"
	chatUC "JustChat/internal/chat/usecase"

	messageHand "JustChat/internal/messages/delivery/http"
	messageRepo "JustChat/internal/messages/repository/postgres"
	messageUC "JustChat/internal/messages/usecase"

	userHand "JustChat/internal/users/delivery/http"
	userRepo "JustChat/internal/users/repository/postgres"
	userUCpkg "JustChat/internal/users/usecase"

	WSHand "JustChat/internal/realtime/websock/handler"
	WSUC "JustChat/internal/realtime/websock/usecase"

	"JustChat/internal/middleware"

	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"os"
)

func main() {
	dbconn := db.InitDB()
	defer dbconn.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fmt.Println("Приложение запущено, база данных подключена.")

	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:8000"}, // или "*", если ты не паришься
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	hub := transport.NewHub()
	messageCh := make(chan []byte, 100) // буфер на 100, чтобы не блокировать горутины

	WSUsecase := WSUC.NewWebSockUsecase(hub)
	WSHandler := WSHand.NewWebSockHandler(messageCh, WSUsecase)

	chatmembersRepository := chatmembersRepo.NewChatMemberRepository(dbconn)
	chatmembersUsecase := chatmembersUC.NewChatMemberUseCase(chatmembersRepository)

	userRepository := userRepo.NewUserRepo(dbconn)
	userUseCase := userUCpkg.NewUserUseCase(userRepository)

	authUC := authUCpkg.NewJWTUsecase("supersecret")
	authHandler := authHand.NewHandler(authUC, userUseCase)

	chatRepository := chatRepo.NewChatRepo(dbconn)
	chatUseCase := chatUC.NewChatUseCase(chatRepository, chatmembersUsecase)

	messageRepository := messageRepo.NewMessageRepo(dbconn)
	messageUseCase := messageUC.NewMessageUseCase(messageRepository, messageCh, chatmembersUsecase)

	api := router.Group("/api")

	api.POST("/login", authHandler.Login)
	api.POST("/register", authHandler.Register)

	// WebSocket
	api.GET("/ws", func(c *gin.Context) {
		transport.ServeWS(hub, chatUseCase, messageUseCase, authUC, c.Writer, c.Request)
	})

	// 401 защита
	protected := api.Group("/")
	protected.Use(middleware.AuthMiddleware(authUC))
	protected.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:8000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	chatmembersHand.NewChatMemberHandler(protected, chatmembersUsecase)
	chatHand.NewChatHandler(protected, chatUseCase)
	messageHand.NewMessageHandler(protected, messageUseCase)
	userHand.NewUserHandler(protected, userUseCase)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	go hub.Run(ctx)
	go WSHandler.ListenAndServe(ctx)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Не удалось запустить сервер: %v", err)
	}
}
