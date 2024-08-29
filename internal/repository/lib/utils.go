package rep_utils

import (
	"github.com/KBcHMFollower/blog_user_service/internal/database"
)

const (
	FailedToGenerateSqlMessage = "failed to generate sql query"
	FailedToExecuteQuery       = "failed to execute query"
)

func GetExecutor(r database.DBWrapper, tx database.Transaction) database.Executor { //TODO: ПОДУМАТЬ ОБ ЭТОМ, ПИЗДАТЕНЬКО ВЫШЛО
	if tx == nil {
		return r
	}
	return tx
}
