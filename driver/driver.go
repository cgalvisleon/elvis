package driver

import (
	"database/sql"

	"github.com/cgalvisleon/elvis/et"
	"github.com/cgalvisleon/elvis/linq"
)

// Driver interface
// Implement this interface to create a new driver for db
type Driver interface {
	Connect(params et.Json) (*sql.DB, error)
	Disconnect() error
	// Ddl model definition
	DDLModel(model *linq.Model) (string, error)
	// Sql definition
	// Execute sql
	Exec(sql string, args ...any) error
	Query(sql string, args ...any) (et.Items, error)
	QueryOne(sql string, args ...any) (et.Item, error)
}
