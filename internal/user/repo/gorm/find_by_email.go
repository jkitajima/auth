package gorm

import (
	"context"
	"fmt"
	"time"

	"auth/internal/user"
	"auth/pkg/otel"

	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

const FileFindByEmail = "find_by_email.go"

func (db *DB) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	const self = "FindByEmail"
	span := trace.SpanFromContext(ctx)

	var model UserModel
	result := db.Where("email = ?", email).First(&model)
	if result.Error != nil {
		switch result.Error {
		case gorm.ErrRecordNotFound:
			return nil, user.ErrNotFoundByEmail
		default:
			span.AddEvent("db query failed")
			db.logger.WarnContext(ctx, otel.FormatLog(Path, FileFindByEmail, self, user.ErrNotFoundByEmail.Error(), result.Error))
			return nil, user.ErrInternal
		}
	}
	db.logger.InfoContext(ctx, otel.FormatLog(Path, FileFindByEmail, self, fmt.Sprintf("found user with email %q", model.Email), nil))
	span.AddEvent(fmt.Sprintf("db query returned user_email %q", model.Email))

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
