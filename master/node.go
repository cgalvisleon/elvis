package master

import (
	"time"

	"github.com/cgalvisleon/elvis/console"
	"github.com/cgalvisleon/elvis/et"
	"github.com/cgalvisleon/elvis/jdb"
	"github.com/cgalvisleon/elvis/msg"
	"github.com/cgalvisleon/elvis/utility"
)

const NodeStatusIdle = 0
const NodeStatusActive = 1
const NodeStatusWorking = 2
const NodeStatusSync = 3
const NodeStatusError = 4

type Node struct {
	Db          int
	ConnStr     string
	Date_make   time.Time `json:"date_make"`
	Date_update time.Time `json:"date_update"`
	Id          string    `json:"_id"`
	Mode        int       `json:"mode"`
	Data        et.Json   `json:"_data"`
	Status      int       `json:"status"`
	Index       int       `json:"index"`
}

func (n *Node) Scan(data *et.Json) error {
	n.Date_make = data.Time("date_make")
	n.Date_update = data.Time("date_update")
	n.Id = data.Str("_id")
	n.Mode = data.Int("mode")
	n.Data = data.Json("_data")
	n.Index = data.Int("index")
	n.Status = NodeStatusIdle

	return nil
}

func (c *Node) LatIndex() int {
	sql := `
	SELECT INDEX FROM core.MODE
	LIMIT 1;`

	item, err := jdb.DBQueryOne(c.Db, sql)
	if err != nil {
		return -1
	}

	return item.Index()
}

func (c *Node) GetSyncByIdT(idT string) (et.Item, error) {
	sql := `
  SELECT *
  FROM core.SYNC
  WHERE _IDT=$1
  LIMIT 1;`

	item, err := jdb.DBQueryOne(c.Db, sql, idT)
	if err != nil {
		return et.Item{}, err
	}

	return item, nil
}

func (c *Node) DelSyncByIndex(index int) error {
	sql := `
  DELETE FROM core.SYNC
  WHERE INDEX=$1;`

	_, err := jdb.DBQueryOne(c.Db, sql, index)
	if err != nil {
		return err
	}

	return nil
}

func NewNode(params *et.Json) (*Node, error) {
	result := &Node{}
	err := result.Scan(params)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func DefineNodes() error {
	if master.InitNodes {
		return nil
	}

	sql := `
  -- DROP TABLE IF EXISTS core.NODES CASCADE;
	-- DROP TABLE IF EXISTS core.SYNC CASCADE;
	CREATE SCHEMA IF NOT EXISTS "core";

  CREATE TABLE IF NOT EXISTS core.NODES(
		DATE_MAKE TIMESTAMP DEFAULT NOW(),
		DATE_UPDATE TIMESTAMP DEFAULT NOW(),
    _ID VARCHAR(80) DEFAULT '',
    MODE INTEGER DEFAULT 0,
		PASSWORD VARCHAR(250) DEFAULT '',
    _DATA JSONB DEFAULT '{}',
		INDEX SERIAL,
		PRIMARY KEY(_ID)
	);

	CREATE OR REPLACE FUNCTION core.NODES_UPSET()
	RETURNS
		TRIGGER AS $$
	BEGIN
		PERFORM pg_notify(
			'node',
			json_build_object(
				'action', TG_OP,
				'node', NEW._ID
			)::text
		);

	RETURN NEW;
	END;
	$$ LANGUAGE plpgsql;

	CREATE OR REPLACE FUNCTION core.NODES_DELETE()
	RETURNS
		TRIGGER AS $$
	BEGIN
		PERFORM pg_notify(
			'node',
			json_build_object(
				'action', TG_OP,
				'node', OLD._ID
			)::text
		);

	RETURN NEW;
	END;
	$$ LANGUAGE plpgsql;

	DROP TRIGGER IF EXISTS NODES_UPSET ON core.NODES CASCADE;
	CREATE TRIGGER NODES_UPSET
	AFTER INSERT OR UPDATE ON core.NODES
	FOR EACH ROW
	EXECUTE PROCEDURE core.NODES_UPSET();

	DROP TRIGGER IF EXISTS NODES_DELETE ON core.NODES CASCADE;
	CREATE TRIGGER NODES_DELETE
	AFTER DELETE ON core.NODES
	FOR EACH ROW
	EXECUTE PROCEDURE core.NODES_DELETE();`

	_, err := jdb.QDDL(sql)
	if err != nil {
		return console.Panic(err)
	}

	master.InitNodes = true

	go master.LoadNodes()

	return nil
}

/**
* Mode
*	Handler for CRUD data
 */
func GetNodeById(id string) (et.Item, error) {
	sql := `
	SELECT
	A._DATA||
  jsonb_build_object(
    'mode', A.MODE,
		'index', A.INDEX
  ) AS _DATA
	FROM core.NODES A
	WHERE A._ID=$1
	LIMIT 1;`

	item, err := jdb.QueryDataOne(sql, id)
	if err != nil {
		return et.Item{}, err
	}

	delete(item.Result, "password")

	return item, nil
}

func UpSertNode(id string, mode int, driver, host string, port int, dbname, user, password string) (et.Item, error) {
	if !utility.ValidId(id) {
		return et.Item{}, console.AlertF(msg.MSG_ATRIB_REQUIRED, "id")
	}

	current, err := GetNodeById(id)
	if err != nil {
		return et.Item{}, err
	}

	now := utility.Now()
	data := et.Json{
		"driver": driver,
		"host":   host,
		"port":   port,
		"dbname": dbname,
		"user":   user,
	}

	if current.Ok {
		sql := `
		UPDATE core.NODES SET
		DATE_UPDATE=$2,
		MODE=$3,
		PASSWORD=$4,
		_DATA=$5
		WHERE _ID=$1
		RETURNING INDEX;`

		item, err := jdb.QueryOne(sql, id, now, mode, password, data)
		if err != nil {
			return et.Item{}, err
		}

		return et.Item{
			Ok: item.Ok,
			Result: et.Json{
				"message": msg.RECORD_UPDATE,
				"_id":     id,
				"index":   item.Index(),
			},
		}, nil
	}

	sql := `
		INSERT INTO core.NODES(DATE_MAKE, DATE_UPDATE, _ID, MODE, _DATA)
		VALUES($1, $1, $2, $3, $4)
		RETURNING INDEX;`

	item, err := jdb.QueryOne(sql, now, id, mode, data)
	if err != nil {
		return et.Item{}, err
	}

	return et.Item{
		Ok: item.Ok,
		Result: et.Json{
			"message": msg.RECORD_CREATE,
			"_id":     id,
			"index":   item.Index(),
		},
	}, nil
}

func DeleteNodeById(id string) (et.Item, error) {
	sql := `
	DELETE FROM core.NODES	
	WHERE _ID=$1
	RETURNING *;`

	item, err := jdb.QueryDataOne(sql, id)
	if err != nil {
		return et.Item{}, err
	}

	delete(item.Result, "password")

	return item, nil
}

func AllNodes(search string, page, rows int) (et.List, error) {
	sql := `
	SELECT COUNT(*) AS COUNT
	FROM core.NODES A
	WHERE CONCAT('MODE:', A.MODE, ':DATA:', A._DATA::TEXT) ILIKE CONCAT('%', $1, '%');`

	all := jdb.QueryCount(sql, search)

	sql = `
	SELECT A._DATA||
	jsonb_build_object(
		'data_make', A.DATE_MAKE,
		'date_update', A.DATE_UPDATE,
		'_id', A._ID,
		'mode', A.MODE,
		'index', A.INDEX
	) AS _DATA
	FROM core.NODES A
	WHERE CONCAT('MODE:', A.MODE, ':DATA:', A._DATA::TEXT) ILIKE CONCAT('%', $1, '%')
	LIMIT $2 OFFSET $3;`

	offset := (page - 1) * rows
	items, err := jdb.Query(sql, search, rows, offset)
	if err != nil {
		return et.List{}, err
	}

	return items.ToList(all, page, rows), nil
}
