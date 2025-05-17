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
	chats := r.Group("/chat")
	{
		chats.GET("/:id", h.GetChat)
		chats.POST("/get", h.GetChats)
		chats.POST("/", h.CreateChat)
		chats.PUT("/:id", h.UpdateChatName)
		chats.DELETE("/:id", h.DeleteChat)
	}
}

func (h *ChatHandler) GetChat(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	myuserIDStr := c.GetHeader("X-User-ID")
	myuserID, err := strconv.ParseInt(myuserIDStr, 10, 64)
	if err != nil || myuserID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	chat, err := h.uc.GetChatByID(c.Request.Context(), id, myuserID)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, chat)
}

func (h *ChatHandler) GetChats(c *gin.Context) {
	type ChatIDsRequest struct {
		Chats []int64 `json:"chats"`
	}
	var req ChatIDsRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chat_ids array"})
		return
	}

	if len(req.Chats) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "chat_ids cannot be empty"})
		return
	}

	myuserIDStr := c.GetHeader("X-User-ID")
	myuserID, err := strconv.ParseInt(myuserIDStr, 10, 64)
	if err != nil || myuserID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	chats, err := h.uc.GetChatsByIDs(c.Request.Context(), req.Chats, myuserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, chats)
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
	chatID, err := h.uc.CreateChat(c.Request.Context(), &chat, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, chatID)
}

func (h *ChatHandler) UpdateChatName(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req struct {
		Name string `json:"name"`
	}
	myuserIDStr := c.GetHeader("X-User-ID")
	myuserID, err := strconv.ParseInt(myuserIDStr, 10, 64)
	if err != nil || myuserID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err = h.uc.UpdateChatName(c.Request.Context(), id, req.Name, myuserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *ChatHandler) DeleteChat(c *gin.Context) {
	chatID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	myuserIDStr := c.GetHeader("X-User-ID")
	myuserID, err := strconv.ParseInt(myuserIDStr, 10, 64)
	if err != nil || myuserID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	if err := h.uc.DeleteChat(c.Request.Context(), chatID, myuserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
