package auth_service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type AuthService struct {
	log *slog.Logger
	tokenTtl time.Duration
	tokenSecret string
}

func New(log *slog.Logger, tokenTtl time.Duration, tokenSecret string) (*AuthService){
	return &AuthService{
		log:log,
		tokenTtl: tokenTtl,
		tokenSecret: tokenSecret,
	}
}

func (a *AuthService) RegisterUser(ctx context.Context, email string, password string) (string, error){

	op:="authService/registerUser"

	log:= a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["email"] = email
	claims["exp"] = time.Now().Add(a.tokenTtl).Unix()
	tokenString, err := token.SignedString([]byte(a.tokenSecret))

	if err != nil{
		log.Info("ошибка при создании токена")
		return  "", fmt.Errorf("ошибка при создании токена: %v", err)
	}


	return tokenString, nil
}