package repository

import (
	"context"
	"fmt"
	"github.com/KBcHMFollower/blog_user_service/database"
	"github.com/KBcHMFollower/blog_user_service/internal/domain/models"
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type EventFilter struct {
}

type EventRepository struct {
	db database.DBWrapper
}

func NewEventRepository(dbDriver database.DBWrapper) (*EventRepository, error) {
	return &EventRepository{db: dbDriver}, nil
}

func (r *EventRepository) GetEvents(ctx context.Context, filterTarget string, filterValue interface{}, limit uint64) ([]*models.EventInfo, error) {
	op := "UserRepository.getSubInfo"

	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	eventInfos := make([]*models.EventInfo, 0)

	query := builder.
		Select("*").
		From("transaction_events").
		Where(squirrel.Eq{filterTarget: filterValue}).
		Limit(limit)

	sql, args, err := query.ToSql()
	if err != nil {
		return eventInfos, fmt.Errorf("%s : %w", op, err)
	}

	rows, err := r.db.QueryContext(ctx, sql, args...)
	if err != nil {
		return eventInfos, fmt.Errorf("%s : %w", op, err)
	}
	defer rows.Close()

	for rows.Next() {
		var eventInfo models.EventInfo

		err := rows.Scan(eventInfo.GetPointersArray()...)
		if err != nil {
			return eventInfos, fmt.Errorf("error in parse post from db: %v", err)
		}

		eventInfos = append(eventInfos, &eventInfo)
	}

	return eventInfos, nil
}

func (r *EventRepository) SetSentStatusesInEvents(ctx context.Context, eventsId []uuid.UUID) error {
	op := "UserRepository.getSubInfo"

	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	query := builder.
		Update("transaction_events").
		Where(squirrel.Eq{"event_id": eventsId}).
		Set("status", "sent")

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
