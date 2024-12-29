package core

import (
	"github.com/cgalvisleon/elvis/console"
	"github.com/cgalvisleon/elvis/et"
	"github.com/cgalvisleon/elvis/jdb"
	"github.com/cgalvisleon/elvis/linq"
	"github.com/cgalvisleon/elvis/strs"
)

var (
	_db         int = 0
	makedSeries bool
	series      map[string]bool = make(map[string]bool)
)

func SetmasterDB(idx int) {
	_db = idx
}

func defineSeries() error {
	if makedSeries {
		return nil
	}

	if err := defineSchemaCore(); err != nil {
		return console.Panic(err)
	}

	sql := `  
  -- DROP TABLE IF EXISTS core.SERIES CASCADE;

  CREATE TABLE IF NOT EXISTS core.SERIES(
		DATE_MAKE TIMESTAMP DEFAULT NOW(),
		DATE_UPDATE TIMESTAMP DEFAULT NOW(),
		SERIE VARCHAR(250) DEFAULT '',
		VALUE BIGINT DEFAULT 0,
		PRIMARY KEY(SERIE)
	);
	
	CREATE OR REPLACE FUNCTION core.nextserie(tag VARCHAR(250))
	RETURNS BIGINT AS $$
	DECLARE
	 result BIGINT;
	BEGIN
	 INSERT INTO core.SERIES AS A (SERIE, VALUE)
	 SELECT tag, 1
	 ON CONFLICT (SERIE) DO UPDATE SET
	 VALUE = A.VALUE + 1
	 RETURNING VALUE INTO result;

	 RETURN COALESCE(result, 0);
	END;
	$$ LANGUAGE plpgsql;
	
	CREATE OR REPLACE FUNCTION core.setserie(tag VARCHAR(250), val BIGINT)
	RETURNS BIGINT AS $$
	DECLARE
	 result BIGINT;
	BEGIN
	 INSERT INTO core.SERIES AS A (SERIE, VALUE)
	 SELECT tag, val
	 ON CONFLICT (SERIE) DO UPDATE SET
	 VALUE = val
	 WHERE A.VALUE < val
	 RETURNING VALUE INTO result;

	 RETURN COALESCE(result, 0);
	END;
	$$ LANGUAGE plpgsql;
	
	CREATE OR REPLACE FUNCTION core.currserie(tag VARCHAR(250))
	RETURNS BIGINT AS $$
	DECLARE
	 result BIGINT;
	BEGIN
	 SELECT VALUE INTO result
	 FROM core.SERIES
	 WHERE SERIE = tag LIMIT 1;

	 RETURN COALESCE(result, 0);
	END;
	$$ LANGUAGE plpgsql;`

	_, err := jdb.QDDL(sql)
	if err != nil {
		return console.Panic(err)
	}

	makedSeries = true

	return nil
}

/**
*
**/
func NextSerie(tag string) int {
	tag = strs.Replace(tag, ".", "_")
	tag = strs.Replace(tag, " ", "_")
	if makedSeries {
		sql := `SELECT core.nextserie($1) AS SERIE;`

		item, err := jdb.DBQueryOne(_db, sql, tag)
		if err != nil {
			console.Error(err)
			return 0
		}

		result := item.Int("serie")

		return result
	}

	if _, ok := series[tag]; !ok {
		exist, err := jdb.ExistSerie(_db, "public", tag)
		if err != nil {
			console.Error(err)
			return 0
		}

		if !exist {
			_, err := jdb.CreateSerie(_db, "public", tag)
			if err != nil {
				console.Error(err)
				return 0
			}
		}

		series[tag] = true
	}

	sql := `SELECT nextval($1) AS SERIE;`

	item, err := jdb.DBQueryOne(_db, sql, tag)
	if err != nil {
		console.Error(err)
		return 0
	}

	result := item.Int("serie")

	return result
}

/**
* Serie
**/
func DefineSerie(model *linq.Model) error {
	if err := defineSeries(); err != nil {
		return err
	}

	tableName := model.Name
	fieldName := model.SerieField

	sql := strs.Format(`
	SELECT MAX(%s) AS INDEX
	FROM %s;`, strs.Uppcase(fieldName), tableName)

	item, err := jdb.DBQueryOne(_db, sql)
	if err != nil {
		return err
	}

	max := item.Int("index")
	tag := strs.Replace(tableName, ".", "_")
	tag = strs.Replace(tag, " ", "_")

	_, err = SetSerie(tag, max)
	if err != nil {
		return err
	}

	model.Trigger(linq.BeforeInsert, func(model *linq.Model, old, new *et.Json, data et.Json) error {
		if model.UseSerie {
			index := NextSerie(model.Name)
			new.Set(model.SerieField, index)
		}

		return nil
	})

	return nil
}

func NextCode(tag, prefix string) string {
	num := NextSerie(tag)

	if len(prefix) == 0 {
		return strs.Format("%08v", num)
	} else {
		return strs.Format("%s%08v", prefix, num)
	}
}

func SetSerie(tag string, val int) (int, error) {
	if makedSeries {
		sql := `SELECT core.setserie($1, $2);`

		_, err := jdb.DBQueryOne(_db, sql, tag, val)
		if err != nil {
			return 0, err
		}

		return val, nil
	}

	if _, ok := series[tag]; !ok {
		exist, err := jdb.ExistSerie(_db, "public", tag)
		if err != nil {
			return 0, err
		}

		if !exist {
			_, err := jdb.CreateSerie(_db, "public", tag)
			if err != nil {
				return 0, err
			}
		}

		series[tag] = true
	}

	sql := `SELECT setval($1, $2);`

	_, err := jdb.DBQueryOne(_db, sql, tag, val)
	if err != nil {
		return 0, err
	}

	return val, nil
}

func LastSerie(tag string) int {
	if makedSeries {
		sql := `SELECT core.currserie($1) AS SERIE;`

		item, err := jdb.DBQueryOne(_db, sql, tag)
		if err != nil {
			return 0
		}

		result := item.Int("serie")

		return result
	}

	if _, ok := series[tag]; !ok {
		exist, err := jdb.ExistSerie(_db, "public", tag)
		if err != nil {
			return 0
		}

		if !exist {
			_, err := jdb.CreateSerie(_db, "public", tag)
			if err != nil {
				return 0
			}
		}

		series[tag] = true
	}

	sql := `SELECT currval($1) AS SERIE;`

	item, err := jdb.DBQueryOne(_db, sql, tag)
	if err != nil {
		return 0
	}

	result := item.Int("serie")

	return result
}

func SyncSerie(c chan int) error {
	var ok bool = true
	var rows int = 30
	var page int = 1
	for ok {
		ok = false

		offset := (page - 1) * rows
		sql := strs.Format(`
		SELECT A.*
		FROM core.SERIES A
		ORDER BY A.SERIE
		LIMIT %d OFFSET %d;`, rows, offset)

		items, err := jdb.Query(sql)
		if err != nil {
			return err
		}

		for _, item := range items.Result {
			tag := item.Str("serie")
			val := item.Int("value")
			_, err = SetSerie(tag, val)
			if err != nil {
				return console.Error(err)
			}

			ok = true
		}

		page++
	}

	c <- _db

	return nil
}
