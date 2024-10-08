package services

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/KBcHMFollower/blog_user_service/internal/clients/amqpclient"
	"github.com/KBcHMFollower/blog_user_service/internal/clients/amqpclient/messages"
	s3client "github.com/KBcHMFollower/blog_user_service/internal/clients/s3/minio"
	ctxerrors "github.com/KBcHMFollower/blog_user_service/internal/domain/errors"
	repositoriestransfer "github.com/KBcHMFollower/blog_user_service/internal/domain/layers_TOs/repositories"
	transfer "github.com/KBcHMFollower/blog_user_service/internal/domain/layers_TOs/services"
	"github.com/KBcHMFollower/blog_user_service/internal/domain/models"
	"github.com/KBcHMFollower/blog_user_service/internal/logger"
	dep "github.com/KBcHMFollower/blog_user_service/internal/services/interfaces/dep"
	servicesutils "github.com/KBcHMFollower/blog_user_service/internal/services/lib"
	"time"

	"github.com/google/uuid"
)

const (
	createdUserIdLogKey = "created-user-id"
	accessTokenLogKey   = "access-token"
	bloggerIdLogKey     = "blogger-id"
	subscriberIdLogKey  = "subscriber-id"
	updateInfoLogKey    = "update-info"
	avatarUruLogKey     = "avatar-uru"
)

type usrSvcEventStore interface {
	dep.EventCreator
	dep.EventGetter
}

type usrSvcImageStore interface {
	dep.ImageGetter
	dep.ImageUploader
}

type usrSvcUsersStore interface {
	dep.UserDeleter
	dep.UserGetter
	dep.UserUpdater
	dep.UserCreator
}

type subsSvcSubscribersStore interface {
	dep.SubscribersGetter
	dep.SubscribersDealer
}

type UserService struct {
	log       logger.Logger
	userRep   usrSvcUsersStore
	eventsRep usrSvcEventStore
	subsRep   subsSvcSubscribersStore
	txCreator dep.TransactionCreator
	imgStore  usrSvcImageStore
}

func NewUserService(log logger.Logger, txCreator dep.TransactionCreator, userRep usrSvcUsersStore, eventsRep usrSvcEventStore, imgStore usrSvcImageStore) *UserService {
	return &UserService{
		log:       log,
		userRep:   userRep,
		imgStore:  imgStore,
		txCreator: txCreator,
		eventsRep: eventsRep,
	}
}

func (a *UserService) GetUserById(ctx context.Context, userId uuid.UUID) (ersUser *transfer.GetUserResult, resErr error) {
	ctx = logger.UpdateLoggerCtx(ctx, logger.ActionUserIdKey, userId)

	a.log.DebugContext(ctx, "try to get user by id")

	tx, err := a.txCreator.BeginTxCtx(ctx, nil)
	if err != nil {
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t begin tx", err))
	}
	defer func() {
		resErr = servicesutils.HandleErrInTransaction(resErr, tx)
	}()

	cacheUser, err := a.userRep.TryGetFromCache(ctx, userId)
	if err != nil {
		a.log.DebugContext(ctx, "can`t get cacheUser from cache: ", "err", err.Error())
	}
	if cacheUser != nil {
		a.log.DebugContext(ctx, "user found in cache")
		return &transfer.GetUserResult{
			User: transfer.GetUserResultFromModel(cacheUser),
		}, nil
	}

	a.log.DebugContext(ctx, "try to get user by id from db")

	user, err := a.userRep.User(ctx, repositoriestransfer.GetUserInfo{
		map[repositoriestransfer.UserFieldTarget]any{
			repositoriestransfer.UserIdCondition: userId,
		},
	}, nil)
	if err != nil {
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t get cacheUser from db", err))
	}

	a.log.DebugContext(ctx, "user found in db")

	if err := a.userRep.SetToCache(ctx, user); err != nil {
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t set cacheUser to db", err))
	}
	a.log.DebugContext(ctx, "user added to cache")

	if err := tx.Commit(); err != nil {
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t commit tx", err))
	}

	a.log.DebugContext(ctx, "get user by id successfully")

	return &transfer.GetUserResult{
		User: transfer.GetUserResultFromModel(user),
	}, nil
}

func (a *UserService) UpdateUser(ctx context.Context, updateInfo *transfer.UpdateUserInfo) (resUser *transfer.UpdateUserResult, resErr error) {
	ctx = logger.UpdateLoggerCtx(ctx, logger.ActionUserIdKey, updateInfo.Id)
	ctx = logger.UpdateLoggerCtx(ctx, updateInfoLogKey, updateInfo.UpdateFields)

	a.log.InfoContext(ctx, "try to update user")

	tx, err := a.txCreator.BeginTxCtx(ctx, nil)
	if err != nil {
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t begin tx", err))
	}
	defer func() {
		resErr = servicesutils.HandleErrInTransaction(resErr, tx)
	}()

	if err := a.userRep.Update(ctx, repositoriestransfer.UpdateUserInfo{
		Id:         updateInfo.Id,
		UpdateInfo: servicesutils.ConvertMapKeysToStrings(updateInfo.UpdateFields),
	}, tx); err != nil {
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t update user in db", err))
	}

	exUser, exErr := a.userRep.User(ctx, repositoriestransfer.GetUserInfo{
		Condition: map[repositoriestransfer.UserFieldTarget]any{
			repositoriestransfer.UserIdCondition: updateInfo.Id,
		},
	}, tx)
	if exErr != nil {
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t get user from db", err))
	}

	if err := a.userRep.DeleteFromCache(ctx, updateInfo.Id); err != nil {
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

func (a *UserService) DeleteUser(ctx context.Context, deleteInfo *transfer.DeleteUserInfo) (resErr error) {
	ctx = logger.UpdateLoggerCtx(ctx, logger.ActionUserIdKey, deleteInfo.Id)

	a.log.InfoContext(ctx, "try to delete user")

	tx, err := a.txCreator.BeginTxCtx(ctx, nil)
	if err != nil {
		return ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t begin transaction", err))
	}
	defer func() {
		resErr = servicesutils.HandleErrInTransaction(resErr, tx)
	}()

	user, err := a.userRep.User(ctx, repositoriestransfer.GetUserInfo{
		Condition: map[repositoriestransfer.UserFieldTarget]any{
			repositoriestransfer.UserIdCondition: deleteInfo.Id,
		},
	}, tx)
	if err != nil {
		return ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t` `GetUser`", err))
	}

	if err := a.userRep.Delete(ctx, repositoriestransfer.DeleteUserInfo{
		Id: deleteInfo.Id,
	}, tx); err != nil {
		return ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t` `DeleteUser`", err))
	}

	a.log.InfoContext(ctx, "user deleted from db")

	if err := a.userRep.DeleteFromCache(ctx, deleteInfo.Id); err != nil {
		return ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t` `DeleteUser`", err))
	}

	a.log.InfoContext(ctx, "user deleted from cache")

	eventId := uuid.New()

	ctx = logger.UpdateLoggerCtx(ctx, logger.EventIdKey, eventId)

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

	if err := a.eventsRep.Create(ctx, repositoriestransfer.CreateEventInfo{
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
	ctx = logger.UpdateLoggerCtx(ctx, logger.ActionUserIdKey, uploadInfo.UserId)

	a.log.DebugContext(ctx, "try to upload avatar")

	tx, err := a.txCreator.BeginTxCtx(ctx, nil)
	if err != nil {
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t` `CreateEvent`", err))
	}
	defer func() {
		resErr = servicesutils.HandleErrInTransaction(resErr, tx)
	}()

	imgUrl, err := a.imgStore.UploadFile(ctx, fmt.Sprintf("%s.jpeg", uuid.New().String()), uploadInfo.Image, s3client.ImageJpeg)
	if err != nil {
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t` `CreateEvent`", err))
	}

	ctx = logger.UpdateLoggerCtx(ctx, avatarUruLogKey, imgUrl)
	a.log.DebugContext(ctx, "avatar is uploaded successfully")

	if err = a.userRep.Update(ctx, repositoriestransfer.UpdateUserInfo{
		Id: uploadInfo.UserId,
		UpdateInfo: map[string]interface{}{
			"avatar":     imgUrl,
			"avatar_min": imgUrl,
		},
	}, nil); err != nil {
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t` `CreateEvent`", err))
	}

	if err := a.userRep.DeleteFromCache(ctx, uploadInfo.UserId); err != nil {
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t` `CreateEvent`", err))
	}

	a.log.DebugContext(ctx, "user deleted from cache")

	if err := tx.Commit(); err != nil {
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t` `CreateEvent`", err))
	}

	a.log.DebugContext(ctx, "avatar is uploaded successfully")

	return &transfer.AvatarResult{
		UserId:     uploadInfo.UserId,
		Avatar:     imgUrl,
		AvatarMini: imgUrl,
	}, nil
}

func (a *UserService) CompensateDeletedUser(ctx context.Context, eventId uuid.UUID) error {
	ctx = logger.UpdateLoggerCtx(ctx, logger.EventIdKey, eventId)

	a.log.InfoContext(ctx, "trying to compensate user")

	eventInfo, err := a.eventsRep.Event(ctx, eventId)
	if err != nil {
		return ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t` `GetEvent`", err))
	}

	var message messages.UserDeletedMessage
	if err := json.Unmarshal([]byte(eventInfo.Payload), &message); err != nil {
		return ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("can`t` `GetEvent`", err))
	}

	ctx = logger.UpdateLoggerCtx(ctx, logger.ActionUserIdKey, message.User.Id)
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
