package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/KBcHMFollower/blog_user_service/internal/clients/cashe"
	"github.com/KBcHMFollower/blog_user_service/internal/database"
	ctxerrors "github.com/KBcHMFollower/blog_user_service/internal/domain/errors"
	transfer "github.com/KBcHMFollower/blog_user_service/internal/domain/layers_TOs/repositories"
	rep_utils "github.com/KBcHMFollower/blog_user_service/internal/repository/lib"
	"time"

	"github.com/KBcHMFollower/blog_user_service/internal/domain/models"
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

const (
	UsersTable             = "users"
	SubscribersTable       = "subscribers"
	TransactionEventsTable = "transaction_events"
	UsersCachePref         = "userId-"
)

type UserRepository struct {
	db    database.DBWrapper
	cache cashe.CasheStorage
}

func NewUserRepository(dbDriver database.DBWrapper, cacheStorage cashe.CasheStorage) *UserRepository {
	return &UserRepository{db: dbDriver, cache: cacheStorage}
}

func (r *UserRepository) getSubInfo(ctx context.Context, getInfo transfer.GetSubscriptionInfo) ([]*models.Subscriber, uint32, error) {
	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	offset := (getInfo.Page - 1) * getInfo.Size

	query := builder.
		Select("*").
		From(SubscribersTable).
		Where(squirrel.Eq{getInfo.TargetType: getInfo.UserId}).
		Limit(getInfo.Size).
		Offset(offset)

	toSql, args, err := query.ToSql()
	if err != nil {
		return nil, 0, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap(rep_utils.FailedToGenerateSqlMessage, err))
	}

	subscribers := make([]*models.Subscriber, 0)
	if err := r.db.SelectContext(ctx, &subscribers, toSql, args...); err != nil {
		return nil, 0, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap(rep_utils.FailedToExecuteQuery, err))
	}

	query = builder.
		Select("COUNT(*)").
		From(SubscribersTable).
		Where(squirrel.Eq{getInfo.TargetType: getInfo.UserId})

	toSql, args, err = query.ToSql()
	if err != nil {
		return subscribers, 0, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap(rep_utils.FailedToGenerateSqlMessage, err))
	}

	var totalCount uint32
	if err := r.db.GetContext(ctx, &totalCount, toSql, args...); err != nil {
		return nil, 0, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap(rep_utils.FailedToExecuteQuery, err))
	}

	return subscribers, totalCount, nil
}

func (r *UserRepository) CreateUser(ctx context.Context, createDto *transfer.CreateUserInfo) (uuid.UUID, error) {
	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	user := models.NewUserModel(createDto.Email, createDto.FName, createDto.LName, createDto.HashPass)

	query := builder.Insert(UsersTable).
		SetMap(map[string]interface{}{
			"id":         user.Id,
			"email":      user.Email,
			"pass_hash":  user.PassHash,
			"avatar":     user.Avatar,
			"avatar_min": user.AvatarMin,
			"fname":      user.FName,
			"lname":      user.LName,
		}).
		Suffix("RETURNING \"id\"")

	sql, args, err := query.ToSql()
	if err != nil {
		return uuid.Nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap(rep_utils.FailedToGenerateSqlMessage, err))
	}

	var id uuid.UUID
	if err := r.db.GetContext(ctx, &id, sql, args...); err != nil {
		return uuid.Nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap(rep_utils.FailedToExecuteQuery, err))
	}

	return id, nil
}

func (r *UserRepository) RollBackUser(ctx context.Context, user models.User) error {
	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	query := builder.Insert(UsersTable).
		SetMap(map[string]interface{}{ //TODO: МОЖЕТ ЕСТЬ КАКАЯ-ТО ХЕРНЯ, КОТОРАЯ САМА МАПИТ?
			"id":           user.Id,
			"email":        user.Email,
			"pass_hash":    user.PassHash,
			"avatar":       user.Avatar,
			"avatar_min":   user.AvatarMin,
			"fname":        user.FName,
			"lname":        user.LName,
			"created_date": user.CreatedDate,
			"updated_date": time.Now(),
		}).
		Suffix("RETURNING \"id\"")

	toSql, args, err := query.ToSql()
	if err != nil {
		return ctxerrors.WrapCtx(ctx, ctxerrors.Wrap(rep_utils.FailedToGenerateSqlMessage, err))
	}

	var id uuid.UUID
	if err := r.db.GetContext(ctx, &id, toSql, args...); err != nil {
		return ctxerrors.WrapCtx(ctx, ctxerrors.Wrap(rep_utils.FailedToExecuteQuery, err))
	}

	return nil
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	sql, args, err := builder.Select("*").
		From(UsersTable).
		Where(squirrel.Eq{"email": email}).
		ToSql()
	if err != nil {
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap(rep_utils.FailedToGenerateSqlMessage, err))
	}

	var user models.User
	if err := r.db.GetContext(ctx, &user, sql, args...); err != nil {
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap(rep_utils.FailedToExecuteQuery, err))
	}

	return &user, nil
}

func (r *UserRepository) GetUserById(ctx context.Context, userId uuid.UUID, tx database.Transaction) (*models.User, error) {
	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	executor := rep_utils.GetExecutor(r.db, tx)

	sql, args, err := builder.Select("*").
		From(UsersTable).
		Where(squirrel.Eq{"id": userId}).
		ToSql()
	if err != nil {
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap(rep_utils.FailedToGenerateSqlMessage, err))
	}

	var user models.User
	if err := executor.GetContext(ctx, &user, sql, args...); err != nil {
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap(rep_utils.FailedToExecuteQuery, err))
	}

	return &user, nil
}

func (r *UserRepository) GetUserSubscribers(ctx context.Context, getInfo transfer.GetUserSubscribersInfo) ([]*models.User, uint32, error) {
	op := "UserRepository.GetUserSubscribers"

	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	subscribers, totalCount, err := r.getSubInfo(ctx, transfer.GetSubscriptionInfo{
		UserId:     getInfo.UserId,
		Page:       getInfo.Page,
		Size:       getInfo.Size,
		TargetType: "blogger_id",
	})
	if err != nil {
		return nil, totalCount, fmt.Errorf("%s : %w", op, err)
	}

	subscribersId := make([]uuid.UUID, 0)
	for _, subscriber := range subscribers {
		subscribersId = append(subscribersId, subscriber.SubscriberId)
	}

	sql, args, err := builder.Select("*").
		From(UsersTable).
		Where(squirrel.Eq{"id": subscribersId}).
		ToSql()
	if err != nil {
		return nil, totalCount, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap(rep_utils.FailedToGenerateSqlMessage, err))
	}

	users := make([]*models.User, 0)
	if err := r.db.SelectContext(ctx, &users, sql, args...); err != nil {
		return nil, totalCount, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap(rep_utils.FailedToExecuteQuery, err))
	}

	return users, totalCount, nil
}

func (r *UserRepository) GetUserSubscriptions(ctx context.Context, getInfo transfer.GetUserSubscriptionsInfo) ([]*models.User, uint32, error) {
	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	subscribers, totalCount, err := r.getSubInfo(ctx, transfer.GetSubscriptionInfo{
		UserId:     getInfo.UserId,
		Page:       getInfo.Page,
		Size:       getInfo.Size,
		TargetType: "subscriber_id",
	})
	if err != nil {
		return nil, totalCount, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap(rep_utils.FailedToGenerateSqlMessage, err))
	}

	subscribersId := make([]uuid.UUID, 0)
	for _, subscriber := range subscribers {
		subscribersId = append(subscribersId, subscriber.BloggerId)
	}

	sql, args, err := builder.Select("*").
		From(UsersTable).
		Where(squirrel.Eq{"id": subscribersId}).
		ToSql()

	users := make([]*models.User, 0)
	if err := r.db.SelectContext(ctx, &users, sql, args...); err != nil {
		return nil, totalCount, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap(rep_utils.FailedToExecuteQuery, err))
	}

	return users, totalCount, nil
}

func (r *UserRepository) UpdateUser(ctx context.Context, updateData transfer.UpdateUserInfo, tx database.Transaction) error { //TODO
	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	executor := rep_utils.GetExecutor(r.db, tx)

	query := builder.
		Update(UsersTable).
		Where(squirrel.Eq{"id": updateData.Id}).
		Set("updated_date", time.Now())

	for _, item := range updateData.UpdateInfo {
		if item.Name == "id" || item.Name == "pass_hash" || item.Name == "created_date" || item.Name == "updated_date" {
			continue
		}
		query = query.Set(item.Name, item.Value)
	}

	sql, args, err := query.ToSql()
	if err != nil {
		return ctxerrors.WrapCtx(ctx, ctxerrors.Wrap(rep_utils.FailedToGenerateSqlMessage, err))
	}

	_, err = executor.ExecContext(ctx, sql, args...)
	if err != nil {
		return ctxerrors.WrapCtx(ctx, ctxerrors.Wrap(rep_utils.FailedToExecuteQuery, err))
	}

	return nil
}

func (r *UserRepository) Subscribe(ctx context.Context, subInfo transfer.SubscribeToUserInfo) error {
	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	subscribers := models.NewSubscriber(subInfo.BloggerId, subInfo.SubscriberId)

	query := builder.Insert(SubscribersTable).
		SetMap(map[string]interface{}{
			"id":            subscribers.Id,
			"blogger_id":    subscribers.BloggerId,
			"subscriber_id": subscribers.SubscriberId,
		})

	sql, args, err := query.ToSql()
	if err != nil {
		return ctxerrors.WrapCtx(ctx, ctxerrors.Wrap(rep_utils.FailedToGenerateSqlMessage, err))
	}

	_, err = r.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return ctxerrors.WrapCtx(ctx, ctxerrors.Wrap(rep_utils.FailedToExecuteQuery, err))
	}

	return nil
}

func (r *UserRepository) Unsubscribe(ctx context.Context, unsubInfo transfer.UnsubscribeInfo) error {
	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	query := builder.Delete(SubscribersTable).
		Where(squirrel.Eq{"blogger_id": unsubInfo.BloggerId, "subscriber_id": unsubInfo.SubscriberId})

	sql, args, err := query.ToSql()
	if err != nil {
		return ctxerrors.WrapCtx(ctx, ctxerrors.Wrap(rep_utils.FailedToGenerateSqlMessage, err))
	}

	_, err = r.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return ctxerrors.WrapCtx(ctx, ctxerrors.Wrap(rep_utils.FailedToExecuteQuery, err))
	}

	return nil
}

func (r *UserRepository) DeleteUser(ctx context.Context, delInfo transfer.DeleteUserInfo, tx database.Transaction) error {
	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	executor := rep_utils.GetExecutor(r.db, tx)

	query := builder.Delete(UsersTable).Where(squirrel.Eq{"id": delInfo.Id})
	sql, args, err := query.ToSql()
	if err != nil {
		return ctxerrors.WrapCtx(ctx, ctxerrors.Wrap(rep_utils.FailedToGenerateSqlMessage, err))
	}

	if _, err := executor.ExecContext(ctx, sql, args...); err != nil {
		return ctxerrors.WrapCtx(ctx, ctxerrors.Wrap(rep_utils.FailedToExecuteQuery, err))
	}

	return nil
}

// TODO: ДОЛЖНО ВЫЗЫВАТЬСЯ ИЗ СЕРВИСА
func (r *UserRepository) TryGetUserFromCache(ctx context.Context, id uuid.UUID) (*models.User, error) {
	data, err := r.cache.Get(ctx, fmt.Sprintf("%s%s", UsersCachePref, id.String()))
	if err != nil {
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("failed to read from cache", err))
	}

	var user *models.User

	if err := json.Unmarshal([]byte(data), &user); err != nil {
		return nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("failed to unmarshal data", err))
	}

	return user, nil
}

func (r *UserRepository) SetUserToCache(ctx context.Context, user *models.User) error {
	userJson, err := json.Marshal(user)
	if err != nil {
		return ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("failed to unmarshal data", err))
	}

	err = r.cache.Set(ctx, fmt.Sprintf("%s%s", UsersCachePref, user.Id.String()), userJson)
	if err != nil {
		return ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("failed to write to cache", err))
	}

	return nil
}

func (r *UserRepository) DeleteUserFromCache(ctx context.Context, id uuid.UUID) error {
	err := r.cache.Delete(ctx, fmt.Sprintf("%s%s", UsersCachePref, id.String()))
	if err != nil {
		return ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("failed to delete from cache", err))
	}

	return nil
} //TODO: ДОБАВИТЬ В УДАЛЕНИЕ ПОЛЬЗОВАТЕЛЯ
