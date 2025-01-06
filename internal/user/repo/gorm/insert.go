package gorm

import (
	"context"
	"fmt"

	"auth/internal/user"
	"auth/pkg/otel"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

const FileInsert = "insert.go"

func (db *DB) Insert(ctx context.Context, u *user.User) error {
	const self = "Insert"

	var expiration int
	if u.VerificationCodeExpiration != nil {
		expiration = int(u.VerificationCodeExpiration.Unix())
	}

	expptr := &expiration
	if expiration == 0 {
		expptr = nil
	}

	model := &UserModel{
		ID:                         u.ID,
		Email:                      u.Email,
		EmailVerified:              u.EmailVerified,
		Password:                   u.Password,
		VerificationCode:           u.VerificationCode,
		VerificationCodeExpiration: expptr,
		CreatedAt:                  u.CreatedAt,
		UpdatedAt:                  u.UpdatedAt,
	}

	if u.DeletedAt != nil {
		model.DeletedAt = gorm.DeletedAt{
			Time:  *u.DeletedAt,
			Valid: true,
		}
	}

	result := db.Create(model)
	if result.Error != nil {
		db.logger.WarnContext(ctx, otel.FormatLog(Path, FileInsert, self, "failed to create user", result.Error))
		err := result.Error.(*pgconn.PgError)
		switch err.Code {
		case "23505":
			return user.ErrEmailAlreadyInUse
		default:
			return user.ErrInternal
		}
	}
	db.logger.InfoContext(ctx, otel.FormatLog(Path, FileInsert, self, fmt.Sprintf("created a new user with id %q", model.ID.String()), nil))

	u.ID = model.ID
	u.EmailVerified = model.EmailVerified
	u.CreatedAt = model.CreatedAt
	u.UpdatedAt = model.UpdatedAt

	return nil
}
