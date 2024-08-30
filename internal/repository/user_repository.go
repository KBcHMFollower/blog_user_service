package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/KBcHMFollower/blog_user_service/internal/clients/cashe"
	"github.com/KBcHMFollower/blog_user_service/internal/database"
	ctxerrors "github.com/KBcHMFollower/blog_user_service/internal/domain/errors"
	transfer "github.com/KBcHMFollower/blog_user_service/internal/domain/layers_TOs/repositories"
	"github.com/KBcHMFollower/blog_user_service/internal/lib"
	rep_utils "github.com/KBcHMFollower/blog_user_service/internal/repository/lib"
	"time"

	"github.com/KBcHMFollower/blog_user_service/internal/domain/models"
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

const (
	usersTable = "users"
)

const (
	UsersCachePref = "userId-"
)

const (
	usersIdCol          = "user_id"
	userEmailCol        = "email"
	usersPassHashCol    = "pass_hash"
	usersAvatarCol      = "avatar"
	usersAvatarMiniCol  = "avatar_min"
	usersFNameCol       = "fname"
	usersLNameCol       = "lname"
	usersCreatedDateCol = "created_date"
	usersUpdatedDateCol = "updated_date"
)

var (
	conditionMap = map[transfer.UserFieldTarget]string{
		transfer.UserIdCondition:    usersIdCol,
		transfer.UserEmailCondition: userEmailCol,
	}

	notUpdatableCols = []string{usersIdCol, usersPassHashCol, usersCreatedDateCol, usersUpdatedDateCol}
)

type UserRepository struct {
	db       database.DBWrapper
	qBuilder squirrel.StatementBuilderType
	cache    cashe.CasheStorage
}

func NewUserRepository(dbDriver database.DBWrapper, cacheStorage cashe.CasheStorage) *UserRepository {
	return &UserRepository{db: dbDriver, cache: cacheStorage, qBuilder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)}
}

func (r *UserRepository) Create(ctx context.Context, createDto *transfer.CreateUserInfo, tx database.Transaction) (uuid.UUID, error) {
	user := models.NewUserModel(createDto.Email, createDto.FName, createDto.LName, createDto.HashPass)

	executor := rep_utils.GetExecutor(r.db, tx)

	query := r.qBuilder.Insert(usersTable).
		SetMap(map[string]interface{}{
			usersIdCol:         user.Id,
			userEmailCol:       user.Email,
			usersPassHashCol:   user.PassHash,
			usersAvatarCol:     user.Avatar,
			usersAvatarMiniCol: user.AvatarMin,
			usersFNameCol:      user.FName,
			usersLNameCol:      user.LName,
		}).
		Suffix("RETURNING \"id\"")

	sql, args, err := query.ToSql()
	if err != nil {
		return uuid.Nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap(rep_utils.FailedToGenerateSqlMessage, err))
	}

	var id uuid.UUID
	if err := executor.GetContext(ctx, &id, sql, args...); err != nil {
		return uuid.Nil, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap(rep_utils.FailedToExecuteQuery, err))
	}

	return id, nil
}

func (r *UserRepository) User(ctx context.Context, condition transfer.GetUserInfo, tx database.Transaction) (*models.User, error) {
	executor := rep_utils.GetExecutor(r.db, tx)

	sql, args, err := r.qBuilder.Select("*").
		From(usersTable).
		Where(squirrel.Eq{conditionMap[condition.Target]: condition.Value}).
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

func (r *UserRepository) Update(ctx context.Context, updateData transfer.UpdateUserInfo, tx database.Transaction) error { //TODO
	executor := rep_utils.GetExecutor(r.db, tx)

	query := r.qBuilder.
		Update(usersTable).
		Where(squirrel.Eq{usersIdCol: updateData.Id}).
		Set("updated_date", time.Now())

	for _, item := range updateData.UpdateInfo {
		if lib.Contains(notUpdatableCols, item.Name) {
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

func (r *UserRepository) Delete(ctx context.Context, delInfo transfer.DeleteUserInfo, tx database.Transaction) error {
	executor := rep_utils.GetExecutor(r.db, tx)

	query := r.qBuilder.Delete(usersTable).
		Where(squirrel.Eq{usersIdCol: delInfo.Id})

	sql, args, err := query.ToSql()
	if err != nil {
		return ctxerrors.WrapCtx(ctx, ctxerrors.Wrap(rep_utils.FailedToGenerateSqlMessage, err))
	}

	if _, err := executor.ExecContext(ctx, sql, args...); err != nil {
		return ctxerrors.WrapCtx(ctx, ctxerrors.Wrap(rep_utils.FailedToExecuteQuery, err))
	}

	return nil
}

func (r *UserRepository) RollBackUser(ctx context.Context, user models.User) error {
	query := r.qBuilder.Insert(usersTable).
		SetMap(map[string]interface{}{
			usersIdCol:          user.Id,
			userEmailCol:        user.Email,
			usersPassHashCol:    user.PassHash,
			usersAvatarCol:      user.Avatar,
			usersAvatarMiniCol:  user.AvatarMin,
			usersFNameCol:       user.FName,
			usersLNameCol:       user.LName,
			usersCreatedDateCol: user.CreatedDate,
			usersUpdatedDateCol: time.Now(),
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

func (r *UserRepository) TryGetFromCache(ctx context.Context, id uuid.UUID) (*models.User, error) {
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

func (r *UserRepository) SetToCache(ctx context.Context, user *models.User) error {
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

func (r *UserRepository) DeleteFromCache(ctx context.Context, id uuid.UUID) error {
	err := r.cache.Delete(ctx, fmt.Sprintf("%s%s", UsersCachePref, id.String()))
	if err != nil {
		return ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("failed to delete from cache", err))
	}

	return nil
}
