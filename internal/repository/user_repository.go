package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/KBcHMFollower/blog_user_service/internal/clients/cashe"
	"github.com/KBcHMFollower/blog_user_service/internal/database"
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
	op := "UserRepository.getSubInfo"

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
		return nil, 0, fmt.Errorf("%s : failed to build sql query : %w", op, err)
	}

	subscribers := make([]*models.Subscriber, 0)
	if err := r.db.SelectContext(ctx, &subscribers, toSql, args...); err != nil {
		return nil, 0, fmt.Errorf("%s : failed to execute sql : %w", op, err)
	}

	query = builder.
		Select("COUNT(*)").
		From(SubscribersTable).
		Where(squirrel.Eq{getInfo.TargetType: getInfo.UserId})

	toSql, args, err = query.ToSql()
	if err != nil {
		return subscribers, 0, fmt.Errorf("%s : failed to build sql query : %w", op, err)
	}

	var totalCount uint32
	if err := r.db.GetContext(ctx, &totalCount, toSql, args...); err != nil {
		return nil, 0, fmt.Errorf("%s : failed to execute sql : %w", op, err)
	}

	return subscribers, totalCount, nil
}

func (r *UserRepository) CreateUser(ctx context.Context, createDto *transfer.CreateUserInfo) (uuid.UUID, error) {
	op := "UserRepository.CreateUser"

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
		return uuid.Nil, fmt.Errorf("%s : failed to build sql query : %w", op, err)
	}

	var id uuid.UUID
	if err := r.db.GetContext(ctx, &id, sql, args...); err != nil {
		return uuid.Nil, fmt.Errorf("%s : failed to execute sql : %w", op, err)
	}

	return id, nil
}

func (r *UserRepository) RollBackUser(ctx context.Context, user models.User) error {
	op := "UserRepository.RollBackUser"

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
		return fmt.Errorf("%s : failed to build toSql query : %w", op, err)
	}

	var id uuid.UUID
	if err := r.db.GetContext(ctx, &id, toSql, args...); err != nil {
		return fmt.Errorf("%s : failed to execute sql : %w", op, err)
	}

	return nil
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	op := "UserRepository.GetUserByEmail"
	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	sql, args, err := builder.Select("*").
		From(UsersTable).
		Where(squirrel.Eq{"email": email}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s : failed to build sql query : %w", op, err)
	}

	var user models.User
	if err := r.db.GetContext(ctx, &user, sql, args...); err != nil {
		return nil, fmt.Errorf("%s : failed to execute sql : %w", op, err)
	}

	return &user, nil
}

func (r *UserRepository) GetUserById(ctx context.Context, userId uuid.UUID, tx database.Transaction) (*models.User, error) {
	op := "UserRepository.GetUserById"

	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	executor := rep_utils.GetExecutor(r.db, tx)

	sql, args, err := builder.Select("*").
		From(UsersTable).
		Where(squirrel.Eq{"id": userId}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s : failed to build sql query : %w", op, err)
	}

	var user models.User
	if err := executor.GetContext(ctx, &user, sql, args...); err != nil {
		return nil, fmt.Errorf("%s : failed to execute sql : %w", op, err)
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
		return nil, totalCount, fmt.Errorf("%s : failed to build sql query : %w", op, err)
	}

	users := make([]*models.User, 0)
	if err := r.db.SelectContext(ctx, &users, sql, args...); err != nil {
		return nil, totalCount, fmt.Errorf("%s : failed to execute sql : %w", op, err)
	}

	return users, totalCount, nil
}

func (r *UserRepository) GetUserSubscriptions(ctx context.Context, getInfo transfer.GetUserSubscriptionsInfo) ([]*models.User, uint32, error) {
	op := "UserRepository.GetUserSubscribers"

	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	subscribers, totalCount, err := r.getSubInfo(ctx, transfer.GetSubscriptionInfo{
		UserId:     getInfo.UserId,
		Page:       getInfo.Page,
		Size:       getInfo.Size,
		TargetType: "subscriber_id",
	})
	if err != nil {
		return nil, totalCount, fmt.Errorf("%s : failed to get sub info : %w", op, err)
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
		return nil, totalCount, fmt.Errorf("%s : failed to execute sql : %w", op, err)
	}

	return users, totalCount, nil
}

func (r *UserRepository) UpdateUser(ctx context.Context, updateData transfer.UpdateUserInfo, tx database.Transaction) error { //TODO
	op := "UserRepository.UpdateUser"
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
		return fmt.Errorf("%s : failed to build sql query : %w", op, err)
	}

	_, err = executor.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("%s : failed to execute sql : %w", op, err)
	}

	return nil
}

func (r *UserRepository) Subscribe(ctx context.Context, subInfo transfer.SubscribeToUserInfo) error {
	op := "UserRepository.Subscribe"
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
		return fmt.Errorf("%s : failed to build sql query : %w", op, err)
	}

	_, err = r.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("%s : failed to execute sql : %w", op, err)
	}

	return nil
}

func (r *UserRepository) Unsubscribe(ctx context.Context, unsubInfo transfer.UnsubscribeInfo) error {
	op := "UserRepository.Unsubscribe"
	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	query := builder.Delete(SubscribersTable).
		Where(squirrel.Eq{"blogger_id": unsubInfo.BloggerId, "subscriber_id": unsubInfo.SubscriberId})

	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("%s : failed to build sql query : %w", op, err)
	}

	_, err = r.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("%s : failed to execute sql : %w", op, err)
	}

	return nil
}

func (r *UserRepository) DeleteUser(ctx context.Context, delInfo transfer.DeleteUserInfo, tx database.Transaction) error {
	op := "UserRepository.DeleteUser"

	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	executor := rep_utils.GetExecutor(r.db, tx)

	query := builder.Delete(UsersTable).Where(squirrel.Eq{"id": delInfo.Id})
	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("%s : failed to build sql query : %w", op, err)
	}

	if _, err := executor.ExecContext(ctx, sql, args...); err != nil {
		return fmt.Errorf("%s : failed to execute sql : %w", op, err)
	}

	return nil
}

// TODO: ДОЛЖНО ВЫЗЫВАТЬСЯ ИЗ СЕРВИСА
func (r *UserRepository) TryGetUserFromCache(ctx context.Context, id uuid.UUID) (*models.User, error) {
	op := "UserRepository.TryGetUserFromCache"

	data, err := r.cache.Get(ctx, fmt.Sprintf("%s%s", UsersCachePref, id.String()))
	if err != nil {
		return nil, fmt.Errorf("%s : failed to read from cache : %w", op, err)
	}

	var user *models.User

	if err := json.Unmarshal([]byte(data), &user); err != nil {
		return nil, fmt.Errorf("%s : failed to unmarshal json : %w", op, err)
	}

	return user, nil
}

func (r *UserRepository) SetUserToCache(ctx context.Context, user *models.User) error {
	op := "UserRepository.SetUserToCache"

	userJson, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("%s : failed to marshal json : %w", op, err)
	}

	err = r.cache.Set(ctx, fmt.Sprintf("%s%s", UsersCachePref, user.Id.String()), userJson)
	if err != nil {
		return fmt.Errorf("%s : failed to write to cache : %w", op, err)
	}

	return nil
}

func (r *UserRepository) DeleteUserFromCache(ctx context.Context, id uuid.UUID) error {
	op := "UserRepository.DeleteUserFromCache"

	err := r.cache.Delete(ctx, fmt.Sprintf("%s%s", UsersCachePref, id.String()))
	if err != nil {
		return fmt.Errorf("%s : failed to delete from cache : %w", op, err)
	}

	return nil
} //TODO: ДОБАВИТЬ В УДАЛЕНИЕ ПОЛЬЗОВАТЕЛЯ
