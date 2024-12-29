package master

import (
	"github.com/cgalvisleon/elvis/et"
	"github.com/cgalvisleon/elvis/jdb"
	"github.com/cgalvisleon/elvis/strs"
	"github.com/cgalvisleon/elvis/utility"
)

func (c *Node) SyncNode() error {
	err := c.SyncMasterToNode()
	if err != nil {
		c.Status = NodeStatusError
		return err
	}

	err = c.SyncNodeToMaster()
	if err != nil {
		c.Status = NodeStatusError
		return err
	}

	go jdb.Listen(c.ConnStr, "sync", "sync", listenSync)

	c.Status = NodeStatusActive

	return nil
}

func (c *Node) SyncMasterToNode() error {
	var lastIndex = c.LatIndex()
	var ok bool = true
	var rows int = 100
	var page int = 1

	for ok {
		c.Status = NodeStatusSync
		ok = false

		offset := (page - 1) * rows
		sql := strs.Format(`
		SELECT A.*
		FROM core.SYNC A
		WHERE A.INDEX>$1
		ORDER BY A.TABLE_SCHEMA, A.TABLE_SCHEMA, A.ACTION, A.INDEX
		LIMIT %d OFFSET %d;`, rows, offset)

		items, err := jdb.Query(sql, lastIndex)
		if err != nil {
			c.Status = NodeStatusError
			return err
		}

		oldSchema := ""
		oldTable := ""
		oldAction := ""
		batch := ""
		for _, item := range items.Result {
			schema := item.Str("table_schema")
			table := item.Str("table_name")
			idT := item.Key("_idt")
			_data := item.Json("_data")
			action := item.Str("action")
			lastIndex = item.Index()
			if !utility.Contains([]string{"INSERT", "UPDATE", "DELETE", "DDL"}, action) {
				continue
			}

			if oldSchema != schema || oldTable != table || oldAction != action {
				err = c.SyncQuery(batch, lastIndex)
				if err != nil {
					c.Status = NodeStatusError
					return err
				}
				oldSchema = schema
				oldTable = table
				oldAction = action
				batch = ""
			}
			if oldSchema == schema && oldTable == table && action == "INSERT" && len(batch) == 0 {
				batch = c.SqlField(schema, table, _data)
			} else if oldSchema == schema && oldTable == table && action == "INSERT" {
				batch = strs.Format(`%s,`, batch)
			}

			query, append := c.ToSql(schema, table, idT, _data, action)
			if append {
				batch = strs.Append(batch, query, "\n")
			}

			ok = true
		}

		err = c.SyncQuery(batch, lastIndex)
		if err != nil {
			c.Status = NodeStatusError
			return err
		}

		page++
	}

	return nil
}

func (c *Node) SyncNodeToMaster() error {
	var ok bool = true
	var rows int = 100

	for ok {
		c.Status = NodeStatusSync
		ok = false

		sql := strs.Format(`
		SELECT A.*
		FROM core.SYNC A
		ORDER BY A.INDEX
		LIMIT %d;`, rows)

		items, err := jdb.DBQuery(c.Db, sql)
		if err != nil {
			c.Status = NodeStatusError
			return err
		}

		for _, item := range items.Result {
			schema := item.Str("table_schema")
			table := item.Str("table_name")
			idT := item.Key("_idt")
			_data := item.Json("_data")
			action := item.Str("action")
			_index := item.Int("index")

			err := c.SyncRecord(schema, table, idT, _data, action, _index)
			if err != nil {
				return err
			}

			ok = true
		}
	}

	return nil
}

func (c *Node) SyncQuery(query string, index int) error {
	sql := strs.Format(`
	UPDATE core.MODE SET
	INDEX=%d
	WHERE INDEX<%d
	RETURNING INDEX;`, index, index)

	query = strs.Append(query, sql, "\n")

	_, err := jdb.DBQuery(c.Db, query)
	if err != nil {
		return err
	}

	return nil
}

func (c *Node) SyncRecord(schema, table, idT string, _data et.Json, action string, _index int) error {
	query := ""
	if action == "INSERT" {
		fields := c.SqlField(schema, table, _data)
		insert, append := c.ToSql(schema, table, idT, _data, action)
		if append {
			query = strs.Append(fields, insert, "\n")
		}
	} else {
		query, _ = c.ToSql(schema, table, idT, _data, action)
	}

	index, err := master.SetSync(schema, table, action, c.Id, idT, _data, query)
	if err != nil {
		c.Status = NodeStatusError
		return err
	}

	for _, n := range master.Nodes {
		n.Status = NodeStatusSync
		if n.Id != c.Id {
			err = c.SyncQuery(query, index)
			if err != nil {
				c.Status = NodeStatusError
				return err
			}

			c.Status = NodeStatusActive
		}
	}

	return c.DelSyncByIndex(_index)
}

func (c *Node) SyncIdT(idT string) error {
	item, err := c.GetSyncByIdT(idT)
	if err != nil {
		c.Status = NodeStatusError
		return err
	}

	if !item.Ok {
		return nil
	}

	schema := item.Str("table_schema")
	table := item.Str("table_name")
	_data := item.Json("_data")
	action := item.Str("action")
	_index := item.Int("index")
	err = c.SyncRecord(schema, table, idT, _data, action, _index)
	if err != nil {
		return err
	}

	return nil
}
