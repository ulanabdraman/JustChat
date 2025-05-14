package handler

import (
	"JustChat/internal/chat/model"
	"net/http"
	"strconv"

	"JustChat/internal/chat/usecase"
	"github.com/gin-gonic/gin"
)

type ChatHandler struct {
	uc usecase.ChatUsecase
}

func NewChatHandler(r *gin.RouterGroup, uc usecase.ChatUsecase) {
	h := &ChatHandler{uc: uc}
	chats := r.Group("/chats")
	{
		chats.GET("/:id", h.GetChat)
		chats.POST("/", h.CreateChat)
		chats.PUT("/:id", h.UpdateChatName)
		chats.POST("/:id/user", h.AddUser)
		chats.DELETE("/:id/user", h.RemoveUser)
		chats.DELETE("/:id/delete", h.DeleteChat)
	}
}

func (h *ChatHandler) GetChat(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	chat, err := h.uc.GetChatByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, chat)
}

func (h *ChatHandler) CreateChat(c *gin.Context) {
	var chat model.Chat
	if err := c.ShouldBindJSON(&chat); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Получаем X-User-ID из заголовка
	userIDStr := c.GetHeader("X-User-ID")
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing X-User-ID header"})
		return
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid X-User-ID header"})
		return
	}

	// Устанавливаем created_by
	chat.CreatedBy = userID

	// Создаём чат
	err = h.uc.CreateChat(c.Request.Context(), &chat)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, chat)
}

func (h *ChatHandler) UpdateChatName(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := h.uc.UpdateChatName(c.Request.Context(), id, req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *ChatHandler) AddUser(c *gin.Context) {
	chatID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var request struct {
		UserID int64 `json:"user_id"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if err := h.uc.AddUserToChat(c.Request.Context(), chatID, request.UserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add user"})
		return
	}

	c.Status(http.StatusNoContent)
}

// RemoveUser удаляет пользователя из чата
func (h *ChatHandler) RemoveUser(c *gin.Context) {
	chatID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var request struct {
		UserID int64 `json:"user_id"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if err := h.uc.RemoveUserFromChat(c.Request.Context(), chatID, request.UserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove user"})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *ChatHandler) DeleteChat(c *gin.Context) {
	chatID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	if err := h.uc.DeleteChat(c.Request.Context(), chatID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear chat"})
		return
	}

	c.Status(http.StatusNoContent)
}
