package master

import (
	"github.com/cgalvisleon/elvis/console"
	"github.com/cgalvisleon/elvis/core"
	"github.com/cgalvisleon/elvis/et"
	"github.com/cgalvisleon/elvis/jdb"
	"github.com/cgalvisleon/elvis/strs"
)

var (
	master *Master
)

func InitDefine() error {
	if master != nil {
		return nil
	}

	master = &Master{}

	if err := DefineNodes(); err != nil {
		return console.Panic(err)
	}

	go jdb.Listen(jdb.DB(0).ConnStr, "node", "node", listenNode)

	console.LogK("CORE", "Init Master")

	return nil
}

/**
*
**/
type Master struct {
	InitNodes  bool
	InitSecret bool
	Nodes      []Node
}

func (c *Master) GetNodeByID(id string) *Node {
	for _, node := range c.Nodes {
		if node.Id == id {
			return &node
		}
	}

	return nil
}

func (c *Master) LoadNode(params et.Json) error {
	id := params.Key()

	node := c.GetNodeByID(id)
	if node == nil {
		node, err := NewNode(&params)
		if err != nil {
			return err
		}

		driver := node.Data.Str("driver")
		host := node.Data.Str("host")
		port := node.Data.Int("port")
		dbname := node.Data.Str("dbname")
		user := node.Data.Str("user")
		password := node.Data.Str("password")

		idx, err := jdb.Connected(driver, host, port, dbname, user, password)
		if err != nil {
			console.Fatal(err)
		}

		node.Db = idx
		node.ConnStr = jdb.DB(idx).ConnStr
		node.Index = len(c.Nodes)
		c.Nodes = append(c.Nodes, *node)

		go node.SyncNode()
	}

	return nil
}

func (c *Master) LoadNodeById(id string) error {
	item, err := GetNodeById(id)
	if err != nil {
		return err
	}

	return c.LoadNode(item.Result)
}

func (c *Master) LoadNodes() error {
	var ok bool = true
	var rows int = 30
	var page int = 1
	for ok {
		ok = false

		offset := (page - 1) * rows
		sql := strs.Format(`
		SELECT A.*,
		0 AS STATUS
		FROM core.NODES A
		ORDER BY A.INDEX
		LIMIT %d OFFSET %d;`, rows, offset)

		items, err := jdb.Query(sql)
		if err != nil {
			return console.Error(err)
		}

		for _, item := range items.Result {
			err = c.LoadNode(item)
			if err != nil {
				return console.Error(err)
			}

			ok = true
		}

		page++
	}

	return nil
}

func (c *Master) UnloadNodeById(id string) error {
	node := c.GetNodeByID(id)
	if node != nil {
		idx := node.Index
		node.Status = NodeStatusIdle
		copy(c.Nodes[idx:], c.Nodes[idx+1:])
		jdb.DBClose(node.Db)
	}

	return nil
}

func (c *Master) GetSyncById(idT string) (et.Item, error) {
	sql := `
  SELECT *
  FROM core.SYNC
  WHERE _IDT=$1
  LIMIT 1;`

	item, err := jdb.QueryOne(sql, idT)
	if err != nil {
		return et.Item{}, err
	}

	return item, nil
}

func (c *Master) SetSync(schema, table, action, node, idT string, data et.Json, query string) (int, error) {
	index := core.NextSerie("main.SYNC")

	sql := `
	INSERT INTO core.SYNC(TABLE_SCHEMA, TABLE_NAME, ACTION, _IDT, _DATA, QUERY, NODE, INDEX)
	VALUES($1, $2, $3, $4, $5)
	ON CONFLICT (_IDT) DO UPDATE SET
	DATE_UPDATE = NOW(),
	ACTION = EXCLUDED.ACTION,
	_DATA = EXCLUDED._DATA,
	QUERY = EXCLUDED.QUERY,
	NODE = EXCLUDED.NODE
	RETURNING *;`

	_, err := jdb.Query(sql, schema, table, action, idT, data.ToString(), query, node, index)
	if err != nil {
		return -1, err
	}

	return index, nil
}
