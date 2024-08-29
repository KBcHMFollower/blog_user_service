package services

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/KBcHMFollower/blog_user_service/internal/clients/amqpclient"
	"github.com/KBcHMFollower/blog_user_service/internal/clients/amqpclient/messages"
	s3client "github.com/KBcHMFollower/blog_user_service/internal/clients/s3"
	ctxerrors "github.com/KBcHMFollower/blog_user_service/internal/domain/errors"
	repositories_transfer "github.com/KBcHMFollower/blog_user_service/internal/domain/layers_TOs/repositories"
	transfer "github.com/KBcHMFollower/blog_user_service/internal/domain/layers_TOs/services"
	"github.com/KBcHMFollower/blog_user_service/internal/domain/models"
	tokens_helper "github.com/KBcHMFollower/blog_user_service/internal/lib/tokens"
	"github.com/KBcHMFollower/blog_user_service/internal/logger"
	dep "github.com/KBcHMFollower/blog_user_service/internal/services/interfaces/dep"
	services_utils "github.com/KBcHMFollower/blog_user_service/internal/services/lib"
	"log/slog"

	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const (
	createdUserIdLogKey = "created-user-id"
	accessTokenLogKey   = "access-token"
	bloggerIdLogKey     = "blogger-id"
	subscriberIdLogKey  = "subscriber-id"
	updateInfoLogKey    = "update-info"
	avatarUruLogKey     = "avatar-uru"
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

// TODO: fname и lname не работают, должна быть транзакция
func (a *UserService) RegisterUser(ctx context.Context, req *transfer.RegisterInfo) (*transfer.TokenResult, error) {
	logger.UpdateLoggerCtx(ctx, logger.ActionEmailKey, req.Email)

	a.log.InfoContext(ctx, "trying to register user")

	hashPass, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t generate hashPass", err))
	}

	a.log.Debug("hash pass is generated successfully")

	userId, err := a.userRep.CreateUser(ctx, &repositories_transfer.CreateUserInfo{
		Email:    req.Email,
		FName:    req.FName,
		LName:    req.LName,
		HashPass: hashPass,
	})
	if err != nil {
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t create user in db", err))
	}

	logger.UpdateLoggerCtx(ctx, createdUserIdLogKey, userId)
	a.log.Debug("user created in db successfully")

	token, err := tokens_helper.CreateNewJwt(userId, req.Email, a.tokenTtl, a.tokenSecret)
	if err != nil {
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t create new jwt", err))
	}

	logger.UpdateLoggerCtx(ctx, accessTokenLogKey, token)
	a.log.Debug("token created successfully")

	a.log.InfoContext(ctx, "user registered successfully")

	return &transfer.TokenResult{
		AccessToken: token,
	}, nil
}

func (a *UserService) LoginUser(ctx context.Context, loginInfo *transfer.LoginInfo) (*transfer.TokenResult, error) {
	logger.UpdateLoggerCtx(ctx, logger.ActionEmailKey, loginInfo.Email)

	a.log.InfoContext(ctx, "user try to login")

	user, err := a.userRep.GetUserByEmail(ctx, loginInfo.Email)
	if err != nil {
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t get user from db", err))
	}

	logger.UpdateLoggerCtx(ctx, logger.ActionUserIdKey, user.Id)
	a.log.Debug("email is exists")

	err = bcrypt.CompareHashAndPassword(user.PassHash, []byte(loginInfo.Password))
	if err != nil {
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("passwords not eq", err))
	}

	a.log.Debug("password is correct")

	token, err := tokens_helper.CreateNewJwt(user.Id, user.Email, a.tokenTtl, a.tokenSecret)
	if err != nil {
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t generate jwt", err))
	}

	a.log.Debug("token created successfully", "as-token", token)
	a.log.InfoContext(ctx, "user logged in successfully")

	return &transfer.TokenResult{
		AccessToken: token,
	}, nil
}

func (a *UserService) CheckAuth(ctx context.Context, authInfo *transfer.CheckAuthInfo) (*transfer.TokenResult, error) {
	parsedToken, err := tokens_helper.Parse(authInfo.AccessToken, a.tokenSecret)

	if err != nil {
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t parse token", err))
	}
	if !parsedToken.Valid {
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("token is invalid", err))
	}

	tokenClaims, err := tokens_helper.GetClaimsValues(parsedToken)
	if err != nil {
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t  parse jwt claims", err))
	}

	newToken, err := tokens_helper.CreateNewJwt(tokenClaims.Id, tokenClaims.Email, a.tokenTtl, a.tokenSecret)
	if err != nil {
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t create jwt", err))
	}

	return &transfer.TokenResult{
		AccessToken: newToken,
	}, nil
}

func (a *UserService) GetUserById(ctx context.Context, userId uuid.UUID) (ersUser *transfer.GetUserResult, resErr error) {
	logger.UpdateLoggerCtx(ctx, logger.ActionUserIdKey, userId)

	a.log.Debug("try to get user by id")

	tx, err := a.txCreator.BeginTxCtx(ctx, nil)
	if err != nil {
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t begin tx", err))
	}
	defer func() {
		resErr = services_utils.HandleErrInTransaction(resErr, tx)
	}()

	cacheUser, err := a.userRep.TryGetUserFromCache(ctx, userId)
	if err != nil {
		a.log.Debug("can`t get cacheUser from cache: ", "err", err.Error())
	}
	if cacheUser != nil {
		a.log.Debug("user found in cache")
		return &transfer.GetUserResult{
			User: transfer.GetUserResultFromModel(cacheUser),
		}, nil
	}

	a.log.Debug("try to get user by id from db")

	user, err := a.userRep.GetUserById(ctx, userId, tx)
	if err != nil {
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t get cacheUser from db", err))
	}

	a.log.Debug("user found in db")

	if err := a.userRep.SetUserToCache(ctx, user); err != nil {
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t set cacheUser to db", err))
	}
	a.log.Debug("user added to cache")

	if err := tx.Commit(); err != nil {
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t commit tx", err))
	}

	a.log.Debug("get user by id successfully")

	return &transfer.GetUserResult{
		User: transfer.GetUserResultFromModel(user),
	}, nil
}

func (a *UserService) GetSubscribers(ctx context.Context, getInfo *transfer.GetSubscribersInfo) (*transfer.GetSubscribersResult, error) {
	logger.UpdateLoggerCtx(ctx, bloggerIdLogKey, getInfo.BloggerId)

	a.log.Debug("try to get subscribers")

	users, totalCount, err := a.userRep.GetUserSubscribers(ctx, repositories_transfer.GetUserSubscribersInfo{
		UserId: getInfo.BloggerId,
		Page:   uint64(getInfo.Page),
		Size:   uint64(getInfo.Size)})
	if err != nil {
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t get user subscribers from db", err))
	}

	a.log.Debug("subscribers found in db")

	return &transfer.GetSubscribersResult{
		Subscribers: transfer.GetSubscribersArrayResultFromModel(users),
		TotalCount:  int32(totalCount),
	}, nil
}

func (a *UserService) GetSubscriptions(ctx context.Context, getInfo *transfer.GetSubscriptionsInfo) (*transfer.GetSubscriptionsResult, error) {
	a.log.Debug("try to get subscriptions")

	users, totalCount, err := a.userRep.GetUserSubscriptions(ctx, repositories_transfer.GetUserSubscriptionsInfo{
		UserId: getInfo.SubscriberId,
		Page:   uint64(getInfo.Page),
		Size:   uint64(getInfo.Size),
	})
	if err != nil {
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t get user subscriptions from db", err))
	}

	a.log.Debug("subscribers found in db")

	return &transfer.GetSubscriptionsResult{
		Subscriptions: transfer.GetSubscribersArrayResultFromModel(users),
		TotalCount:    int32(totalCount),
	}, nil
}

func (a *UserService) UpdateUser(ctx context.Context, updateInfo *transfer.UpdateUserInfo) (resUser *transfer.UpdateUserResult, resErr error) {
	logger.UpdateLoggerCtx(ctx, logger.ActionUserIdKey, updateInfo.Id)
	logger.UpdateLoggerCtx(ctx, updateInfoLogKey, updateInfo.UpdateFields)

	a.log.InfoContext(ctx, "try to update user")

	tx, err := a.txCreator.BeginTxCtx(ctx, nil)
	if err != nil {
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t begin tx", err))
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
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t update user in db", err))
	}

	exUser, exErr := a.userRep.GetUserById(ctx, updateInfo.Id, tx)
	if exErr != nil {
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t get user from db", err))
	}

	if err := a.userRep.DeleteUserFromCache(ctx, updateInfo.Id); err != nil {
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t update user in db", err))
	}

	if err := tx.Commit(); err != nil {
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t commit tx", err))
	}

	a.log.InfoContext(ctx, "user updated successfully")

	return &transfer.UpdateUserResult{
		User: transfer.GetUserResultFromModel(exUser),
	}, nil
}

func (a *UserService) Subscribe(ctx context.Context, subInfo *transfer.SubscribeInfo) error {
	logger.UpdateLoggerCtx(ctx, subscriberIdLogKey, subInfo.SubscriberId)
	logger.UpdateLoggerCtx(ctx, bloggerIdLogKey, subInfo.BloggerId)

	a.log.InfoContext(ctx, "try to subscribe to blogger")

	err := a.userRep.Subscribe(ctx, repositories_transfer.SubscribeToUserInfo{
		BloggerId:    subInfo.BloggerId,
		SubscriberId: subInfo.SubscriberId,
	})
	if err != nil {
		return ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t` `Subscribe`", err))
	}

	a.log.InfoContext(ctx, "subscribed to blogger")

	return nil
}

func (a *UserService) Unsubscribe(ctx context.Context, subInfo *transfer.SubscribeInfo) error {
	logger.UpdateLoggerCtx(ctx, subscriberIdLogKey, subInfo.SubscriberId)
	logger.UpdateLoggerCtx(ctx, bloggerIdLogKey, subInfo.BloggerId)

	a.log.InfoContext(ctx, "try to unsubscribe from blogger")

	err := a.userRep.Unsubscribe(ctx, repositories_transfer.UnsubscribeInfo{
		BloggerId:    subInfo.BloggerId,
		SubscriberId: subInfo.SubscriberId,
	})
	if err != nil {
		return ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t` `Unsubscribe`", err))
	}

	a.log.InfoContext(ctx, "unsubscribed from blogger")

	return nil
}

func (a *UserService) DeleteUser(ctx context.Context, deleteInfo *transfer.DeleteUserInfo) (resErr error) {
	logger.UpdateLoggerCtx(ctx, logger.ActionUserIdKey, deleteInfo.Id)

	a.log.InfoContext(ctx, "try to delete user")

	tx, err := a.txCreator.BeginTxCtx(ctx, nil)
	if err != nil {
		return ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t begin transaction", err))
	}
	defer func() {
		resErr = services_utils.HandleErrInTransaction(resErr, tx)
	}()

	user, err := a.userRep.GetUserById(ctx, deleteInfo.Id, tx)
	if err != nil {
		return ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t` `GetUser`", err))
	}

	if err := a.userRep.DeleteUser(ctx, repositories_transfer.DeleteUserInfo{
		Id: deleteInfo.Id,
	}, tx); err != nil {
		return ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t` `DeleteUser`", err))
	}

	a.log.InfoContext(ctx, "user deleted from db")

	if err := a.userRep.DeleteUserFromCache(ctx, deleteInfo.Id); err != nil {
		return ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t` `DeleteUser`", err))
	}

	a.log.InfoContext(ctx, "user deleted from cache")

	eventId := uuid.New()

	logger.UpdateLoggerCtx(ctx, logger.EventIdKey, eventId)

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
		return ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t` `MarshalMessage`", err))
	}

	if err := a.eventsRep.Create(ctx, repositories_transfer.CreateEventInfo{
		EventId:   eventId,
		EventType: amqpclient.UserDeletedEventKey,
		Payload:   messageJson,
	}, tx); err != nil {
		return ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t` `CreateEvent`", err))
	}

	a.log.InfoContext(ctx, "delete event is created successfully")

	if err := tx.Commit(); err != nil {
		return ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t` `CreateEvent`", err))
	}

	return nil
}

func (a *UserService) UploadAvatar(ctx context.Context, uploadInfo *transfer.UploadAvatarInfo) (resAvatar *transfer.AvatarResult, resErr error) {
	logger.UpdateLoggerCtx(ctx, logger.ActionUserIdKey, uploadInfo.UserId)

	a.log.Debug("try to upload avatar")

	tx, err := a.txCreator.BeginTxCtx(ctx, nil)
	if err != nil {
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t` `CreateEvent`", err))
	}
	defer func() {
		resErr = services_utils.HandleErrInTransaction(resErr, tx)
	}()

	imgUrl, err := a.imgStore.UploadFile(ctx, fmt.Sprintf("%s.jpeg", uuid.New().String()), uploadInfo.Image, s3client.ImageJpeg)
	if err != nil {
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t` `CreateEvent`", err))
	}

	logger.UpdateLoggerCtx(ctx, avatarUruLogKey, imgUrl)
	a.log.Debug("avatar is uploaded successfully")

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
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t` `CreateEvent`", err))
	}

	if err := a.userRep.DeleteUserFromCache(ctx, uploadInfo.UserId); err != nil {
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t` `CreateEvent`", err))
	}

	a.log.Debug("user deleted from cache")

	if err := tx.Commit(); err != nil {
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t` `CreateEvent`", err))
	}

	a.log.Debug("avatar is uploaded successfully")

	return &transfer.AvatarResult{
		UserId:     uploadInfo.UserId,
		Avatar:     imgUrl,
		AvatarMini: imgUrl,
	}, nil
}

func (a *UserService) CompensateDeletedUser(ctx context.Context, eventId uuid.UUID) error {
	logger.UpdateLoggerCtx(ctx, logger.EventIdKey, eventId)

	a.log.InfoContext(ctx, "trying to compensate user")

	eventInfo, err := a.eventsRep.GetEventById(ctx, eventId)
	if err != nil {
		return ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t` `GetEvent`", err))
	}

	var message messages.UserDeletedMessage
	if err := json.Unmarshal([]byte(eventInfo.Payload), &message); err != nil {
		return ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t` `GetEvent`", err))
	}

	logger.UpdateLoggerCtx(ctx, logger.ActionUserIdKey, message.User.Id)
	a.log.InfoContext(ctx, "event is found")

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
		return ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t` `CreateEvent`", err))
	}

	a.log.InfoContext(ctx, "deleted user is compensated successfully")

	return nil
}
