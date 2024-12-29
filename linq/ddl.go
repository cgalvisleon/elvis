package linq

import (
	"github.com/cgalvisleon/elvis/et"
	"github.com/cgalvisleon/elvis/generic"
	"github.com/cgalvisleon/elvis/jdb"
	"github.com/cgalvisleon/elvis/strs"
)

/**
* The DDLCoulmn function generates a Data Definition Language (DDL) statement for a given column in a database.
* It takes into account various column properties such as foreign key constraints, default values, and the column's type and name.
* The generated DDL statement can be used to create or alter the column in a database.
**/

func DDLColumn(col *Column) string {
	var result string

	switch col.Driver() {
	default:
		_default := generic.New(col.Default)

		if _default.Str() == "NOW()" {
			result = strs.Append(`DEFAULT NOW()`, result, " ")
		} else {
			result = strs.Append(strs.Format(`DEFAULT %v`, et.Unquote(col.Default)), result, " ")
		}

		if col.Type == "SERIAL" {
			result = strs.Uppcase(col.Type)
		} else if len(col.Type) > 0 {
			result = strs.Append(strs.Uppcase(col.Type), result, " ")
		}
		if len(col.name) > 0 {
			result = strs.Append(strs.Uppcase(col.name), result, " ")
		}
	}

	return result
}

func DDLIndex(col *Column) string {
	var result string

	switch col.Driver() {
	default:
		result = jdb.SQLDDL(`CREATE INDEX IF NOT EXISTS $2_$3_IDX ON $1($3);`, strs.Lowcase(col.Model.Name), strs.Uppcase(col.Model.Table), strs.Uppcase(col.name))
	}

	return result
}

func DDLUniqueIndex(col *Column) string {
	var result string

	switch col.Driver() {
	default:
		result = jdb.SQLDDL(`CREATE UNIQUE INDEX IF NOT EXISTS $2_$3_IDX ON $1($3);`, strs.Lowcase(col.Model.Name), strs.Uppcase(col.Model.Table), strs.Uppcase(col.name))
	}

	return result
}
