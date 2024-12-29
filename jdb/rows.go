package jdb

import (
	"database/sql"

	"github.com/cgalvisleon/elvis/et"
)

/**
* Data Definition Language
**/

func rowsItems(rows *sql.Rows) et.Items {
	var result et.Items = et.Items{Result: []et.Json{}}

	for rows.Next() {
		var item et.Item
		item.Scan(rows)
		result.Result = append(result.Result, item.Result)
		result.Ok = true
		result.Count++
	}

	return result
}

func atribItems(rows *sql.Rows, atrib string) et.Items {
	var result et.Items = et.Items{Result: []et.Json{}}

	for rows.Next() {
		var item et.Item
		item.Scan(rows)
		result.Result = append(result.Result, item.Result.Json(atrib))
		result.Ok = true
		result.Count++
	}

	return result
}
