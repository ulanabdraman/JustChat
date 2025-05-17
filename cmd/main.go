package main

import (
	"JustChat/db"
	authHand "JustChat/internal/auth/delivery/http"
	authUCpkg "JustChat/internal/auth/usecase"
	"net/http"
	"os/signal"
	"syscall"

	chatHand "JustChat/internal/chat/delivery/http"
	chatRepo "JustChat/internal/chat/repository/postgres"
	chatUC "JustChat/internal/chat/usecase"

	chatmembersHand "JustChat/internal/chatmembers/delivery/http"
	chatmembersRepo "JustChat/internal/chatmembers/repository/postgres"
	chatmembersUC "JustChat/internal/chatmembers/usecase"

	messageHand "JustChat/internal/messages/delivery/http"
	messageRepo "JustChat/internal/messages/repository/postgres"
	messageUC "JustChat/internal/messages/usecase"

	userHand "JustChat/internal/users/delivery/http"
	userRepo "JustChat/internal/users/repository/postgres"
	userUCpkg "JustChat/internal/users/usecase"

	WSHand "JustChat/internal/realtime/websock/handler"
	WSTP "JustChat/internal/realtime/websock/transport"
	WSUC "JustChat/internal/realtime/websock/usecase"

	StreamHub "JustChat/pkg/streamhub"

	"JustChat/internal/middleware"

	"context"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"log"
	"os"
	"time"
)

func main() {
	dbconn := db.InitDB()
	defer dbconn.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fmt.Println("Приложение запущено, база данных подключена.")

	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:8000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	streamHub := StreamHub.NewStreamHub(100)
	defer streamHub.Stop()
	hub := WSTP.NewHub()
	messageCh := make(chan []byte, 100)
	defer close(messageCh)

	wsUsecase := WSUC.NewWebSockUsecase(hub)
	wsHandler := WSHand.NewWebSockHandler(messageCh, wsUsecase)

	chatmembersRepository := chatmembersRepo.NewChatMemberRepository(dbconn)
	chatmembersUseCase := chatmembersUC.NewChatMemberUseCase(chatmembersRepository)

	userRepository := userRepo.NewUserRepo(dbconn)
	userUseCase := userUCpkg.NewUserUseCase(userRepository)

	authUC := authUCpkg.NewJWTUsecase("supersecret")
	authHandler := authHand.NewHandler(authUC, userUseCase)

	chatRepository := chatRepo.NewChatRepo(dbconn)
	chatUseCase := chatUC.NewChatUseCase(chatRepository, chatmembersUseCase)

	messageRepository := messageRepo.NewMessageRepo(dbconn)
	messageUseCase := messageUC.NewMessageUseCase(messageRepository, messageCh, chatmembersUseCase)

	api := router.Group("/api")

	api.POST("/login", authHandler.Login)
	api.POST("/register", authHandler.Register)

	api.GET("/ws", func(c *gin.Context) {
		WSTP.ServeWS(hub, chatUseCase, messageUseCase, authUC, c.Writer, c.Request)
	})

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

	chatmembersHand.NewChatMemberHandler(protected, chatmembersUseCase)
	chatHand.NewChatHandler(protected, chatUseCase)
	messageHand.NewMessageHandler(protected, messageUseCase)
	userHand.NewUserHandler(protected, userUseCase)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	go hub.Run(ctx)
	go wsHandler.ListenAndServe(ctx)
	go streamHub.Start()

	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	// Запуск сервера в отдельной горутине
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Ошибка сервера: %v\n", err)
		}
	}()

	// Ожидание сигнала завершения
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	log.Println("Выключение сервера...")

	// Контекст с таймаутом для graceful shutdown
	Grctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(Grctx); err != nil {
		log.Fatalf("Ошибка во время shutdown: %v\n", err)
	}

	log.Println("Сервер остановлен")
}
