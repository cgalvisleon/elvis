package driver

import (
	"database/sql"

	"github.com/cgalvisleon/elvis/et"
	"github.com/cgalvisleon/elvis/linq"
)

// Postgres driver
type Postgres struct {
	db *sql.DB
}

// Connect to db
func (p *Postgres) Connect(params et.Json) (*sql.DB, error) {
	return nil, nil
}

// Disconnect from db
func (p *Postgres) Disconnect() error {
	return p.db.Close()
}

// DDLModel model definition
func (p *Postgres) DDLModel(model *linq.Model) (string, error) {
	return "", nil
}

// Exec sql
func (p *Postgres) Exec(sql string, args ...any) error {
	return nil
}

// Query sql
func (p *Postgres) Query(sql string, args ...any) (et.Items, error) {

	return et.Items{}, nil
}

// QueryOne sql
func (p *Postgres) QueryOne(sql string, args ...any) (et.Item, error) {
	return et.Item{}, nil
}
