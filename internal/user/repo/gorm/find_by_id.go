package gorm

import (
	"context"
	"fmt"
	"time"

	"auth/internal/user"
	"auth/pkg/otel"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

const FileFindByID = "find_by_id.go"

func (db *DB) FindByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	const self = "FindByID"
	span := trace.SpanFromContext(ctx)

	var model UserModel
	result := db.First(&model, "id = ?", id.String())
	if result.Error != nil {
		switch result.Error {
		case gorm.ErrRecordNotFound:
			return nil, user.ErrNotFoundByID
		default:
			span.AddEvent("db query failed")
			db.logger.WarnContext(ctx, otel.FormatLog(Path, FileFindByID, self, user.ErrNotFoundByID.Error(), result.Error))
			return nil, user.ErrInternal
		}
	}
	db.logger.InfoContext(ctx, otel.FormatLog(Path, FileFindByID, self, fmt.Sprintf("found user with id %q", model.ID.String()), nil))
	span.AddEvent(fmt.Sprintf("db query returned user_id %q", id.String()))

	expiration := model.VerificationCodeExpiration
	var unixts *time.Time
	if expiration != nil {
		t := time.Unix(int64(*model.VerificationCodeExpiration), 0)
		unixts = &t
	}
	user := user.User{
		ID:                         model.ID,
		Email:                      model.Email,
		EmailVerified:              model.EmailVerified,
		Password:                   model.Password,
		VerificationCode:           model.VerificationCode,
		VerificationCodeExpiration: unixts,
		CreatedAt:                  model.CreatedAt,
		UpdatedAt:                  model.UpdatedAt,
	}
	return &user, nil
}
