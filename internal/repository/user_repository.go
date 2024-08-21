package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/KBcHMFollower/blog_user_service/internal/clients/cashe"
	transfer "github.com/KBcHMFollower/blog_user_service/internal/domain/layers_TOs/repositories"
	"time"

	"github.com/KBcHMFollower/blog_user_service/database"
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

func NewUserRepository(dbDriver database.DBWrapper, cacheStorage cashe.CasheStorage) (*UserRepository, error) {
	return &UserRepository{db: dbDriver, cache: cacheStorage}, nil
}

func (r *UserRepository) getSubInfo(ctx context.Context, getInfo transfer.GetSubscriptionInfo) ([]*models.Subscriber, uint32, error) {
	op := "UserRepository.getSubInfo"

	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	subscribers := make([]*models.Subscriber, 0)

	offset := (getInfo.Page - 1) * getInfo.Size

	query := builder.
		Select("*").
		From(SubscribersTable).
		Where(squirrel.Eq{getInfo.TargetType: getInfo.UserId}).
		Limit(getInfo.Size).
		Offset(offset)

	sql, args, err := query.ToSql()
	if err != nil {
		return subscribers, 0, fmt.Errorf("%s : %w", op, err)
	}

	rows, err := r.db.QueryContext(ctx, sql, args...)
	if err != nil {
		return subscribers, 0, fmt.Errorf("%s : %w", op, err)
	}
	defer rows.Close()

	for rows.Next() {
		var subscriber models.Subscriber

		err := rows.Scan(subscriber.GetPointersArray()...)
		if err != nil {
			return subscribers, 0, fmt.Errorf("error in parse post from db: %v", err)
		}

		subscribers = append(subscribers, &subscriber)
	}

	query = builder.
		Select("COUNT(*)").
		From(SubscribersTable).
		Where(squirrel.Eq{getInfo.TargetType: getInfo.UserId})

	sql, args, err = query.ToSql()
	if err != nil {
		return subscribers, 0, fmt.Errorf("%s : %w", op, err)
	}

	var totalCount uint32

	countRow := r.db.QueryRowContext(ctx, sql, args...)
	if err := countRow.Scan(&totalCount); err != nil {
		return nil, 0, fmt.Errorf("%s : %w", op, err)
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
		return uuid.Nil, fmt.Errorf("%s : %w", op, err)
	}

	row := r.db.QueryRowContext(ctx, sql, args...)

	var id uuid.UUID

	err = row.Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%s : %w", op, err)
	}

	return id, nil
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	op := "UserRepository.GetUserByEmail"
	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	sql, args, err := builder.Select("*").
		From(UsersTable).
		Where(squirrel.Eq{"email": email}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s : %w", op, err)
	}

	var user models.User

	row := r.db.QueryRowContext(ctx, sql, args...)

	err = row.Scan(user.GetPointersArray()...)
	if err != nil {
		return nil, fmt.Errorf("%s : %w", op, err)
	}

	return &user, nil
}

func (r *UserRepository) GetUserById(ctx context.Context, userId uuid.UUID) (*models.User, error) {
	op := "UserRepository.GetUserById"

	if user, err := r.tryGetUserFromCache(ctx, userId); err == nil {
		return user, nil
	}

	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	sql, args, err := builder.Select("*").
		From(UsersTable).
		Where(squirrel.Eq{"id": userId}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s : %w", op, err)
	}

	var user models.User

	row := r.db.QueryRowContext(ctx, sql, args...)

	err = row.Scan(user.GetPointersArray()...)
	if err != nil {
		return nil, fmt.Errorf("%s : %w", op, err)
	}

	if err := r.SetUserToCache(ctx, &user); err != nil {
		return &user, err
	}

	return &user, nil
}

func (r *UserRepository) GetUserSubscribers(ctx context.Context, getInfo transfer.GetUserSubscribersInfo) ([]*models.User, uint32, error) {
	op := "UserRepository.GetUserSubscribers"

	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	users := make([]*models.User, 0)

	subscribers, totalCount, err := r.getSubInfo(ctx, transfer.GetSubscriptionInfo{
		UserId:     getInfo.UserId,
		Page:       getInfo.Page,
		Size:       getInfo.Size,
		TargetType: "blogger_id",
	})
	if err != nil {
		return users, totalCount, fmt.Errorf("%s : %w", op, err)
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
		return users, totalCount, fmt.Errorf("%s : %w", op, err)
	}

	rows, err := r.db.QueryContext(ctx, sql, args...)
	if err != nil {
		return users, 0, fmt.Errorf("%s : %w", op, err)
	}
	defer rows.Close()

	for rows.Next() {
		var user models.User
		err = rows.Scan(user.GetPointersArray()...)
		if err != nil {
			return users, 0, fmt.Errorf("%s : %w", op, err)
		}
		users = append(users, &user)
	}

	return users, totalCount, nil
}

func (r *UserRepository) GetUserSubscriptions(ctx context.Context, getInfo transfer.GetUserSubscriptionsInfo) ([]*models.User, uint32, error) {
	op := "UserRepository.GetUserSubscribers"

	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	users := make([]*models.User, 0)

	subscribers, totalCount, err := r.getSubInfo(ctx, transfer.GetSubscriptionInfo{
		UserId:     getInfo.UserId,
		Page:       getInfo.Page,
		Size:       getInfo.Size,
		TargetType: "subscriber_id",
	})
	if err != nil {
		return users, totalCount, fmt.Errorf("%s : %w", op, err)
	}

	subscribersId := make([]uuid.UUID, 0)
	for _, subscriber := range subscribers {
		subscribersId = append(subscribersId, subscriber.BloggerId)
	}

	sql, args, err := builder.Select("*").
		From(UsersTable).
		Where(squirrel.Eq{"id": subscribersId}).
		ToSql()

	rows, err := r.db.QueryContext(ctx, sql, args...)
	if err != nil {
		return users, 0, fmt.Errorf("%s : %w", op, err)
	}
	defer rows.Close()

	for rows.Next() {
		var user models.User
		err = rows.Scan(user.GetPointersArray()...)
		if err != nil {
			return users, 0, fmt.Errorf("%s : %w", op, err)
		}
		users = append(users, &user)
	}

	return users, totalCount, nil
}

func (r *UserRepository) UpdateUser(ctx context.Context, updateData transfer.UpdateUserInfo) (*models.User, error) {
	op := "UserRepository.UpdateUser"
	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

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
		return nil, fmt.Errorf("%s : %w", op, err)
	}

	_, err = r.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("%s : %w", op, err)
	}

	queryGetPost := builder.
		Select("*").
		From(UsersTable).
		Where(squirrel.Eq{"id": updateData.Id})
	sqlGetPost, argsGetPost, _ := queryGetPost.ToSql()

	row := r.db.QueryRowContext(ctx, sqlGetPost, argsGetPost...)

	var user models.User
	err = row.Scan(user.GetPointersArray()...)
	if err != nil {
		return nil, fmt.Errorf("%s : %w", op, err)
	}

	if err := r.DeleteUserFromCache(ctx, updateData.Id); err != nil {
		return &user, nil
	}

	return &user, nil
}

func (r *UserRepository) Subscribe(ctx context.Context, subInfo transfer.SubscribeToUserInfo) error {
	op := "UserRepository.Subscribe"
	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	subscribers := models.NewSubscriber(subInfo.BloggerId, subInfo.SubscriberId)

	query := builder.Insert(SubscribersTable).
		Columns("id", "blogger_id", "subscriber_id").
		Values(subscribers.GetPointersArray()...)

	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("%s : %w", op, err)
	}

	_, err = r.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("%s : %w", op, err)
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
		return fmt.Errorf("%s : %w", op, err)
	}

	_, err = r.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("%s : %w", op, err)
	}

	return nil
}

func (r *UserRepository) DeleteUser(ctx context.Context, delInfo transfer.DeleteUserInfo) error {
	op := "UserRepository.DeleteUser"

	user, err := r.GetUserById(ctx, delInfo.Id)
	if err != nil {
		return fmt.Errorf("%s : %w", op, err)
	}

	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%s : %w", op, err)
	}
	defer func() { //TODO
		if err != nil {
			if txErr := tx.Rollback(); txErr != nil {
				err = fmt.Errorf("%s : %w", op, txErr)
			}
		}
	}()

	query := builder.Delete(UsersTable).Where(squirrel.Eq{"id": delInfo.Id})
	sql, args, err := query.ToSql()

	if _, err := tx.ExecContext(ctx, sql, args...); err != nil {
		return fmt.Errorf("%s : %w", op, err)
	}

	eventPayload, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("%s : %w", op, err)
	}

	insertBuilder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	insertQuery := insertBuilder.Insert(TransactionEventsTable).
		SetMap(map[string]interface{}{
			"event_id":   uuid.New(),
			"event_type": "userDeleted",
			"payload":    string(eventPayload),
		})

	sql, args, err = insertQuery.ToSql()
	if err != nil {
		return fmt.Errorf("%s : %w", op, err)
	}

	_, err = tx.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("%s : %w", op, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("%s : %w", op, err)
	}

	return nil
}

func (r *UserRepository) tryGetUserFromCache(ctx context.Context, id uuid.UUID) (*models.User, error) {
	op := "UserRepository.tryGetUserFromCache"

	data, err := r.cache.Get(ctx, fmt.Sprintf("%s%s", UsersCachePref, id.String()))
	if err != nil {
		return nil, fmt.Errorf("%s : %w", op, err)
	}

	var user *models.User

	if err := json.Unmarshal([]byte(data), &user); err != nil {
		return nil, fmt.Errorf("%s : %w", op, err)
	}

	return user, nil
}

func (r *UserRepository) SetUserToCache(ctx context.Context, user *models.User) error {
	op := "UserRepository.SetUserToCache"

	userJson, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("%s : %w", op, err)
	}

	err = r.cache.Set(ctx, fmt.Sprintf("%s%s", UsersCachePref, user.Id.String()), userJson)
	if err != nil {
		return fmt.Errorf("%s : %w", op, err)
	}

	return nil
}

func (r *UserRepository) DeleteUserFromCache(ctx context.Context, id uuid.UUID) error {
	op := "UserRepository.DeleteUserFromCache"

	err := r.cache.Delete(ctx, fmt.Sprintf("%s%s", UsersCachePref, id.String()))
	if err != nil {
		return fmt.Errorf("%s : %w", op, err)
	}

	return nil
}
