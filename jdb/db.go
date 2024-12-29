package jdb

import (
	"database/sql"

	"github.com/cgalvisleon/elvis/et"
	"github.com/cgalvisleon/elvis/strs"
)

const Postgres = "postgres"
const Mysql = "mysql"
const Sqlserver = "sqlserver"
const Firebird = "firebird"

type Db struct {
	Index       int
	Description string
	Driver      string
	Host        string
	Port        int
	Dbname      string
	User        string
	ConnStr     string
	Db          *sql.DB
}

func (c *Db) Close() error {
	err := c.Db.Close()
	if err != nil {
		return err
	}

	return nil
}

func (c *Db) Describe() et.Json {
	host := strs.Format(`%s:%d`, c.Host, c.Port)
	return et.Json{
		"name":        c.Dbname,
		"description": c.Description,
		"driver":      c.Driver,
		"host":        host,
	}
}
