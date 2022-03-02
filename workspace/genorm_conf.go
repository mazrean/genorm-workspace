//go:build genorm
// +build genorm

//go:generate ../genorm/genorm -source=$GOFILE -destination=genorm -package=orm -module=github.com/mazrean/genorm-workspace/workspace/genorm

package main

import (
	"time"

	"github.com/google/uuid"
	"github.com/mazrean/genorm"
)

type User struct {
	ID       uuid.UUID `genorm:"id"`
	Name     string    `genorm:"name"`
	Password string    `genorm:"password"`
}

func (u *User) TableName() string {
	return "users"
}

type Message struct {
	ID        uuid.UUID `genorm:"id"`
	UserID    uuid.UUID `genorm:"user_id"`
	Content   string    `genorm:"content"`
	CreatedAt time.Time `genorm:"created_at"`
	User      genorm.Ref[User]
}

func (m *Message) TableName() string {
	return "messages"
}
