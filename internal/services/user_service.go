package services

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/KBcHMFollower/blog_user_service/internal/clients/amqpclient"
	"github.com/KBcHMFollower/blog_user_service/internal/clients/amqpclient/messages"
	"github.com/KBcHMFollower/blog_user_service/internal/clients/s3"
	"github.com/KBcHMFollower/blog_user_service/internal/domain"
	repositories_transfer "github.com/KBcHMFollower/blog_user_service/internal/domain/layers_TOs/repositories"
	transfer "github.com/KBcHMFollower/blog_user_service/internal/domain/layers_TOs/services"
	"github.com/KBcHMFollower/blog_user_service/internal/domain/models"
	"github.com/KBcHMFollower/blog_user_service/internal/lib/tokens"
	dep "github.com/KBcHMFollower/blog_user_service/internal/services/interfaces/dep"
	services_utils "github.com/KBcHMFollower/blog_user_service/internal/services/lib"

	"log/slog"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type EventStore interface {
	dep.EventCreator
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
	txCreator   dep.TransactionCreator
	imgStore    ImageStore
}

func NewUserService(log *slog.Logger, tokenTtl time.Duration, tokenSecret string, txCreator dep.TransactionCreator, userRep UserStore, eventsRep EventStore, imgStore ImageStore) *UserService {
	return &UserService{
		log:         log,
		tokenTtl:    tokenTtl,
		tokenSecret: tokenSecret,
		userRep:     userRep,
		imgStore:    imgStore,
		txCreator:   txCreator,
		eventsRep:   eventsRep,
	}
}

// TODO: fname и lname не работают
func (a *UserService) RegisterUser(ctx context.Context, req *transfer.RegisterInfo) (*transfer.TokenResult, error) {

	op := "UserService/registerUser"

	log := a.log.With(
		slog.String("op", op),
		slog.String("email", req.Email),
	)

	hashPass, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("can`t generate hashPass: ", err)
		return nil, domain.AddOpInErr(err, op)
	}

	userId, err := a.userRep.CreateUser(ctx, &repositories_transfer.CreateUserInfo{
		Email:    req.Email,
		FName:    req.FName,
		LName:    req.LName,
		HashPass: hashPass,
	})
	if err != nil {
		log.Error("can`t create user in db: ", err)
		return nil, domain.AddOpInErr(err, op)
	}

	token, err := tokens_helper.CreateNewJwt(userId, req.Email, a.tokenTtl, a.tokenSecret)
	if err != nil {
		log.Error("can`t create jwt: ", err)
		return nil, domain.AddOpInErr(err, op)
	}

	return &transfer.TokenResult{
		AccessToken: token,
	}, nil
}

func (a *UserService) LoginUser(ctx context.Context, loginInfo *transfer.LoginInfo) (*transfer.TokenResult, error) {

	op := "UserService/loginUser"

	log := a.log.With(
		slog.String("op", op),
		slog.String("email", loginInfo.Email),
	)

	user, err := a.userRep.GetUserByEmail(ctx, loginInfo.Email)
	if err != nil {
		log.Error("can`t get user from db: ", err)
		return nil, domain.AddOpInErr(err, op)
	}

	err = bcrypt.CompareHashAndPassword(user.PassHash, []byte(loginInfo.Password))
	if err != nil {
		log.Error("passwords not eq: ", err)
		return nil, domain.AddOpInErr(err, op)
	}

	token, err := tokens_helper.CreateNewJwt(user.Id, user.Email, a.tokenTtl, a.tokenSecret)
	if err != nil {
		log.Error("can`t generate jwt: ", err)
		return nil, domain.AddOpInErr(err, op)
	}

	return &transfer.TokenResult{
		AccessToken: token,
	}, nil
}

func (a *UserService) CheckAuth(ctx context.Context, authInfo *transfer.CheckAuthInfo) (*transfer.TokenResult, error) {

	op := "UserService/checkAuth"

	log := a.log.With(
		slog.String("op", op),
	)

	parsedToken, err := tokens_helper.Parse(authInfo.AccessToken, a.tokenSecret)
	if err != nil {
		log.Error("can`t parse token: ", err)
		return nil, domain.AddOpInErr(err, op)
	}
	if !parsedToken.Valid {
		log.Error("token is invalid: ", err)
		return nil, domain.AddOpInErr(err, op)
	}

	tokenClaims, err := tokens_helper.GetClaimsValues(parsedToken)
	if err != nil {
		log.Error("can`t  parse jwt claims: ", err)
		return nil, domain.AddOpInErr(err, op)
	}

	newToken, err := tokens_helper.CreateNewJwt(tokenClaims.Id, tokenClaims.Email, a.tokenTtl, a.tokenSecret)
	if err != nil {
		log.Error("can`t create jwt: ", err)
		return nil, domain.AddOpInErr(err, op)
	}

	return &transfer.TokenResult{
		AccessToken: newToken,
	}, nil
}

func (a *UserService) GetUserById(ctx context.Context, userId uuid.UUID) (ersUser *transfer.GetUserResult, resErr error) {
	op := "UserService/GetUserById"

	log := a.log.With(
		slog.String("op", op),
	)

	tx, err := a.txCreator.BeginTxCtx(ctx, nil)
	if err != nil {
		log.Error("can`t begin tx: ", err)
		return nil, domain.AddOpInErr(err, op)
	}
	defer func() {
		resErr = services_utils.HandleErrInTransaction(resErr, tx)
	}()

	cacheUser, err := a.userRep.TryGetUserFromCache(ctx, userId)
	if err != nil {
		log.Error("can`t get cacheUser from cache: ", err)
	}
	if cacheUser != nil {
		return &transfer.GetUserResult{
			User: transfer.GetUserResultFromModel(cacheUser),
		}, nil
	}

	user, err := a.userRep.GetUserById(ctx, userId, tx)
	if err != nil {
		log.Error("can`t get cacheUser from db: ", err)
		return nil, domain.AddOpInErr(err, op)
	}

	if err := a.userRep.SetUserToCache(ctx, user); err != nil {
		log.Error("can`t set cacheUser to db: ", err)
		return nil, domain.AddOpInErr(err, op)
	}

	if err := tx.Commit(); err != nil {
		log.Error("can`t commit tx: ", err)
		return nil, domain.AddOpInErr(err, op)
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
		return nil, domain.AddOpInErr(err, op)
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
		return nil, domain.AddOpInErr(err, op)
	}

	return &transfer.GetSubscriptionsResult{
		Subscriptions: transfer.GetSubscribersArrayResultFromModel(users),
		TotalCount:    int32(totalCount),
	}, nil
}

func (a *UserService) UpdateUser(ctx context.Context, updateInfo *transfer.UpdateUserInfo) (resUser *transfer.UpdateUserResult, resErr error) {
	op := "UserService/UpdateUser"
	log := a.log.With(
		slog.String("op", op))

	tx, err := a.txCreator.BeginTxCtx(ctx, nil)
	if err != nil {
		log.Error("can`t begin tx: ", err)
		return nil, domain.AddOpInErr(err, op)
	}
	defer func() {
		resErr = services_utils.HandleErrInTransaction(resErr, tx)
	}()

	var updateItems = make([]*repositories_transfer.UserFieldInfo, 0)
	for _, fieldInfo := range updateInfo.UpdateFields {
		updateItems = append(updateItems, &repositories_transfer.UserFieldInfo{
			Name:  fieldInfo.Name,
			Value: fieldInfo.Value,
		})
	}

	if err := a.userRep.UpdateUser(ctx, repositories_transfer.UpdateUserInfo{
		Id:         updateInfo.Id,
		UpdateInfo: updateItems,
	}, tx); err != nil {
		log.Error("can`t update user in db: ", err)
		return nil, domain.AddOpInErr(err, op)
	}

	exUser, exErr := a.userRep.GetUserById(ctx, updateInfo.Id, tx)
	if exErr != nil {
		log.Error("can`t get user from db: ", err)
		return nil, domain.AddOpInErr(err, op)
	}

	if err := a.userRep.DeleteUserFromCache(ctx, updateInfo.Id); err != nil {
		log.Error("can`t update user in db: ", err)
		return nil, domain.AddOpInErr(err, op)
	}

	if err := tx.Commit(); err != nil {
		log.Error("can`t commit tx: ", err)
		return nil, domain.AddOpInErr(err, op)
	}

	return &transfer.UpdateUserResult{
		User: transfer.GetUserResultFromModel(exUser),
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
		return domain.AddOpInErr(err, op)
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
		return domain.AddOpInErr(err, op)
	}

	return nil
}

func (a *UserService) DeleteUser(ctx context.Context, deleteInfo *transfer.DeleteUserInfo) (resErr error) {
	op := "UserService/DeleteUser"
	log := a.log.With(
		slog.String("op", op))

	tx, err := a.txCreator.BeginTxCtx(ctx, nil)
	if err != nil {
		log.Error("can`t begin transaction: ", resErr)
		return fmt.Errorf("%s : %w", op, resErr)
	}
	defer func() {
		resErr = services_utils.HandleErrInTransaction(resErr, tx)
	}()

	user, err := a.userRep.GetUserById(ctx, deleteInfo.Id, tx)
	if err != nil {
		log.Error("can`t get user by id: ", resErr)
		return domain.AddOpInErr(err, op)
	}

	if err := a.userRep.DeleteUser(ctx, repositories_transfer.DeleteUserInfo{
		Id: deleteInfo.Id,
	}, tx); err != nil {
		log.Error("can`t delete user in db: ", err)
		return domain.AddOpInErr(err, op)
	}

	if err := a.userRep.DeleteUserFromCache(ctx, deleteInfo.Id); err != nil {
		log.Error("can`t delete user from cache: ", err)
		return domain.AddOpInErr(err, op)
	}

	eventId := uuid.New()

	messageEntity := messages.UserDeletedMessage{
		User: messages.UserMessage{
			FName:       user.FName,
			LName:       user.LName,
			Email:       user.Email,
			PassHash:    user.PassHash,
			Id:          user.Id,
			Avatar:      user.Avatar,
			AvatarMin:   user.AvatarMin,
			CreatedDate: user.CreatedDate,
			UpdatedDate: user.UpdatedDate,
		},
		EventId: eventId,
	}

	messageJson, err := json.Marshal(messageEntity)
	if err != nil {
		log.Error("can`t marshal message: ", err)
		return domain.AddOpInErr(err, op)
	}

	if err := a.eventsRep.Create(ctx, repositories_transfer.CreateEventInfo{
		EventId:   eventId,
		EventType: amqpclient.UserDeletedEventKey,
		Payload:   messageJson,
	}, tx); err != nil {
		log.Error("can`t create event in db: ", err)
		return domain.AddOpInErr(err, op)
	}

	if err := tx.Commit(); err != nil {
		log.Error("can`t commit transaction: ", err)
		return domain.AddOpInErr(err, op)
	}

	return nil
}

func (a *UserService) UploadAvatar(ctx context.Context, uploadInfo *transfer.UploadAvatarInfo) (resAvatar *transfer.AvatarResult, resErr error) {
	op := "UserService/UploadAvatar"

	log := a.log.With(
		slog.String("op", op))

	tx, err := a.txCreator.BeginTxCtx(ctx, nil)
	if err != nil {
		log.Error("can`t begin tx: ", err)
		return nil, domain.AddOpInErr(err, op)
	}
	defer func() {
		resErr = services_utils.HandleErrInTransaction(resErr, tx)
	}()

	imgUrl, err := a.imgStore.UploadFile(ctx, fmt.Sprintf("%s.jpeg", uuid.New().String()), uploadInfo.Image, s3client.ImageJpeg)
	if err != nil {
		log.Error("can`t upload image: ", err)
		return nil, domain.AddOpInErr(err, op)
	}

	//TODO
	err = a.userRep.UpdateUser(ctx, repositories_transfer.UpdateUserInfo{
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
	}, nil)
	if err != nil {
		log.Error("can`t update user in db: ", err)
		return nil, domain.AddOpInErr(err, op)
	}

	if err := a.userRep.DeleteUserFromCache(ctx, uploadInfo.UserId); err != nil {
		log.Error("can`t update user in db: ", err)
		return nil, domain.AddOpInErr(err, op)
	}

	if err := tx.Commit(); err != nil {
		log.Error("can`t commit tx: ", err)
		return nil, domain.AddOpInErr(err, op)
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
		return domain.AddOpInErr(err, op)
	}

	var message messages.UserDeletedMessage
	if err := json.Unmarshal([]byte(eventInfo.Payload), &message); err != nil {
		log.Error("can`t unmarshal event: ", err)
		return domain.AddOpInErr(err, op)
	}

	err = a.userRep.RollBackUser(ctx, models.User{
		Id:          message.User.Id,
		Email:       message.User.Email,
		PassHash:    message.User.PassHash,
		FName:       message.User.FName,
		LName:       message.User.LName,
		CreatedDate: message.User.CreatedDate,
		UpdatedDate: time.Now(),
		Avatar:      message.User.Avatar,
		AvatarMin:   message.User.AvatarMin,
	})
	if err != nil {
		log.Error("can`t create user in db: ", err)
		return domain.AddOpInErr(err, op)
	}

	return nil
}
