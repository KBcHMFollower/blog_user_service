package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/KBcHMFollower/blog_user_service/internal/cashe"
	"time"

	"github.com/KBcHMFollower/blog_user_service/database"
	"github.com/KBcHMFollower/blog_user_service/internal/domain/models"
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

const (
	USERS_TABLE_NAME       = "users"
	SUBSCRIBERS_TABLE_NAME = "subscribers"
	TE_TABLE_NAME          = "transaction_events"

	USERS_CASHE_PREFIX = "userId-"
)

type CreateUserDto struct {
	Email    string
	HashPass []byte
	FName    string
	LName    string
}

type UserRepository struct {
	db    database.DBWrapper
	cashe cashe.CasheStorage
}

func NewUserRepository(dbDriver database.DBWrapper, casheStorage cashe.CasheStorage) (*UserRepository, error) {
	return &UserRepository{db: dbDriver, cashe: casheStorage}, nil
}

func (r *UserRepository) getSubInfo(ctx context.Context, userId uuid.UUID, page uint64, size uint64, targetType string) ([]*models.Subscriber, uint32, error) {
	op := "UserRepository.getSubInfo"

	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	subscribers := make([]*models.Subscriber, 0)

	offset := (page - 1) * size

	query := builder.
		Select("*").
		From(SUBSCRIBERS_TABLE_NAME).
		Where(squirrel.Eq{targetType: userId}).
		Limit(size).
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
		From(SUBSCRIBERS_TABLE_NAME).
		Where(squirrel.Eq{targetType: userId})

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

func (r *UserRepository) CreateUser(ctx context.Context, createDto *CreateUserDto) (uuid.UUID, error) {
	op := "UserRepository.CreateUser"

	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	user := models.NewUserModel(createDto.Email, createDto.FName, createDto.LName, createDto.HashPass)

	query := builder.Insert(USERS_TABLE_NAME).
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
		From(USERS_TABLE_NAME).
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

	if user, err := r.tryGetUserFromCashe(ctx, userId); err == nil {
		return user, nil
	}

	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	sql, args, err := builder.Select("*").
		From(USERS_TABLE_NAME).
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

	if err := r.SetUserToCashe(ctx, &user); err != nil {
		return &user, err
	}

	return &user, nil
}

func (r *UserRepository) GetUserSubscribers(ctx context.Context, userId uuid.UUID, page uint64, size uint64) ([]*models.User, uint32, error) {
	op := "UserRepository.GetUserSubscribers"

	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	users := make([]*models.User, 0)

	subscribers, totalCount, err := r.getSubInfo(ctx, userId, page, size, "blogger_id")
	if err != nil {
		return users, totalCount, fmt.Errorf("%s : %w", op, err)
	}

	subscribersId := make([]uuid.UUID, 0)
	for _, subscriber := range subscribers {
		subscribersId = append(subscribersId, subscriber.SubscriberId)
	}

	sql, args, err := builder.Select("*").
		From(USERS_TABLE_NAME).
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

func (r *UserRepository) GetUserSubscriptions(ctx context.Context, userId uuid.UUID, page uint64, size uint64) ([]*models.User, uint32, error) {
	op := "UserRepository.GetUserSubscribers"

	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	users := make([]*models.User, 0)

	subscribers, totalCount, err := r.getSubInfo(ctx, userId, page, size, "subscriber_id")
	if err != nil {
		return users, totalCount, fmt.Errorf("%s : %w", op, err)
	}

	subscribersId := make([]uuid.UUID, 0)
	for _, subscriber := range subscribers {
		subscribersId = append(subscribersId, subscriber.BloggerId)
	}

	sql, args, err := builder.Select("*").
		From(USERS_TABLE_NAME).
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

func (r *UserRepository) UpdateUser(ctx context.Context, updateData UpdateData) (*models.User, error) {
	op := "UserRepository.UpdateUser"
	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	query := builder.
		Update(USERS_TABLE_NAME).
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
		From(USERS_TABLE_NAME).
		Where(squirrel.Eq{"id": updateData.Id})
	sqlGetPost, argsGetPost, _ := queryGetPost.ToSql()

	row := r.db.QueryRowContext(ctx, sqlGetPost, argsGetPost...)

	var user models.User
	err = row.Scan(user.GetPointersArray()...)
	if err != nil {
		return nil, fmt.Errorf("%s : %w", op, err)
	}

	if err := r.DeleteUserFromCashe(ctx, updateData.Id); err != nil {
		return &user, nil
	}

	return &user, nil
}

func (r *UserRepository) Subscribe(ctx context.Context, bloggerId uuid.UUID, subscriberId uuid.UUID) error {
	op := "UserRepository.Subscribe"
	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	subscribers := models.NewSubscriber(bloggerId, subscriberId)

	query := builder.Insert(SUBSCRIBERS_TABLE_NAME).
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

func (r *UserRepository) Unsubscribe(ctx context.Context, bloggerId uuid.UUID, subscriberId uuid.UUID) error {
	op := "UserRepository.Unsubscribe"
	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	query := builder.Delete(SUBSCRIBERS_TABLE_NAME).
		Where(squirrel.Eq{"blogger_id": bloggerId, "subscriber_id": subscriberId})

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

func (r *UserRepository) DeleteUser(ctx context.Context, userId uuid.UUID) error {
	op := "UserRepository.DeleteUser"

	user, err := r.GetUserById(ctx, userId)
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

	query := builder.Delete(USERS_TABLE_NAME).Where(squirrel.Eq{"id": userId})
	sql, args, err := query.ToSql()

	if _, err := tx.ExecContext(ctx, sql, args...); err != nil {
		return fmt.Errorf("%s : %w", op, err)
	}

	eventPayload, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("%s : %w", op, err)
	}

	insertBuilder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	insertQuery := insertBuilder.Insert(TE_TABLE_NAME).
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

func (r *UserRepository) tryGetUserFromCashe(ctx context.Context, id uuid.UUID) (*models.User, error) {
	op := "UserRepository.tryGetUserFromCashe"

	data, err := r.cashe.Get(ctx, fmt.Sprintf("%s%s", USERS_CASHE_PREFIX, id.String()))
	if err != nil {
		return nil, fmt.Errorf("%s : %w", op, err)
	}

	var user *models.User

	if err := json.Unmarshal([]byte(data), &user); err != nil {
		return nil, fmt.Errorf("%s : %w", op, err)
	}

	return user, nil
}

func (r *UserRepository) SetUserToCashe(ctx context.Context, user *models.User) error {
	op := "UserRepository.SetUserToCashe"

	userJson, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("%s : %w", op, err)
	}

	err = r.cashe.Set(ctx, fmt.Sprintf("%s%s", USERS_CASHE_PREFIX, user.Id.String()), userJson)
	if err != nil {
		return fmt.Errorf("%s : %w", op, err)
	}

	return nil
}

func (r *UserRepository) DeleteUserFromCashe(ctx context.Context, id uuid.UUID) error {
	op := "UserRepository.DeleteUserFromCashe"

	err := r.cashe.Delete(ctx, fmt.Sprintf("%s%s", USERS_CASHE_PREFIX, id.String()))
	if err != nil {
		return fmt.Errorf("%s : %w", op, err)
	}

	return nil
}
