package core

import (
	"github.com/cgalvisleon/elvis/console"
	"github.com/cgalvisleon/elvis/jdb"
	"github.com/cgalvisleon/elvis/linq"
	"github.com/cgalvisleon/elvis/strs"
)

var makedRecycling bool

func defineRecycling() error {
	if makedRecycling {
		return nil
	}

	if err := defineSchemaCore(); err != nil {
		return console.Panic(err)
	}

	sql := `  
  -- DROP TABLE IF EXISTS core.RECYCLING CASCADE;

  CREATE TABLE IF NOT EXISTS core.RECYCLING(
		TABLE_SCHEMA VARCHAR(80) DEFAULT '',
    TABLE_NAME VARCHAR(80) DEFAULT '',
    _IDT VARCHAR(80) DEFAULT '-1',
    INDEX SERIAL,
		PRIMARY KEY(TABLE_SCHEMA, TABLE_NAME, _IDT)
	);
  CREATE INDEX IF NOT EXISTS RECYCLING_INDEX_IDX ON core.RECYCLING(INDEX);
	CREATE INDEX IF NOT EXISTS RECYCLING_TABLE_SCHEMA_IDX ON core.RECYCLING(TABLE_SCHEMA);
	CREATE INDEX IF NOT EXISTS RECYCLING_TABLE_NAME_IDX ON core.RECYCLING(TABLE_NAME);
	CREATE INDEX IF NOT EXISTS RECYCLING__IDT_IDX ON core.RECYCLING(INDEX);

	CREATE OR REPLACE FUNCTION core.RECYCLING()
  RETURNS
    TRIGGER AS $$
  BEGIN
		IF NEW._STATE != OLD._STATE AND NEW._STATE = '-2' THEN
    	INSERT INTO core.RECYCLING(TABLE_SCHEMA, TABLE_NAME, _IDT)
    	VALUES (TG_TABLE_SCHEMA, TG_TABLE_NAME, NEW._IDT);

      PERFORM pg_notify(
      'recycling',
      json_build_object(
        '_idt', NEW._IDT
      )::text
      );
		ELSEIF NEW._STATE != OLD._STATE AND OLD._STATE = '-2' THEN
			DELETE FROM core.RECYCLING WHERE _IDT=NEW._IDT;
    END IF;

  RETURN NEW;
  END;
  $$ LANGUAGE plpgsql;

	CREATE OR REPLACE FUNCTION core.ERASE()
  RETURNS
    TRIGGER AS $$
  BEGIN
		DELETE FROM core.RECYCLING WHERE _IDT=OLD._IDT;
  	RETURN OLD;
  END;
  $$ LANGUAGE plpgsql;`

	_, err := jdb.QDDL(sql)
	if err != nil {
		return console.Panic(err)
	}

	makedRecycling = true

	return nil
}

func SetRecycligTrigger(model *linq.Model) error {
	if err := defineRecycling(); err != nil {
		return err
	}

	schema := model.Schema
	table := model.Table
	_, err := jdb.CreateColumn(0, schema, table, "_IDT", "VARCHAR(80)", "-1")
	if err != nil {
		return err
	}

	_, err = jdb.CreateIndex(0, schema, table, "_IDT")
	if err != nil {
		return err
	}

	created, err := jdb.CreateColumn(0, schema, table, "_STATE", "VARCHAR(80)", "0")
	if err != nil {
		return err
	}

	if created {
		_, err = jdb.CreateIndex(0, schema, table, "_STATE")
		if err != nil {
			return err
		}

		tableName := strs.Append(strs.Lowcase(schema), strs.Uppcase(table), ".")
		sql := jdb.SQLDDL(`
    CREATE INDEX IF NOT EXISTS $2_IDT_IDX ON $1(_STATE);

    DROP TRIGGER IF EXISTS RECYCLING ON $1 CASCADE;
    CREATE TRIGGER RECYCLING
    AFTER UPDATE ON $1
    FOR EACH ROW
    EXECUTE PROCEDURE core.RECYCLING();

    DROP TRIGGER IF EXISTS ERASE ON $1 CASCADE;
    CREATE TRIGGER ERASE
    AFTER DELETE ON $1
    FOR EACH ROW
    EXECUTE PROCEDURE core.ERASE();`, tableName, strs.Uppcase(table))

		_, err := jdb.QDDL(sql)
		if err != nil {
			return err
		}
	}

	return nil
}
