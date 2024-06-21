package repository

import (
	"context"
	"github.com/KBcHMFollower/auth-service/database"
	"github.com/KBcHMFollower/auth-service/internal/domain/models"
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

const (
	USERS_TABLE_NAME = "users"
)

type CreateUserDto struct {
	Email    string
	HashPass []byte
	Fname    string
	Lname    string
}

type UserRepository struct {
	db database.DBWrapper
}

func NewUserRepository(dbDriver database.DBWrapper) (*UserRepository, error) {
	return &UserRepository{db: dbDriver}, nil
}

func (r *UserRepository) CreateUser(ctx context.Context, createDto *CreateUserDto) (uuid.UUID, error) {
	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	user := models.NewUserModel(createDto.Email, createDto.Fname, createDto.Lname, createDto.HashPass)

	query := builder.Insert(USERS_TABLE_NAME).
		SetMap(map[string]interface{}{
			"id":           user.Id,
			"email":        user.Email,
			"pass_hash":    user.PassHash,
			"fname":        user.Fname,
			"lname":        user.Lname,
			"created_date": user.CreatedDate,
			"updated_date": user.UpdatedDate,
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

	err = row.Scan(&user.Id, &user.Email, &user.Fname, &user.Lname, &user.PassHash, &user.CreatedDate, &user.UpdatedDate)
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

	err = row.Scan(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
