package services_utils

import (
	"database/sql"
	"errors"
	"fmt"
)

func HandleErrInTransaction(err error, tx *sql.Tx) error {
	if err == nil {
		return nil
	}

	if txErr := tx.Rollback(); txErr != nil {
		return errors.Join(err, fmt.Errorf("error rolling back transaction: %v", txErr))
	}

	return err
}
