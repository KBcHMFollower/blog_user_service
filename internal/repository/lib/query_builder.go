package rep_utils

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/KBcHMFollower/blog_user_service/internal/database"
	"github.com/Masterminds/squirrel"
	"reflect"
)

var (
	QBuilder = QueryBuilder{}
)

type QueryBuilder struct {
	squirrel.StatementBuilderType
}

type InsertBuilder struct {
	squirrel.InsertBuilder
}

type SelectBuilder struct {
	squirrel.SelectBuilder
}

type UpdateBuilder struct {
	squirrel.UpdateBuilder
}

func (qb QueryBuilder) PHFormat(f squirrel.PlaceholderFormat) QueryBuilder {
	return QueryBuilder{QBuilder.PlaceholderFormat(f)}
}

func (qb QueryBuilder) Ins(table string) InsertBuilder {
	return InsertBuilder{qb.Insert(table)}
}

func (qb QueryBuilder) Sel(table string) SelectBuilder {
	return SelectBuilder{qb.Select(table)}
}

func (qb QueryBuilder) Updt(table string) UpdateBuilder {
	return UpdateBuilder{qb.Update(table)}
}

func (ub UpdateBuilder) Wr(pred interface{}, args ...interface{}) UpdateBuilder {
	return UpdateBuilder{ub.Where(pred, args...)}
}

func (ub UpdateBuilder) St(column string, value interface{}) UpdateBuilder {
	return UpdateBuilder{ub.Set(column, value)}
}

func (ib InsertBuilder) SetModelMap(model any) InsertBuilder {
	modelMap := convertModelToMap(model)
	return InsertBuilder{ib.SetMap(modelMap)}
}

func (ib InsertBuilder) Cols(columns ...string) InsertBuilder {
	return InsertBuilder{ib.Columns(columns...)}
}

func (ib InsertBuilder) Vls(values ...interface{}) InsertBuilder {
	return InsertBuilder{ib.Values(values...)}
}

func (sb SelectBuilder) Wr(pred interface{}, args ...interface{}) SelectBuilder {
	return SelectBuilder{sb.Where(pred, args...)}
}

func (sb SelectBuilder) ExcCtx(ctx context.Context, executor database.Executor) (sql.Result, error) {
	toSql, args, err := sb.ToSql()
	if err != nil {
		return nil, err
	}

	return executor.ExecContext(ctx, toSql, args...)
}

func (ub UpdateBuilder) ExcCtx(ctx context.Context, executor database.Executor) (sql.Result, error) {
	toSql, args, err := ub.ToSql()
	if err != nil {
		return nil, err
	}

	return executor.ExecContext(ctx, toSql, args...)
}

func (ib InsertBuilder) ExcCtx(ctx context.Context, executor database.Executor) (sql.Result, error) {
	toSql, args, err := ib.ToSql()
	if err != nil {
		return nil, err
	}

	return executor.ExecContext(ctx, toSql, args...)
}

func (sb SelectBuilder) Frm(table string) SelectBuilder {
	return SelectBuilder{sb.From(table)}
}

func (sb SelectBuilder) Lim(limit uint64) SelectBuilder {
	return SelectBuilder{sb.Limit(limit)}
}

func (sb SelectBuilder) QryRowCtx(ctx context.Context, executor database.Executor, model any) (resErr error) {
	toSql, args, err := sb.ToSql()
	if err != nil {
		return err
	}

	row := executor.QueryRowContext(ctx, toSql, args...)
	if err := row.Scan(&model); err != nil {
		return err
	}

	val := reflect.ValueOf(model).Elem()

	columns := make([]interface{}, val.NumField())

	for i := 0; i < val.NumField(); i++ {
		columns[i] = val.Field(i).Addr().Interface()
	}

	if err := row.Scan(columns...); err != nil {
		return err
	}

	return nil
}

func (sb SelectBuilder) QryCtx(
	ctx context.Context,
	executor database.Executor,
	model interface{},
) (resErr error) {
	toSql, args, err := sb.ToSql()
	if err != nil {
		return err
	}

	rows, err := executor.QueryContext(ctx, toSql, args...)
	if err != nil {
		return err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			resErr = errors.Join(resErr, err)
		}
	}()

	modelValue := reflect.ValueOf(model)
	if modelValue.Kind() != reflect.Ptr || modelValue.Elem().Kind() != reflect.Slice {
		return fmt.Errorf("model should be a pointer to a slice")
	}

	sliceValue := modelValue.Elem()
	elemType := sliceValue.Type().Elem()

	for rows.Next() {
		newModel := reflect.New(elemType).Interface()

		val := reflect.ValueOf(newModel).Elem()
		typ := val.Type()

		fieldMap := make(map[string]reflect.Value)
		for i := 0; i < val.NumField(); i++ {
			field := typ.Field(i)
			dbTag := field.Tag.Get("db")
			if dbTag != "" {
				fieldMap[dbTag] = val.Field(i)
			}
		}

		columns, err := rows.Columns()
		if err != nil {
			return fmt.Errorf("failed to get columns: %w", err)
		}

		values := make([]interface{}, len(columns))
		for i, colName := range columns {
			if field, ok := fieldMap[colName]; ok {
				values[i] = field.Addr().Interface()
			} else {
				var discard interface{}
				values[i] = &discard
			}
		}

		if err := rows.Scan(values...); err != nil {
			return fmt.Errorf("error scanning row: %w", err)
		}

		sliceValue.Set(reflect.Append(sliceValue, reflect.ValueOf(newModel).Elem()))
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("rows iteration error: %w", err)
	}

	return nil
}

func convertModelToMap(model any) map[string]any {
	resMap := make(map[string]any)
	v := reflect.ValueOf(model)
	typeOfS := v.Type()

	for i := 0; i < v.NumField(); i++ {
		fieldValue := v.Field(i).Interface()
		if fieldValue == nil {
			continue
		}

		fieldName := typeOfS.Field(i).Tag.Get("db")
		if fieldName == "" {
			fieldName = typeOfS.Field(i).Name
		}

		resMap[fieldName] = fieldValue
	}

	return resMap
}
