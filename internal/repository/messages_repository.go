package repository

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/KBcHMFollower/blog_user_service/internal/database"
	repositories_transfer "github.com/KBcHMFollower/blog_user_service/internal/domain/layers_TOs/repositories"
	"github.com/KBcHMFollower/blog_user_service/internal/domain/models"
	rep_utils "github.com/KBcHMFollower/blog_user_service/internal/repository/lib"
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

const (
	messagesTable = "amqp_messages"
)

const (
	messagesStatusCol    = "status"
	messagesEventIdCol   = "event_id"
	messagesAllCol       = "*"
	messagesEventTypeCol = "event_type"
	messagesPayloadCol   = "payload"
)

const (
	SentStatus = "sent"
)

type EventFilter struct {
}

type EventRepository struct {
	db database.DBWrapper
}

func NewEventRepository(dbDriver database.DBWrapper) *EventRepository {
	return &EventRepository{db: dbDriver}
}

func (r *EventRepository) GetEvents(ctx context.Context, filterTarget string, filterValue interface{}, limit uint64) ([]*models.EventInfo, error) {
	op := "UserRepository.getSubInfo"

	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	eventInfos := make([]*models.EventInfo, 0)

	query := builder.
		Select(messagesAllCol).
		From(messagesTable).
		Where(squirrel.Eq{filterTarget: filterValue}).
		Limit(limit)

	toSql, args, err := query.ToSql()
	if err != nil {
		return eventInfos, fmt.Errorf("%s : failed to generate sql query: %w", op, err)
	}

	rows, err := r.db.QueryContext(ctx, toSql, args...)
	if err != nil {
		return eventInfos, fmt.Errorf("%s : failed to fetch events: %w", op, err)
	}
	defer rows.Close()

	for rows.Next() {
		var eventInfo models.EventInfo

		err := rows.Scan(eventInfo.GetPointersArray()...)
		if err != nil {
			return eventInfos, fmt.Errorf("%s : error in parse post from db: %w", op, err)
		}

		eventInfos = append(eventInfos, &eventInfo)
	}

	return eventInfos, nil
}

func (r *EventRepository) SetSentStatusesInEvents(ctx context.Context, eventsId []uuid.UUID) error {
	op := "UserRepository.getSubInfo"

	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	query := builder.
		Update(messagesTable).
		Where(squirrel.Eq{messagesEventIdCol: eventsId}).
		Set(messagesStatusCol, "sent")

	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("%s : failed to fetch events %w", op, err)
	}

	_, err = r.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("%s : failed to set events sent status: %w", op, err)
	}

	return nil
}

func (r *EventRepository) GetEventById(ctx context.Context, eventId uuid.UUID) (*models.EventInfo, error) {
	op := "UserRepository.getEventById"

	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	query := builder.
		Select(messagesAllCol).
		From(messagesTable).
		Where(squirrel.Eq{messagesEventIdCol: eventId})

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s : failed to generate sql query: %w", op, err)
	}

	row := r.db.QueryRowContext(ctx, sql, args...)
	if err := row.Err(); err != nil {
		return nil, fmt.Errorf("%s : failed to fetch event: %w", op, err)
	}

	var eventInfo models.EventInfo
	if err := row.Scan(eventInfo.GetPointersArray()...); err != nil {
		return nil, fmt.Errorf("%s : error in parse post from db: %w", op, err)
	}

	return &eventInfo, nil
}

func (r *EventRepository) Create(ctx context.Context, info repositories_transfer.CreateEventInfo, tx *sql.Tx) error {
	op := "UserRepository.create"

	executor := rep_utils.GetExecutor(r.db, tx)
	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	query := builder.
		Insert(messagesTable).
		Columns(messagesEventIdCol, messagesEventTypeCol, messagesPayloadCol).
		Values(info.EventId, info.EventType, info.Payload) //TODO: ВОЗМОЖНО СТОИТ С ЭТИМ ЧТО-ТО ПРИДУМАТЬ, ПОТОМУ ЧТО СЕЙЧАТ, ЕСЛИ МЕНЯЕТСЯ МОДЕЛЬ, ПРИЙДЕТСЯ ИДТИ СЮДА И МЕНЯТЬ, ПОПРОБУЮ МАПУ

	toSql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("%s : failed to generate sql query: %w", op, err)
	}

	row := executor.QueryRowContext(ctx, toSql, args...)
	if row.Err() != nil {
		return fmt.Errorf("%s : error in parse post from db: %w", op, row.Err())
	}

	return nil
}
