package auth_service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	tokens_helper "github.com/KBcHMFollower/auth-service/internal/domain/lib/tokens"
	"github.com/KBcHMFollower/auth-service/internal/domain/models"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserSaver interface {
	CreateUser(ctx context.Context, email string, hashPass []byte) (uuid.UUID, error)
}

type UserGetter interface {
	GetUser(ctx context.Context, email string) (models.User, error)
}

type AuthService struct {
	log         *slog.Logger
	tokenTtl    time.Duration
	tokenSecret string
	userSaver   UserSaver
	userGetter  UserGetter
}

func New(log *slog.Logger, tokenTtl time.Duration, tokenSecret string, userSaver UserSaver, userGetter UserGetter) *AuthService {
	return &AuthService{
		log:         log,
		tokenTtl:    tokenTtl,
		tokenSecret: tokenSecret,
		userSaver:   userSaver,
		userGetter:  userGetter,
	}
}

func (a *AuthService) RegisterUser(ctx context.Context, email string, password string) (string, error) {

	op := "authService/registerUser"

	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	hashPass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("ошибка генерации хэша пароля: ", err)
		return "", fmt.Errorf("%w", err)
	}

	userId, err := a.userSaver.CreateUser(ctx, email, hashPass)
	if err != nil {
		log.Error("при создании пользователя: ", err)
		return "", fmt.Errorf("%w", err)
	}

	token, err := tokens_helper.CreateNewJwt(userId, email, a.tokenTtl, a.tokenSecret)
	if err != nil {
		log.Error("ошибка при создании jwt: ", err)
		return "", fmt.Errorf("%w", err)
	}

	return token, nil
}

func (a *AuthService) LoginUser(ctx context.Context, email string, password string) (string, error) {

	op := "authService/loginUser"

	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	user, err := a.userGetter.GetUser(ctx, email)
	if err != nil {
		log.Error("ошибка при получении пользователя: ", err)
		return "", fmt.Errorf("%w", err)
	}

	err = bcrypt.CompareHashAndPassword(user.PassHash, []byte(password))
	if err != nil {
		log.Error("пароли не совпадают: ", err)
		return "", fmt.Errorf("%w", err)
	}

	token, err := tokens_helper.CreateNewJwt(user.Id, user.Email, a.tokenTtl, a.tokenSecret)
	if err != nil {
		log.Error("ошибка при генерации токена: ", err)
		return "", fmt.Errorf("%w", err)
	}

	return token, nil
}

func (a *AuthService) CheckAuth(ctx context.Context, token string) (string, error) {

	op := "authService/checkAuth"

	log := a.log.With(
		slog.String("op", op),
	)

	parsedToken, err := tokens_helper.Parse(token, a.tokenSecret)

	if err != nil {
		log.Error("ошибка при парсинге токена: ", err)
		return "", fmt.Errorf("%w", err)
	}

	if !parsedToken.Valid {
		log.Error("токен не валиден: ", err)
		return "", fmt.Errorf("%w", err)
	}

	tokenClaims, err := tokens_helper.GetClaimsValues(parsedToken)
	if err != nil {
		log.Error("ошибка при парсинге claims: ", err)
		return "", fmt.Errorf("%w", err)
	}

	newToken, err := tokens_helper.CreateNewJwt(tokenClaims.Id, tokenClaims.Email, a.tokenTtl, a.tokenSecret)
	if err != nil {
		log.Error("ошибка при создании нового токена: ", err)
		return "", fmt.Errorf("%w", err)
	}

	return newToken, nil
}