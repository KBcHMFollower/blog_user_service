package services_utils

import (
	"errors"
	"fmt"
	"github.com/KBcHMFollower/blog_user_service/internal/database"
)

func HandleErrInTransaction(err error, tx database.Transaction) error {
	if err == nil {
		return nil
	}

	if txErr := tx.Rollback(); txErr != nil {
		return errors.Join(err, fmt.Errorf("error rolling back transaction: %v", txErr))
	}

	return err
}

func ConvertMapKeysToStrings[T ~string](m map[T]any) map[string]any {
	result := make(map[string]any)
	for k, v := range m {
		result[string(k)] = v
	}

	return result
}
