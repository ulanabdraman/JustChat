package http

import (
	"JustChat/internal/users/model"
	"JustChat/internal/users/usecase"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type UserHandler struct {
	uc usecase.UserUseCase
}

func NewUserHandler(r *gin.RouterGroup, uc usecase.UserUseCase) {
	h := &UserHandler{uc: uc}
	me := r.Group("/users")
	{
		me.GET("/me", h.GetMe)
		me.POST("/get", h.GetManyUsers)
		me.POST("/", h.CreateUser)
	}
}

func (h *UserHandler) GetMe(c *gin.Context) {
	// Предположим, что user_id приходит из middleware и кладётся в context (можно заменить потом)
	userIDStr := c.GetHeader("X-User-ID")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil || userID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	user, err := h.uc.GetCurrentUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}
func (h *UserHandler) GetManyUsers(c *gin.Context) {
	// Внутренняя структура для парсинга тела запроса
	var req struct {
		UserIDs []int64 `json:"user_ids"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_ids array"})
		return
	}
	if len(req.UserIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_ids cannot be empty"})
		return
	}

	users, err := h.uc.GetUsersByIDs(c.Request.Context(), req.UserIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, users)
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var user model.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}
	err := h.uc.CreateUser(c.Request.Context(), &user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	c.JSON(http.StatusCreated, user)
}
