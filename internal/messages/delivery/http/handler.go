package http

import (
	"JustChat/internal/messages/model"
	"JustChat/internal/messages/usecase"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type MessageHandler struct {
	uc usecase.MessageUseCase
}

func NewMessageHandler(r *gin.RouterGroup, uc usecase.MessageUseCase) {
	h := &MessageHandler{uc: uc}
	chats := r.Group("/message")
	{
		chats.GET("/:id", h.GetMessageByID)
		chats.GET("/chat/:chat_id", h.GetMessageByChatID)
		chats.POST("/", h.SaveMessage)
		chats.DELETE("/:id", h.DeleteMessageByID)

	}
}

func (h *MessageHandler) GetMessageByID(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	message, err := h.uc.GetMessageByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, message)
}
func (h *MessageHandler) DeleteMessageByID(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	err := h.uc.DeleteMessage(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": id})
}
func (h *MessageHandler) GetMessageByChatID(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("chat_id"), 10, 64)
	messages, err := h.uc.GetMessageByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	}
	c.JSON(http.StatusOK, messages)
}
func (h *MessageHandler) SaveMessage(c *gin.Context) {
	var message model.Message

	if err := c.ShouldBindJSON(&message); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	tosend, err := h.uc.SaveMessage(c.Request.Context(), &message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save message: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, tosend)
}
