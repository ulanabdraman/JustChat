package http

import (
	"JustChat/internal/auth/usecase"
	userUC "JustChat/internal/users/usecase"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

type Handler struct {
	authUC usecase.JWTUsecase
	userUC userUC.UserUseCase
}

func NewHandler(authUC usecase.JWTUsecase, userUC userUC.UserUseCase) *Handler {
	return &Handler{
		authUC: authUC,
		userUC: userUC,
	}
}

func (h *Handler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	user, err := h.userUC.GetByUsername(c, req.Username)
	if err != nil || user.Password != req.Password {
		// Здесь должна быть проверка через bcrypt, если используешь
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	token, err := h.authUC.GenerateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token generation failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}
func (h *Handler) Register(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// Проверим, существует ли пользователь
	_, err := h.userUC.GetByUsername(c, req.Username)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "user already exists"})
		return
	}

	// Хэшируем пароль
	_, err = bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	//var user userMode
	//
	//// Создаем пользователя
	//err = h.userUC.CreateUser(c, req.Username, string(hashedPassword))
	//if err != nil {
	//	c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
	//	return
	//}

	c.JSON(http.StatusOK, gin.H{"message": "user registered successfully"})
}
