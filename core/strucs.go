package core

import (
	"github.com/cgalvisleon/elvis/console"
	"github.com/cgalvisleon/elvis/jdb"
	"github.com/cgalvisleon/elvis/strs"
)

var makedStructs bool

func defineStructs() error {
	if makedStructs {
		return nil
	}

	if err := defineSchemaCore(); err != nil {
		return console.Panic(err)
	}

	sql := `  
  -- DROP TABLE IF EXISTS core.STRUCTS CASCADE;

  CREATE TABLE IF NOT EXISTS core.STRUCTS(
		KIND VARCHAR(80) DEFAULT '',
		SCHEMA VARCHAR(80) DEFAULT '',
    NAME VARCHAR(80) DEFAULT '',
    SQL TEXT DEFAULT '',
    INDEX SERIAL,
		PRIMARY KEY(KIND, SCHEMA, NAME)
	);
  CREATE INDEX IF NOT EXISTS STRUCTS_INDEX_IDX ON core.STRUCTS(INDEX);
	CREATE INDEX IF NOT EXISTS STRUCTS_KIND_IDX ON core.STRUCTS(KIND);
	CREATE INDEX IF NOT EXISTS STRUCTS_SCHEMA_IDX ON core.STRUCTS(SCHEMA);
	CREATE INDEX IF NOT EXISTS STRUCTS_NAME_IDX ON core.STRUCTS(NAME);

	CREATE OR REPLACE FUNCTION core.setstruct(
	VKIND VARCHAR(80),
	VSCHEMA VARCHAR(80),
	VNAME VARCHAR(80),
	VSQL TEXT)
	RETURNS INT AS $$
	DECLARE
	 result INT;
	BEGIN
	 INSERT INTO core.STRUCTS AS A (KIND, SCHEMA, NAME, SQL)
	 SELECT VKIND, VSCHEMA, VNAME, VSQL
	 ON CONFLICT (KIND, SCHEMA, NAME) DO UPDATE SET
	 SQL = VSQL
	 RETURNING INDEX INTO result;

	 RETURN COALESCE(result, 0);
	END;
	$$ LANGUAGE plpgsql;`

	_, err := jdb.QDDL(sql)
	if err != nil {
		return console.Panic(err)
	}

	makedStructs = true

	return nil
}

func SetStruct(kind, schema, name, query string) error {
	if err := defineStructs(); err != nil {
		return err
	}

	query = strs.Replace(query, `'`, `"`)
	query = "\n" + query + "\n"
	sql := `SELECT core.setstruct($1, $2, $3, $4) AS INDEX;`
	_, err := jdb.Query(sql, kind, schema, name, query)
	if err != nil {
		return err
	}

	return nil
}

func DeleteStruct(kind, schema, name string) error {
	if err := defineStructs(); err != nil {
		return err
	}

	sql := `DELETE
	FROM core.STRUCTS
	WHERE KIND = $1
	AND SCHEMA = $2
	AND NAME = $3`
	_, err := jdb.Query(sql, kind, schema, name)
	if err != nil {
		return err
	}

	return nil
}
