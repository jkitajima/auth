package gorm

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserModel struct {
	ID                         uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4()"`
	Email                      string    `gorm:"not null;unique"`
	EmailVerified              bool      `gorm:"not null;default:false"`
	Password                   string    `gorm:"not null"`
	VerificationCode           *string
	VerificationCodeExpiration *int
	CreatedAt                  time.Time      `gorm:"not null"`
	UpdatedAt                  time.Time      `gorm:"not null"`
	DeletedAt                  gorm.DeletedAt `gorm:"index"`
}

func (*UserModel) TableName() string {
	return "User"
}
