package types

import (
	"database/sql/driver"

	"github.com/google/uuid"
)

type MessageID uuid.UUID

func (mid *MessageID) Scan(src any) error {
	return (*uuid.UUID)(mid).Scan(src)
}

func (mid MessageID) Value() (driver.Value, error) {
	return uuid.UUID(mid).Value()
}

type UserID uuid.UUID

func (uid *UserID) Scan(src any) error {
	return (*uuid.UUID)(uid).Scan(src)
}

func (uid UserID) Value() (driver.Value, error) {
	return uuid.UUID(uid).Value()
}
