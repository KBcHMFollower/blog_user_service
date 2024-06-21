package services

import (
	"context"
	"fmt"
	ssov1 "github.com/KBcHMFollower/auth-service/api/protos/gen/auth"
	"github.com/KBcHMFollower/auth-service/internal/lib/tokens"
	"github.com/KBcHMFollower/auth-service/internal/repository"
	"log/slog"
	"time"

	"github.com/KBcHMFollower/auth-service/internal/domain/models"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserSaver interface {
	CreateUser(ctx context.Context, createDto *repository.CreateUserDto) (uuid.UUID, error)
}

type UserGetter interface {
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
}

type UserService struct {
	log         *slog.Logger
	tokenTtl    time.Duration
	tokenSecret string
	userSaver   UserSaver
	userGetter  UserGetter
}

func New(log *slog.Logger, tokenTtl time.Duration, tokenSecret string, userSaver UserSaver, userGetter UserGetter) *UserService {
	return &UserService{
		log:         log,
		tokenTtl:    tokenTtl,
		tokenSecret: tokenSecret,
		userSaver:   userSaver,
		userGetter:  userGetter,
	}
}

func (a *UserService) RegisterUser(ctx context.Context, req *ssov1.RegisterDTO) (string, error) {

	op := "authService/registerUser"

	log := a.log.With(
		slog.String("op", op),
		slog.String("email", req.Email),
	)

	hashPass, err := bcrypt.GenerateFromPassword([]byte(req.GetPassword()), bcrypt.DefaultCost)
	if err != nil {
		log.Error("can`t generate hashPass: ", err)
		return "", fmt.Errorf("%w", err)
	}

	userId, err := a.userSaver.CreateUser(ctx, &repository.CreateUserDto{
		Email:    req.GetEmail(),
		Fname:    req.GetFname(),
		Lname:    req.GetLname(),
		HashPass: hashPass,
	})
	if err != nil {
		log.Error("can`t create user in db: ", err)
		return "", fmt.Errorf("%w", err)
	}

	token, err := tokens_helper.CreateNewJwt(userId, req.GetEmail(), a.tokenTtl, a.tokenSecret)
	if err != nil {
		log.Error("can`t create jwt: ", err)
		return "", fmt.Errorf("%w", err)
	}

	return token, nil
}

func (a *UserService) LoginUser(ctx context.Context, email string, password string) (string, error) {

	op := "authService/loginUser"

	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	user, err := a.userGetter.GetUserByEmail(ctx, email)
	if err != nil {
		log.Error("can`t get user from db: ", err)
		return "", fmt.Errorf("%w", err)
	}

	err = bcrypt.CompareHashAndPassword(user.PassHash, []byte(password))
	if err != nil {
		log.Error("passwords not eq: ", err)
		return "", fmt.Errorf("%w", err)
	}

	token, err := tokens_helper.CreateNewJwt(user.Id, user.Email, a.tokenTtl, a.tokenSecret)
	if err != nil {
		log.Error("can`t generate jwt: ", err)
		return "", fmt.Errorf("%w", err)
	}

	return token, nil
}

func (a *UserService) CheckAuth(ctx context.Context, token string) (string, error) {

	op := "authService/checkAuth"

	log := a.log.With(
		slog.String("op", op),
	)

	parsedToken, err := tokens_helper.Parse(token, a.tokenSecret)
	if err != nil {
		log.Error("can`t parse token: ", err)
		return "", fmt.Errorf("%w", err)
	}
	if !parsedToken.Valid {
		log.Error("token is invalid: ", err)
		return "", fmt.Errorf("%w", err)
	}

	tokenClaims, err := tokens_helper.GetClaimsValues(parsedToken)
	if err != nil {
		log.Error("can`t  parse jwt claims: ", err)
		return "", fmt.Errorf("%w", err)
	}

	newToken, err := tokens_helper.CreateNewJwt(tokenClaims.Id, tokenClaims.Email, a.tokenTtl, a.tokenSecret)
	if err != nil {
		log.Error("can`t create jwt: ", err)
		return "", fmt.Errorf("%w", err)
	}

	return newToken, nil
}
