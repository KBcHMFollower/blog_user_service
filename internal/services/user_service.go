package services

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/KBcHMFollower/blog_user_service/internal/clients/s3"
	repositories_transfer "github.com/KBcHMFollower/blog_user_service/internal/domain/layers_TOs/repositories"
	transfer "github.com/KBcHMFollower/blog_user_service/internal/domain/layers_TOs/services"
	"github.com/KBcHMFollower/blog_user_service/internal/lib/tokens"
	dep "github.com/KBcHMFollower/blog_user_service/internal/services/interfaces/dep"

	"log/slog"
	"time"

	"github.com/KBcHMFollower/blog_user_service/internal/domain/models"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type EventStore interface {
	dep.EventGetter
}

type ImageStore interface {
	dep.ImageGetter
	dep.ImageUploader
}

type UserStore interface {
	dep.UserDeleter
	dep.UserGetter
	dep.UserUpdater
	dep.SubscribeManager
	dep.UserCreator
}

type UserService struct {
	log         *slog.Logger
	tokenTtl    time.Duration
	tokenSecret string
	userRep     UserStore
	eventsRep   EventStore
	imgStore    ImageStore
}

func NewUserService(log *slog.Logger, tokenTtl time.Duration, tokenSecret string, userRep UserStore, eventsRep EventStore, imgStore ImageStore) *UserService {
	return &UserService{
		log:         log,
		tokenTtl:    tokenTtl,
		tokenSecret: tokenSecret,
		userRep:     userRep,
		imgStore:    imgStore,
		eventsRep:   eventsRep,
	}
}

func (a *UserService) RegisterUser(ctx context.Context, req *transfer.RegisterInfo) (*transfer.TokenResult, error) {

	op := "authService/registerUser"

	log := a.log.With(
		slog.String("op", op),
		slog.String("email", req.Email),
	)

	hashPass, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("can`t generate hashPass: ", err)
		return nil, fmt.Errorf("%s : %w", op, err)
	}

	userId, err := a.userRep.CreateUser(ctx, &repositories_transfer.CreateUserInfo{
		Email:    req.Email,
		FName:    req.FName,
		LName:    req.LName,
		HashPass: hashPass,
	})
	if err != nil {
		log.Error("can`t create user in db: ", err)
		return nil, fmt.Errorf("%s : %w", op, err)
	}

	token, err := tokens_helper.CreateNewJwt(userId, req.Email, a.tokenTtl, a.tokenSecret)
	if err != nil {
		log.Error("can`t create jwt: ", err)
		return nil, fmt.Errorf("%s : %w", op, err)
	}

	return &transfer.TokenResult{
		AccessToken: token,
	}, nil
}

func (a *UserService) LoginUser(ctx context.Context, loginInfo *transfer.LoginInfo) (*transfer.TokenResult, error) {

	op := "authService/loginUser"

	log := a.log.With(
		slog.String("op", op),
		slog.String("email", loginInfo.Email),
	)

	user, err := a.userRep.GetUserByEmail(ctx, loginInfo.Email)
	if err != nil {
		log.Error("can`t get user from db: ", err)
		return nil, fmt.Errorf("%s : %w", op, err)
	}

	err = bcrypt.CompareHashAndPassword(user.PassHash, []byte(loginInfo.Password))
	if err != nil {
		log.Error("passwords not eq: ", err)
		return nil, fmt.Errorf("%s : %w", op, err)
	}

	token, err := tokens_helper.CreateNewJwt(user.Id, user.Email, a.tokenTtl, a.tokenSecret)
	if err != nil {
		log.Error("can`t generate jwt: ", err)
		return nil, fmt.Errorf("%s : %w", op, err)
	}

	return &transfer.TokenResult{
		AccessToken: token,
	}, nil
}

func (a *UserService) CheckAuth(ctx context.Context, authInfo *transfer.CheckAuthInfo) (*transfer.TokenResult, error) {

	op := "authService/checkAuth"

	log := a.log.With(
		slog.String("op", op),
	)

	parsedToken, err := tokens_helper.Parse(authInfo.AccessToken, a.tokenSecret)
	if err != nil {
		log.Error("can`t parse token: ", err)
		return nil, fmt.Errorf("%s : %w", op, err)
	}
	if !parsedToken.Valid {
		log.Error("token is invalid: ", err)
		return nil, fmt.Errorf("%s : %w", op, err)
	}

	tokenClaims, err := tokens_helper.GetClaimsValues(parsedToken)
	if err != nil {
		log.Error("can`t  parse jwt claims: ", err)
		return nil, fmt.Errorf("%s : %w", op, err)
	}

	newToken, err := tokens_helper.CreateNewJwt(tokenClaims.Id, tokenClaims.Email, a.tokenTtl, a.tokenSecret)
	if err != nil {
		log.Error("can`t create jwt: ", err)
		return nil, fmt.Errorf("%s : %w", op, err)
	}

	return &transfer.TokenResult{
		AccessToken: newToken,
	}, nil
}

func (a *UserService) GetUserById(ctx context.Context, userId uuid.UUID) (*transfer.GetUserResult, error) {
	op := "UserService/GetUserById"

	log := a.log.With(
		slog.String("op", op),
	)

	user, err := a.userRep.GetUserById(ctx, userId)
	if err != nil {
		log.Error("can`t get user from db: ", err)
		return nil, fmt.Errorf("%s : %w", op, err)
	}

	return &transfer.GetUserResult{
		User: transfer.GetUserResultFromModel(user),
	}, nil
}

func (a *UserService) GetSubscribers(ctx context.Context, getInfo *transfer.GetSubscribersInfo) (*transfer.GetSubscribersResult, error) {
	op := "UserService/GetSubscribers"
	log := a.log.With(
		slog.String("op", op))

	users, totalCount, err := a.userRep.GetUserSubscribers(ctx, repositories_transfer.GetUserSubscribersInfo{
		UserId: getInfo.BloggerId,
		Page:   uint64(getInfo.Page),
		Size:   uint64(getInfo.Size)})
	if err != nil {
		log.Error("can`t get user subscribers from db: ", err)
		return nil, fmt.Errorf("%s : %w", op, err)
	}

	return &transfer.GetSubscribersResult{
		Subscribers: transfer.GetSubscribersArrayResultFromModel(users),
		TotalCount:  int32(totalCount),
	}, nil
}

func (a *UserService) GetSubscriptions(ctx context.Context, getInfo *transfer.GetSubscriptionsInfo) (*transfer.GetSubscriptionsResult, error) {
	op := "UserService/GetSubscriptions"
	log := a.log.With(
		slog.String("op", op))

	users, totalCount, err := a.userRep.GetUserSubscriptions(ctx, repositories_transfer.GetUserSubscriptionsInfo{
		UserId: getInfo.SubscriberId,
		Page:   uint64(getInfo.Page),
		Size:   uint64(getInfo.Size),
	})
	if err != nil {
		log.Error("can`t get user bloggers from db: ", err)
		return nil, fmt.Errorf("%s : %w", op, err)
	}

	return &transfer.GetSubscriptionsResult{
		Subscriptions: transfer.GetSubscribersArrayResultFromModel(users),
		TotalCount:    int32(totalCount),
	}, nil
}

func (a *UserService) UpdateUser(ctx context.Context, updateInfo *transfer.UpdateUserInfo) (*transfer.UpdateUserResult, error) {
	op := "UserService/UpdateUser"
	log := a.log.With(
		slog.String("op", op))

	var updateItems = make([]*repositories_transfer.UserFieldInfo, 0)
	for _, fieldInfo := range updateInfo.UpdateFields {
		updateItems = append(updateItems, &repositories_transfer.UserFieldInfo{
			Name:  fieldInfo.Name,
			Value: fieldInfo.Value,
		})
	}

	user, err := a.userRep.UpdateUser(ctx, repositories_transfer.UpdateUserInfo{
		Id:         updateInfo.Id,
		UpdateInfo: updateItems,
	})
	if err != nil {
		log.Error("can`t update user in db: ", err)
		return nil, fmt.Errorf("%s : %w", op, err)
	}

	return &transfer.UpdateUserResult{
		User: transfer.GetUserResultFromModel(user),
	}, nil
}

func (a *UserService) Subscribe(ctx context.Context, subInfo *transfer.SubscribeInfo) error {
	op := "UserService/Subscribe"
	log := a.log.With(
		slog.String("op", op))

	err := a.userRep.Subscribe(ctx, repositories_transfer.SubscribeToUserInfo{
		BloggerId:    subInfo.BloggerId,
		SubscriberId: subInfo.SubscriberId,
	})
	if err != nil {
		log.Error("can`t subscribe to user in db: ", err)
		return fmt.Errorf("%s : %w", op, err)
	}

	return nil
}

func (a *UserService) Unsubscribe(ctx context.Context, subInfo *transfer.SubscribeInfo) error {
	op := "UserService/Subscribe"
	log := a.log.With(
		slog.String("op", op))

	err := a.userRep.Unsubscribe(ctx, repositories_transfer.UnsubscribeInfo{
		BloggerId:    subInfo.BloggerId,
		SubscriberId: subInfo.SubscriberId,
	})
	if err != nil {
		log.Error("can`t unsubscribe in db: ", err)
		return fmt.Errorf("%s : %w", op, err)
	}

	return nil
}

func (a *UserService) DeleteUser(ctx context.Context, deleteInfo *transfer.DeleteUserInfo) error {
	op := "UserService/DeleteUser"
	log := a.log.With(
		slog.String("op", op))

	err := a.userRep.DeleteUser(ctx, repositories_transfer.DeleteUserInfo{
		Id: deleteInfo.Id,
	})
	if err != nil {
		log.Error("can`t delete user in db: ", err)
		return fmt.Errorf("%s : %w", op, err)
	}

	return nil
}

func (a *UserService) UploadAvatar(ctx context.Context, uploadInfo *transfer.UploadAvatarInfo) (*transfer.AvatarResult, error) {
	op := "UserService/UploadAvatar"
	log := a.log.With(
		slog.String("op", op))

	imgUrl, err := a.imgStore.UploadFile(ctx, fmt.Sprintf("%s.jpeg", uuid.New().String()), uploadInfo.Image, s3client.ImageJpeg)
	if err != nil {
		log.Error("can`t upload image: ", err)
		return nil, fmt.Errorf("%s : %w", op, err)
	}

	//TODO
	_, err = a.userRep.UpdateUser(ctx, repositories_transfer.UpdateUserInfo{
		Id: uploadInfo.UserId,
		UpdateInfo: []*repositories_transfer.UserFieldInfo{
			{
				Name:  "avatar",
				Value: imgUrl,
			},
			{
				Name:  "avatar_min",
				Value: imgUrl,
			},
		},
	})
	if err != nil {
		log.Error("can`t update user in db: ", err)
		return nil, fmt.Errorf("%s : %w", op, err)
	}

	return &transfer.AvatarResult{
		UserId:     uploadInfo.UserId,
		Avatar:     imgUrl,
		AvatarMini: imgUrl,
	}, nil
}

func (a *UserService) CompensateDeletedUser(ctx context.Context, eventId uuid.UUID) error {
	op := "UserService/CompensateDeletedUser"
	log := a.log.With(
		slog.String("op", op))

	eventInfo, err := a.eventsRep.GetEventById(ctx, eventId)
	if err != nil {
		log.Error("can`t get event info: ", err)
		return fmt.Errorf("%s : %w", op, err)
	}

	var user models.User
	if err := json.Unmarshal([]byte(eventInfo.Payload), &user); err != nil {
		log.Error("can`t unmarshal event: ", err)
		return fmt.Errorf("%s : %w", op, err)
	}

	_, err = a.userRep.CreateUser(ctx, &repositories_transfer.CreateUserInfo{
		Email:    user.Email,
		FName:    user.FName,
		LName:    user.LName,
		HashPass: user.PassHash,
	}) //TODO : Create new metod from user compensate
	if err != nil {
		log.Error("can`t create user in db: ", err)
		return fmt.Errorf("%s : %w", op, err)
	}

	return nil
}
