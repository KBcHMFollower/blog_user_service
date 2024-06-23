package repository

import (
	"context"
	"fmt"
	"github.com/KBcHMFollower/blog_user_service/database"
	"github.com/KBcHMFollower/blog_user_service/internal/domain/models"
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"time"
)

const (
	USERS_TABLE_NAME       = "users"
	SUBSCRIBERS_TABLE_NAME = "subscribers"
)

type CreateUserDto struct {
	Email    string
	HashPass []byte
	FName    string
	LName    string
}

type UserRepository struct {
	db database.DBWrapper
}

func NewUserRepository(dbDriver database.DBWrapper) (*UserRepository, error) {
	return &UserRepository{db: dbDriver}, nil
}

func (r *UserRepository) getSubInfo(ctx context.Context, userId uuid.UUID, page uint64, size uint64, targetType string) ([]*models.Subscriber, uint32, error) {
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
		return subscribers, 0, fmt.Errorf("error in generate sql-query : %v", err)
	}

	rows, err := r.db.QueryContext(ctx, sql, args...)
	if err != nil {
		return subscribers, 0, fmt.Errorf("error in quey for db: %v", err)
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
		return subscribers, 0, fmt.Errorf("error in generate sql-query : %v", err)
	}

	var totalCount uint32

	countRow := r.db.QueryRowContext(ctx, sql, args...)
	if err := countRow.Scan(&totalCount); err != nil {
		return nil, 0, fmt.Errorf("can`t scan properties from db : %v", err)
	}

	return subscribers, totalCount, nil
}

func (r *UserRepository) CreateUser(ctx context.Context, createDto *CreateUserDto) (uuid.UUID, error) {
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
		return uuid.Nil, err
	}

	row := r.db.QueryRowContext(ctx, sql, args...)

	var id uuid.UUID

	err = row.Scan(&id)
	if err != nil {
		return uuid.Nil, err
	}

	return id, nil
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	sql, args, err := builder.Select("*").
		From(USERS_TABLE_NAME).
		Where(squirrel.Eq{"email": email}).
		ToSql()
	if err != nil {
		return nil, err
	}

	var user models.User

	row := r.db.QueryRowContext(ctx, sql, args...)

	err = row.Scan(user.GetPointersArray()...)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) GetUserById(ctx context.Context, userId uuid.UUID) (*models.User, error) {
	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	sql, args, err := builder.Select("*").
		From(USERS_TABLE_NAME).
		Where(squirrel.Eq{"id": userId}).
		ToSql()
	if err != nil {
		return nil, err
	}

	var user models.User

	row := r.db.QueryRowContext(ctx, sql, args...)

	err = row.Scan(user.GetPointersArray()...)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) GetUserSubscribers(ctx context.Context, userId uuid.UUID, page uint64, size uint64) ([]*models.User, uint32, error) {
	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	users := make([]*models.User, 0)

	subscribers, totalCount, err := r.getSubInfo(ctx, userId, page, size, "blogger_id")
	if err != nil {
		return users, totalCount, err
	}

	subscribersId := make([]uuid.UUID, 0)
	for _, subscriber := range subscribers {
		subscribersId = append(subscribersId, subscriber.SubscriberId)
	}

	sql, args, err := builder.Select("*").
		From(USERS_TABLE_NAME).
		Where(squirrel.Eq{"id": subscribersId}).
		ToSql()

	rows, err := r.db.QueryContext(ctx, sql, args...)
	if err != nil {
		return users, 0, fmt.Errorf("error in generate sql-query : %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var user models.User
		err = rows.Scan(user.GetPointersArray()...)
		if err != nil {
			return users, 0, fmt.Errorf("error in parse post from db : %v", err)
		}
		users = append(users, &user)
	}

	return users, totalCount, nil
}

func (r *UserRepository) GetUserSubscriptions(ctx context.Context, userId uuid.UUID, page uint64, size uint64) ([]*models.User, uint32, error) {
	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	users := make([]*models.User, 0)

	subscribers, totalCount, err := r.getSubInfo(ctx, userId, page, size, "subscriber_id")
	if err != nil {
		return users, totalCount, err
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
		return users, 0, fmt.Errorf("error in generate sql-query : %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var user models.User
		err = rows.Scan(user.GetPointersArray()...)
		if err != nil {
			return users, 0, fmt.Errorf("error in parse post from db : %v", err)
		}
		users = append(users, &user)
	}

	return users, totalCount, nil
}

func (r *UserRepository) UpdateUser(ctx context.Context, updateData UpdateData) (*models.User, error) {
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
		return nil, fmt.Errorf("error in generate sql-query : %v", err)
	}

	_, err = r.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("error in execute sql-query : %v", err)
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
		return nil, fmt.Errorf("error scanning updated post : %v", err)
	}

	return &user, nil
}

func (r *UserRepository) Subscribe(ctx context.Context, bloggerId uuid.UUID, subscriberId uuid.UUID) error {
	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	subscribers := models.NewSubscriber(bloggerId, subscriberId)

	query := builder.Insert(SUBSCRIBERS_TABLE_NAME).
		Columns("id", "blogger_id", "subscriber_id").
		Values(subscribers.GetPointersArray()...)

	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("error in generate sql-query : %v", err)
	}

	_, err = r.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("error in execute sql-query : %v", err)
	}

	return nil
}

func (r *UserRepository) Unsubscribe(ctx context.Context, bloggerId uuid.UUID, subscriberId uuid.UUID) error {
	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	query := builder.Delete(SUBSCRIBERS_TABLE_NAME).
		Where(squirrel.Eq{"blogger_id": bloggerId, "subscriber_id": subscriberId})

	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("error in generate sql-query : %v", err)
	}

	_, err = r.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("error in execute sql-query : %v", err)
	}

	return nil
}

func (r *UserRepository) DeleteUser(ctx context.Context, userId uuid.UUID) error {
	_, err := r.UpdateUser(ctx, UpdateData{
		Id: userId,
		UpdateInfo: []*UpdateInfo{
			&UpdateInfo{
				Name:  "is_deleted",
				Value: "true",
			},
		},
	})

	if err != nil {
		return fmt.Errorf("error in delete user : %v", err)
	}

	return nil
}
