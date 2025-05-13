package usecase

import (
	"JustChat/internal/auth/model"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

type JWTUsecase interface {
	GenerateToken(userID int64) (string, error)
	ParseToken(tokenString string) (*model.AuthClaims, error)
}
type jwtUC struct {
	secretKey []byte
}

func NewJWTUsecase(secret string) JWTUsecase {
	return &jwtUC{secretKey: []byte(secret)}
}

func (u *jwtUC) GenerateToken(userID int64) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(u.secretKey)
}

func (u *jwtUC) ParseToken(tokenStr string) (*model.AuthClaims, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return u.secretKey, nil
	})

	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid claims")
	}

	uidFloat, ok := claims["user_id"].(float64)
	if !ok {
		return nil, errors.New("user_id not found")
	}

	return &model.AuthClaims{UserID: int64(uidFloat)}, nil
}
