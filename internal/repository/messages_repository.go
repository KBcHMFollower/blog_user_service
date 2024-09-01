package repository

import (
	"context"
	"github.com/KBcHMFollower/blog_user_service/internal/database"
	repositoriestransfer "github.com/KBcHMFollower/blog_user_service/internal/domain/layers_TOs/repositories"
	"github.com/KBcHMFollower/blog_user_service/internal/domain/models"
	reputils "github.com/KBcHMFollower/blog_user_service/internal/repository/lib"
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

const (
	messagesTable = "amqp_messages"
)

const (
	MessagesStatusCol    = "status"
	messagesEventIdCol   = "event_id"
	messagesAllCol       = "*"
	messagesEventTypeCol = "event_type"
	messagesPayloadCol   = "payload"
)

type EventFilter struct {
}

type EventRepository struct {
	db       database.DBWrapper
	qBuilder squirrel.StatementBuilderType
}

func NewEventRepository(dbDriver database.DBWrapper) *EventRepository {
	return &EventRepository{db: dbDriver, qBuilder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)}
}

func (r *EventRepository) Events(ctx context.Context, info repositoriestransfer.GetEventsInfo) ([]*models.EventInfo, error) {
	info.Page, info.Size = reputils.GetPageAndSize(info.Page, info.Size)

	offset := (info.Page - 1) * info.Size

	query := r.qBuilder.
		Select(messagesAllCol).
		From(messagesTable).
		Where(squirrel.Eq(reputils.ConvertMapKeysToStrings(info.Condition))).
		Limit(info.Size).
		Offset(offset)

	toSql, args, err := query.ToSql()
	if err != nil {
		return nil, reputils.ReturnGenerateSqlError(ctx, err)
	}

	eventInfos := make([]*models.EventInfo, 0)
	if err := r.db.SelectContext(ctx, &eventInfos, toSql, args...); err != nil {
		return nil, reputils.ReturnExecuteSqlError(ctx, err)
	}

	return eventInfos, nil
}

//func (r *EventRepository) EventsWithStatus(ctx context.Context, status repositories_transfer.MessageStatus, limit uint64) ([]*models.EventInfo, error) {
//	query := r.qBuilder.
//		Select(messagesAllCol).
//		From(messagesTable).
//		Where(squirrel.Eq{MessagesStatusCol: status}).
//		Limit(limit)
//
//	toSql, args, err := query.ToSql()
//	if err != nil {
//		return nil, rep_utils.ReturnGenerateSqlError(ctx, err)
//	}
//
//	eventInfos := make([]*models.EventInfo, 0)
//	if err := r.db.SelectContext(ctx, &eventInfos, toSql, args...); err != nil {
//		return nil, rep_utils.ReturnExecuteSqlError(ctx, err)
//	}
//
//	return eventInfos, nil
//}

func (r *EventRepository) Event(ctx context.Context, eventId uuid.UUID) (*models.EventInfo, error) {
	query := r.qBuilder.
		Select(messagesAllCol).
		From(messagesTable).
		Where(squirrel.Eq{messagesEventIdCol: eventId})

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, reputils.ReturnGenerateSqlError(ctx, err)
	}

	var eventInfo models.EventInfo
	if err := r.db.GetContext(ctx, &eventInfo, sql, args...); err != nil {
		return nil, reputils.ReturnExecuteSqlError(ctx, err)
	}

	return &eventInfo, nil
}

func (r *EventRepository) Create(ctx context.Context, info repositoriestransfer.CreateEventInfo, tx database.Transaction) error {
	executor := reputils.GetExecutor(r.db, tx)

	query := r.qBuilder.
		Insert(messagesTable).
		SetMap(map[string]interface{}{
			messagesEventIdCol:   info.EventId,
			messagesPayloadCol:   info.Payload,
			messagesEventTypeCol: info.EventType,
		})

	toSql, args, err := query.ToSql()
	if err != nil {
		return reputils.ReturnGenerateSqlError(ctx, err)
	}

	if _, err := executor.ExecContext(ctx, toSql, args...); err != nil {
		return reputils.ReturnExecuteSqlError(ctx, err)
	}

	return nil
}

func (r *EventRepository) Update(ctx context.Context, info repositoriestransfer.UpdateEventInfo) error {
	query := r.qBuilder.
		Update(messagesTable).
		Where(squirrel.Eq{messagesEventIdCol: info.EventId}).
		SetMap(info.UpdateData)

	sql, args, err := query.ToSql()
	if err != nil {
		return reputils.ReturnGenerateSqlError(ctx, err)
	}

	if _, err = r.db.ExecContext(ctx, sql, args...); err != nil {
		return reputils.ReturnExecuteSqlError(ctx, err)
	}

	return nil
}

func (r *EventRepository) UpdateMany(ctx context.Context, info repositoriestransfer.UpdateManyEventInfo) error {
	query := r.qBuilder.
		Update(messagesTable).
		Where(squirrel.Eq{messagesEventIdCol: info.EventId}).
		SetMap(info.UpdateData)

	sql, args, err := query.ToSql()
	if err != nil {
		return reputils.ReturnGenerateSqlError(ctx, err)
	}

	if _, err = r.db.ExecContext(ctx, sql, args...); err != nil {
		return reputils.ReturnExecuteSqlError(ctx, err)
	}

	return nil
}
