package gorm

import (
	"context"
	"fmt"

	"auth/internal/user"
	"auth/pkg/otel"

	"github.com/google/uuid"
)

const FileHardDeleteByID = "hard_delete_by_id.go"

func (db *DB) HardDeleteByID(ctx context.Context, id uuid.UUID) error {
	const self = "HardDeleteByID"

	model := UserModel{ID: id}
	result := db.Unscoped().Delete(&model)
	if result.Error != nil {
		db.logger.WarnContext(ctx, otel.FormatLog(Path, FileHardDeleteByID, self, "failed to hard delete user", result.Error))
		return user.ErrInternal
	}

	db.logger.InfoContext(ctx, otel.FormatLog(Path, FileHardDeleteByID, self, fmt.Sprintf("deleted user with id %q", model.ID.String()), nil))

	return nil
}
