package messages

import (
	handler "JustChat/internal/messages/delivery/http"
	"JustChat/internal/messages/repository/postgres"
	"JustChat/internal/messages/usecase"
	"database/sql"
	"github.com/gin-gonic/gin"
)

func Init(router *gin.RouterGroup, db *sql.DB) {
	repo := postgres.NewMessageRepo(db)
	uc := usecase.NewMessageUseCase(repo)
	handler.NewMessageHandler(router, uc)
}
