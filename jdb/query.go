package jdb

import (
	"strings"

	"github.com/cgalvisleon/elvis/console"
	"github.com/cgalvisleon/elvis/et"
	"github.com/cgalvisleon/elvis/event"
	"github.com/cgalvisleon/elvis/msg"
	"github.com/cgalvisleon/elvis/strs"
	"github.com/cgalvisleon/elvis/utility"
)

/**
* Data Definition Language
**/
func SQLQuote(sql string) string {
	sql = strings.TrimSpace(sql)

	result := strs.Replace(sql, `'`, `"`)
	result = strs.Trim(result)

	return result
}

func SQLDDL(sql string, args ...any) string {
	sql = strings.TrimSpace(sql)

	for i, arg := range args {
		old := strs.Format(`$%d`, i+1)
		new := strs.Format(`%v`, arg)
		sql = strings.ReplaceAll(sql, old, new)
	}

	return sql
}

func SQLParse(sql string, args ...any) string {
	for i := range args {
		old := strs.Format(`$%d`, i+1)
		new := strs.Format(`{$%d}`, i+1)
		sql = strings.ReplaceAll(sql, old, new)
	}

	for i, arg := range args {
		old := strs.Format(`{$%d}`, i+1)
		new := strs.Format(`%v`, et.Unquote(arg))
		sql = strings.ReplaceAll(sql, old, new)
	}

	return sql
}

/**
* DBQDDL
**/
func DBQDDL(db int, sql string, args ...any) (et.Items, error) {
	if conn == nil || len(conn.Db) == 0 || conn.Db[db].Db == nil {
		return et.Items{}, console.AlertF(msg.ERR_COMM)
	}

	sql = SQLParse(sql, args...)
	rows, err := conn.Db[db].Db.Query(sql)
	if err != nil {
		return et.Items{}, console.ErrorF(msg.ERR_SQL, err.Error(), sql)
	}
	defer rows.Close()

	items := rowsItems(rows)

	event.Action("sql/ddl", et.Json{
		"sql": sql,
	})

	return items, nil
}

func DBQuery(db int, sql string, args ...any) (et.Items, error) {
	if conn == nil || len(conn.Db) == 0 || conn.Db[db].Db == nil {
		return et.Items{}, console.AlertF(msg.ERR_COMM)
	}

	sql = SQLParse(sql, args...)
	rows, err := conn.Db[db].Db.Query(sql)
	if err != nil {
		return et.Items{}, console.ErrorF(msg.ERR_SQL, err.Error(), sql)
	}
	defer rows.Close()

	items := rowsItems(rows)

	event.Action("sql/query", et.Json{
		"sql": sql,
	})

	return items, nil
}

func DBQueryOne(db int, sql string, args ...any) (et.Item, error) {
	items, err := DBQuery(db, sql, args...)
	if err != nil {
		return et.Item{}, err
	}

	if items.Count == 0 {
		return et.Item{
			Ok:     false,
			Result: et.Json{},
		}, nil
	}

	return et.Item{
		Ok:     items.Ok,
		Result: items.Result[0],
	}, nil
}

func DBQueryCount(db int, sql string, args ...any) int {
	item, err := DBQueryOne(db, sql, args...)
	if err != nil {
		return -1
	}

	return item.Int("count")
}

/**
*
**/
func DBQueryAtrib(db int, sql, atrib string, args ...any) (et.Items, error) {
	if conn == nil || len(conn.Db) == 0 || conn.Db[db].Db == nil {
		return et.Items{}, console.AlertF(msg.ERR_COMM)
	}

	sql = SQLParse(sql, args...)
	rows, err := conn.Db[db].Db.Query(sql)
	if err != nil {
		return et.Items{}, console.ErrorF(msg.ERR_SQL, err.Error(), sql)
	}
	defer rows.Close()

	items := atribItems(rows, atrib)

	return items, nil
}

func DBQueryAtribOne(db int, sql, atrib string, args ...any) (et.Item, error) {
	items, err := DBQueryAtrib(db, sql, atrib, args...)
	if err != nil {
		return et.Item{}, err
	}

	if items.Count == 0 {
		return et.Item{
			Ok:     false,
			Result: et.Json{},
		}, nil
	}

	return et.Item{
		Ok:     items.Ok,
		Result: items.Result[0],
	}, nil
}

func DBQueryData(db int, sql string, args ...any) (et.Items, error) {
	return DBQueryAtrib(db, sql, "_data", args...)
}

func DBQueryDataOne(db int, sql string, args ...any) (et.Item, error) {
	return DBQueryAtribOne(db, sql, "_data", args...)
}

/**
* Query
**/
func QDDL(sql string, args ...any) (et.Items, error) {
	return DBQDDL(0, sql, args...)
}

func Query(sql string, args ...any) (et.Items, error) {
	return DBQuery(0, sql, args...)
}

func QueryOne(sql string, args ...any) (et.Item, error) {
	return DBQueryOne(0, sql, args...)
}

func QueryCount(sql string, args ...any) int {
	return DBQueryCount(0, sql, args...)
}

func QueryAtrib(sql, atrib string, args ...any) (et.Items, error) {
	return DBQueryAtrib(0, sql, atrib, args...)
}

func QueryAtribOne(sql, atrib string, args ...any) (et.Item, error) {
	return DBQueryAtribOne(0, sql, atrib, args...)
}

func QueryData(sql string, args ...any) (et.Items, error) {
	return DBQueryData(0, sql, args...)
}

func QueryDataOne(sql string, args ...any) (et.Item, error) {
	return DBQueryDataOne(0, sql, args...)
}

/**
*
**/
func HttpQuery(sql string, args []any) (et.Items, error) {
	if !utility.ValidStr(sql, 0, []string{""}) {
		return et.Items{}, console.AlertF("SQL is empty")
	}

	return Query(sql, args...)
}
