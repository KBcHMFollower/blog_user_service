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
