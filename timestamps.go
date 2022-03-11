package orm

import (
	"database/sql"
)

type Timestamps struct {
	CreatedAt sql.NullTime
	UpdatedAt sql.NullTime
	DeletedAt sql.NullTime
}
