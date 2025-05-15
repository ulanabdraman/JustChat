package http

import (
	"JustChat/internal/chatmembers/usecase"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type ChatMemberHandler struct {
	uc usecase.ChatMemberUseCase
}

func NewChatMemberHandler(r *gin.RouterGroup, uc usecase.ChatMemberUseCase) {
	h := &ChatMemberHandler{uc: uc}
	chats := r.Group("/chats")
	{
		chats.GET("/:chat_id/users", h.GetUsersByChat)                 // Получить пользователей чата
		chats.GET("/me", h.GetMyChats)                                 // Получить чаты пользователя
		chats.POST("/:chat_id/users", h.AddUserToChat)                 // Добавить пользователя в чат
		chats.DELETE("/:chat_id/users/:user_id", h.RemoveUserFromChat) // Удалить пользователя из чата
		chats.PUT("/:chat_id/users/", h.UpdateUserRole)                // Обновить роль пользователя в чате
	}
}

func (h *ChatMemberHandler) GetUsersByChat(c *gin.Context) {
	chatIDStr := c.Param("chat_id")
	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chat id"})
		return
	}
	myuserIDStr := c.GetHeader("X-User-ID")
	myuserID, err := strconv.ParseInt(myuserIDStr, 10, 64) // здесь заменили на ParseInt
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user id"})
		return
	}
	users, err := h.uc.GetUsersByChat(c.Request.Context(), chatID, myuserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"users": users})
}

func (h *ChatMemberHandler) GetMyChats(c *gin.Context) {
	myuserIDStr := c.GetHeader("X-User-ID")
	myuserID, err := strconv.ParseInt(myuserIDStr, 10, 64)
	if err != nil || myuserID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	chatIDs, err := h.uc.GetChatsByUser(c.Request.Context(), myuserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"chats": chatIDs})
}

func (h *ChatMemberHandler) AddUserToChat(c *gin.Context) {
	chatIDStr := c.Param("chat_id")
	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chat id"})
		return
	}

	myuserIDStr := c.GetHeader("X-User-ID")
	myuserID, err := strconv.ParseInt(myuserIDStr, 10, 64)
	if err != nil || myuserID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid my user id"})
		return
	}

	var requestBody struct {
		UserID int64  `json:"user_id"`
		Role   string `json:"role"` // Role теперь передается в теле запроса
	}

	// Чтение тела запроса
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Если роль не указана, по умолчанию будет "user"
	if requestBody.Role == "" {
		requestBody.Role = "user"
	}

	err = h.uc.AddUserToChat(c.Request.Context(), chatID, requestBody.UserID, requestBody.Role, myuserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User added to chat"})
}

func (h *ChatMemberHandler) RemoveUserFromChat(c *gin.Context) {
	chatIDStr := c.Param("chat_id")
	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chat id"})
		return
	}

	myuserIDStr := c.GetHeader("X-User-ID")
	myuserID, err := strconv.ParseInt(myuserIDStr, 10, 64)
	if err != nil || myuserID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	var requestBody struct {
		UserID int64 `json:"user_id"`
	}

	// Чтение тела запроса
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	err = h.uc.RemoveUserFromChat(c.Request.Context(), chatID, requestBody.UserID, myuserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User removed from chat"})
}

func (h *ChatMemberHandler) UpdateUserRole(c *gin.Context) {
	chatIDStr := c.Param("chat_id")
	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chat id"})
		return
	}

	myuserIDStr := c.GetHeader("X-User-ID")
	myuserID, err := strconv.ParseInt(myuserIDStr, 10, 64)
	if err != nil || myuserID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	var requestBody struct {
		UserID int64  `json:"user_id"`
		Role   string `json:"role"`
	}

	// Чтение тела запроса
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if requestBody.Role == "" {
		requestBody.Role = "user"
	}

	err = h.uc.UpdateUserRole(c.Request.Context(), chatID, requestBody.UserID, requestBody.Role, myuserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User role updated"})
}
