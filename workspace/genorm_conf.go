//go:build genorm
// +build genorm

//go:generate go run github.com/mazrean/genorm/cmd/genorm@latest -source=$GOFILE -destination=genorm -package=orm -module=github.com/mazrean/genorm-workspace/workspace/genorm

package main

import (
	"time"

	"github.com/google/uuid"
	"github.com/mazrean/genorm"
	"github.com/mazrean/genorm-workspace/workspace/types"
)

type User struct {
	ID        types.UserID `genorm:"id"`
	Name      string       `genorm:"name"`
	CreatedAt time.Time    `genorm:"created_at"`
	Message   genorm.Ref[Message]
}

func (u *User) TableName() string {
	return "users"
}

type Message struct {
	ID        types.MessageID `genorm:"id"`
	UserID    types.UserID    `genorm:"user_id"`
	Content   string          `genorm:"content"`
	CreatedAt time.Time       `genorm:"created_at"`
	Option1   genorm.Ref[MessageOption1]
	Option2   genorm.Ref[MessageOption2]
	Option3   genorm.Ref[MessageOption3]
	Option4   genorm.Ref[MessageOption4]
}

func (m *Message) TableName() string {
	return "messages"
}

type MessageOption1 struct {
	ID        uuid.UUID       `genorm:"id"`
	MessageID types.MessageID `genorm:"message_id"`
}

func (m *MessageOption1) TableName() string {
	return "message_option1s"
}

type MessageOption2 struct {
	ID        uuid.UUID       `genorm:"id"`
	MessageID types.MessageID `genorm:"message_id"`
}

func (m *MessageOption2) TableName() string {
	return "message_option2s"
}

type MessageOption3 struct {
	ID        uuid.UUID       `genorm:"id"`
	MessageID types.MessageID `genorm:"message_id"`
}

func (m *MessageOption3) TableName() string {
	return "message_option3s"
}

type MessageOption4 struct {
	ID        uuid.UUID       `genorm:"id"`
	MessageID types.MessageID `genorm:"message_id"`
}

func (m *MessageOption4) TableName() string {
	return "message_option4s"
}
