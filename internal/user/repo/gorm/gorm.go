package gorm

import (
	"log/slog"

	"auth/internal/user"

	"gorm.io/gorm"
)

const (
	Path string = "auth/internal/user/repo/gorm"
)

type DB struct {
	*gorm.DB
	logger *slog.Logger
}

func NewRepo(db *gorm.DB, logger *slog.Logger) user.Repoer {
	return &DB{db, logger}
}
