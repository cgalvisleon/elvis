package core

import (
	"github.com/cgalvisleon/elvis/console"
	"github.com/cgalvisleon/elvis/jdb"
	"github.com/cgalvisleon/elvis/linq"
	"github.com/cgalvisleon/elvis/strs"
)

var makedListener bool

func DefineListener() error {
	if makedListener {
		return nil
	}

	if err := defineSchemaCore(); err != nil {
		return console.Panic(err)
	}

	sql := `
  CREATE OR REPLACE FUNCTION core.SYNC_INSERT()
  RETURNS
    TRIGGER AS $$
  DECLARE
   CHANNEL VARCHAR(250);
  BEGIN
    IF NEW._IDT = '-1' THEN
      NEW._IDT = uuid_generate_v4();
		END IF;

		CHANNEL = TG_TABLE_SCHEMA || '.' || TG_TABLE_NAME;
		PERFORM pg_notify(
		CHANNEL,
		json_build_object(
			'option', TG_OP,
			'_idt', NEW._IDT
		)::text
		);
    
  RETURN NEW;
  END;
  $$ LANGUAGE plpgsql;

  CREATE OR REPLACE FUNCTION core.SYNC_UPDATE()
  RETURNS
    TRIGGER AS $$
  DECLARE
    CHANNEL VARCHAR(250);
  BEGIN
    IF NEW._IDT = '-1' THEN
			NEW._IDT = uuid_generate_v4();    
    END IF;
    
		CHANNEL = TG_TABLE_SCHEMA || '.' || TG_TABLE_NAME;
		PERFORM pg_notify(
		CHANNEL,
		json_build_object(
			'option', TG_OP,
			'_idt', NEW._IDT  
		)::text
		);

  RETURN NEW;
  END;
  $$ LANGUAGE plpgsql;

  CREATE OR REPLACE FUNCTION core.SYNC_DELETE()
  RETURNS
    TRIGGER AS $$
  DECLARE
    CHANNEL VARCHAR(250);
  BEGIN
		CHANNEL = TG_TABLE_SCHEMA || '.' || TG_TABLE_NAME;
		PERFORM pg_notify(
		CHANNEL,
		json_build_object(
			'option', TG_OP,
			'_idt', OLD._IDT
		)::text
		);

  RETURN OLD;
  END;
  $$ LANGUAGE plpgsql;`

	_, err := jdb.QDDL(sql)
	if err != nil {
		return console.Panic(err)
	}

	makedListener = true

	return nil
}

func SetListenTrigger(model *linq.Model) error {
	if !makedListener {
		return nil
	}

	schema := model.Schema
	table := model.Table
	created, err := jdb.CreateColumn(0, schema, table, "_IDT", "VARCHAR(80)", "-1")
	if err != nil {
		return err
	}

	if created {
		_, err = jdb.CreateIndex(0, schema, table, "_IDT")
		if err != nil {
			return err
		}

		tableName := strs.Append(strs.Lowcase(schema), strs.Uppcase(table), ".")
		sql := jdb.SQLDDL(`    
    DROP TRIGGER IF EXISTS SYNC_INSERT ON $1 CASCADE;
    CREATE TRIGGER SYNC_INSERT
    BEFORE INSERT ON $1
    FOR EACH ROW
    EXECUTE PROCEDURE core.SYNC_INSERT();

    DROP TRIGGER IF EXISTS SYNC_UPDATE ON $1 CASCADE;
    CREATE TRIGGER SYNC_UPDATE
    BEFORE UPDATE ON $1
    FOR EACH ROW
    EXECUTE PROCEDURE core.SYNC_UPDATE();

    DROP TRIGGER IF EXISTS SYNC_DELETE ON $1 CASCADE;
    CREATE TRIGGER SYNC_DELETE
    BEFORE DELETE ON $1
    FOR EACH ROW
    EXECUTE PROCEDURE core.SYNC_DELETE();`, tableName, strs.Uppcase(table))

		_, err := jdb.QDDL(sql)
		if err != nil {
			return err
		}
	}

	channel := strs.Append(strs.Lowcase(schema), strs.Uppcase(table), ".")
	connStr := jdb.DB(model.Db).ConnStr
	go jdb.Listen(connStr, channel, "listen", model.OnListener)

	return nil
}
