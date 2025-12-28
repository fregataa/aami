package repository

import (
	"errors"
	"fmt"
	"strings"

	domainerrors "github.com/fregataa/aami/config-server/internal/errors"
	"gorm.io/gorm"
)

// fromGormError converts GORM errors to domain errors
// This function is internal to the repository layer and handles the translation
// between infrastructure (GORM) errors and domain errors.
func fromGormError(err error) error {
	if err == nil {
		return nil
	}

	// Check for GORM's ErrRecordNotFound
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domainerrors.ErrNotFound
	}

	// Convert error to string for pattern matching
	errStr := err.Error()

	// PostgreSQL error patterns
	// SQLSTATE 23505 - unique_violation
	if strings.Contains(errStr, "duplicate key") ||
		strings.Contains(errStr, "UNIQUE constraint") ||
		strings.Contains(errStr, "violates unique constraint") {
		return domainerrors.ErrDuplicateKey
	}

	// SQLSTATE 23503 - foreign_key_violation
	if strings.Contains(errStr, "foreign key constraint") ||
		strings.Contains(errStr, "violates foreign key") ||
		strings.Contains(errStr, "FOREIGN KEY constraint") {
		return domainerrors.ErrForeignKeyViolation
	}

	// SQLSTATE 23xxx - integrity_constraint_violation (generic)
	if strings.Contains(errStr, "constraint") &&
		(strings.Contains(errStr, "violates") || strings.Contains(errStr, "violation")) {
		return domainerrors.ErrConstraintViolation
	}

	// Wrap unknown database errors
	return fmt.Errorf("%w: %v", domainerrors.ErrDatabaseError, err)
}
