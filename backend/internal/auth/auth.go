// Package auth — хеширование паролей (bcrypt) и JWT-токены.
package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	secret []byte
	ttl    time.Duration
}

func New(secret string) *Service {
	return &Service{secret: []byte(secret), ttl: 7 * 24 * time.Hour}
}

func HashPassword(pw string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	return string(b), err
}

func CheckPassword(hash, pw string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pw)) == nil
}

type claims struct {
	UserID int64 `json:"uid"`
	jwt.RegisteredClaims
}

func (s *Service) GenerateToken(userID int64) (string, error) {
	c := claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString(s.secret)
}

func (s *Service) ParseToken(tokenStr string) (int64, error) {
	c := &claims{}
	token, err := jwt.ParseWithClaims(tokenStr, c, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("неверный метод подписи")
		}
		return s.secret, nil
	})
	if err != nil || !token.Valid {
		return 0, errors.New("невалидный токен")
	}
	return c.UserID, nil
}
