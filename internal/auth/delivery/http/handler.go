package http

import (
	"JustChat/internal/auth/usecase"
	userModel "JustChat/internal/users/model"
	userUC "JustChat/internal/users/usecase"
	"JustChat/pkg/hash"
	"github.com/gin-gonic/gin"
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
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}
	if !hash.ComparePassword(user.Password, req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "wrong password"})
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
		Type     string `json:"type"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	founduser, err := h.userUC.GetByUsername(c, req.Username)
	if founduser != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "user already exists"})
		return
	}

	hashedPassword, err := hash.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	user := &userModel.User{
		Username: req.Username,
		Password: hashedPassword,
		Type:     req.Type,
	}

	if err := h.userUC.CreateUser(c, user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user registered successfully"})
}
