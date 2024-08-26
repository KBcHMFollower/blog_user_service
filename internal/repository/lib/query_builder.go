package rep_utils

import (
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

func (qb *QueryBuilder) PHFormat(f squirrel.PlaceholderFormat) QueryBuilder {
	return QueryBuilder{QBuilder.PlaceholderFormat(f)}
}

func (qb *QueryBuilder) ModelInsert(table string) InsertBuilder {
	return InsertBuilder{qb.Insert(table)}
}

func (qb *QueryBuilder) ModelSelect(typeS reflect.Type) squirrel.SelectBuilder {
	fieldNames := getModelFieldNames(typeS)

	return qb.Select(fieldNames...)
}

func (ib *InsertBuilder) SetModelMap(model any) squirrel.InsertBuilder {
	modelMap := convertModelToMap(model)
	return ib.SetMap(modelMap)
}

func getModelFieldNames(model reflect.Type) []string {
	result := make([]string, 0)

	for i := 0; i < model.NumField(); i++ {
		fieldName := model.Field(i).Tag.Get("db")
		if fieldName != "" {
			fieldName = model.Field(i).Name
		}

		result = append(result, fieldName)
	}

	return result
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
