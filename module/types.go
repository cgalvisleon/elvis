package module

import (
	"github.com/cgalvisleon/elvis/cache"
	"github.com/cgalvisleon/elvis/console"
	"github.com/cgalvisleon/elvis/core"
	"github.com/cgalvisleon/elvis/et"
	"github.com/cgalvisleon/elvis/event"
	"github.com/cgalvisleon/elvis/linq"
	"github.com/cgalvisleon/elvis/msg"
	"github.com/cgalvisleon/elvis/utility"
)

var Types *linq.Model

func DefineTypes() error {
	if err := DefineSchemaModule(); err != nil {
		return console.Panic(err)
	}

	if Types != nil {
		return nil
	}

	Types = linq.NewModel(SchemaModule, "TYPES", "Tabla de tipo", 1)
	Types.DefineColum("date_make", "", "TIMESTAMP", "NOW()")
	Types.DefineColum("date_update", "", "TIMESTAMP", "NOW()")
	Types.DefineColum("_state", "", "VARCHAR(80)", utility.ACTIVE)
	Types.DefineColum("_id", "", "VARCHAR(80)", "-1")
	Types.DefineColum("project_id", "", "VARCHAR(80)", "-1")
	Types.DefineColum("kind", "", "VARCHAR(80)", "")
	Types.DefineColum("name", "", "VARCHAR(250)", "")
	Types.DefineColum("description", "", "TEXT", "")
	Types.DefineColum("_data", "", "JSONB", "{}")
	Types.DefineColum("index", "", "INTEGER", 0)
	Types.DefinePrimaryKey([]string{"_id"})
	Types.DefineForeignKey("project_id", Projects.Col("_id"))
	Types.DefineIndex([]string{
		"date_make",
		"date_update",
		"_state",
		"project_id",
		"kind",
		"name",
		"index",
	})
	Types.OnListener = func(data et.Json) {
		option := data.Str("option")
		_idt := data.Str("_idt")
		if option == "insert" {
			item, err := GetTypeByIdT(_idt)
			if err != nil {
				return
			}

			_id := item.Key("_id")
			event.WsPublish(_id, item.Result, "")
		} else if option == "update" {
			item, err := GetTypeByIdT(_idt)
			if err != nil {
				return
			}

			_id := item.Key("_id")
			cache.Del(_idt)
			cache.Del(_id)
			event.WsPublish(_id, item.Result, "")
		} else if option == "delete" {
			_id, err := cache.Get(_idt, "-1")
			if err != nil {
				return
			}

			cache.Del(_idt)
			cache.Del(_id)
		}
	}

	if err := core.InitModel(Types); err != nil {
		return console.Panic(err)
	}

	return nil
}

/**
* Types
*	Handler for CRUD data
**/
func GetTypeByIdT(_idt string) (et.Item, error) {
	return Types.Data().
		Where(Types.Column("_idt").Eq(_idt)).
		First()
}

func GetTypeByName(kind, name string) (et.Item, error) {
	return Types.Data().
		Where(Types.Column("kind").Eq(kind)).
		And(Types.Column("name").Eq(name)).
		First()
}

func GetTypeById(id string) (et.Item, error) {
	return Types.Data().
		Where(Types.Column("_id").Eq(id)).
		First()
}

func GetTypeByIndex(idx int) (et.Item, error) {
	return Types.Data().
		Where(Types.Column("index").Eq(idx)).
		First()
}

func InitType(projectId, id, state, kind, name, description string) (et.Item, error) {
	if !utility.ValidId(kind) {
		return et.Item{}, console.AlertF(msg.MSG_ATRIB_REQUIRED, "kind")
	}

	if !utility.ValidStr(name, 0, []string{""}) {
		return et.Item{}, console.AlertF(msg.MSG_ATRIB_REQUIRED, "name")
	}

	current, err := GetTypeByName(kind, name)
	if err != nil {
		return et.Item{}, err
	}

	if current.Ok && current.Id() != id {
		return et.Item{
			Ok: current.Ok,
			Result: et.Json{
				"message": msg.RECORD_FOUND,
			},
		}, nil
	}

	id = utility.GenId(id)
	data := et.Json{}
	data["project_id"] = projectId
	data["_id"] = id
	data["kind"] = kind
	data["name"] = name
	data["description"] = description
	return Types.Upsert(data).
		Where(Types.Column("_id").Eq(id)).
		CommandOne()
}

func UpSetType(projectId, id, kind, name, description string) (et.Item, error) {
	if !utility.ValidId(id) {
		return et.Item{}, console.AlertF(msg.MSG_ATRIB_REQUIRED, "_id")
	}

	if !utility.ValidId(kind) {
		return et.Item{}, console.AlertF(msg.MSG_ATRIB_REQUIRED, "kind")
	}

	if !utility.ValidStr(name, 0, []string{""}) {
		return et.Item{}, console.AlertF(msg.MSG_ATRIB_REQUIRED, "name")
	}

	current, err := GetTypeByName(kind, name)
	if err != nil {
		return et.Item{}, err
	}

	if current.Ok && current.Id() != id {
		return et.Item{
			Ok: current.Ok,
			Result: et.Json{
				"message": msg.RECORD_FOUND,
				"_id":     id,
				"index":   current.Index(),
			},
		}, nil
	}

	id = utility.GenId(id)
	data := et.Json{}
	data["project_id"] = projectId
	data["_id"] = id
	data["kind"] = kind
	data["name"] = name
	data["description"] = description
	return Types.Upsert(data).
		Where(Types.Column("_id").Eq(id)).
		CommandOne()
}

func StateType(id, state string) (et.Item, error) {
	if !utility.ValidId(state) {
		return et.Item{}, console.AlertF(msg.MSG_ATRIB_REQUIRED, "state")
	}

	return Types.Update(et.Json{
		"_state": state,
	}).
		Where(Types.Column("_id").Eq(id)).
		And(Types.Column("_state").Neg(state)).
		CommandOne()
}

func DeleteType(id string) (et.Item, error) {
	return StateType(id, utility.FOR_DELETE)
}

func AllTypes(projectId, kind, state, search string, page, rows int, _select string) (et.List, error) {
	if !utility.ValidId(kind) {
		return et.List{}, console.AlertF(msg.MSG_ATRIB_REQUIRED, "kind")
	}

	if state == "" {
		state = utility.ACTIVE
	}

	auxState := state

	if search != "" {
		return Types.Data(_select).
			Where(Types.Column("kind").Eq(kind)).
			And(Types.Column("project_id").In("-1", projectId)).
			And(Types.Concat("NAME:", Types.Column("name"), ":DESCRIPTION", Types.Column("description"), ":DATA:", Types.Column("_data"), ":").Like("%"+search+"%")).
			OrderBy(Types.Column("name"), true).
			List(page, rows)
	} else if auxState == "*" {
		state = utility.FOR_DELETE

		return Types.Data(_select).
			Where(Types.Column("kind").Eq(kind)).
			And(Types.Column("_state").Neg(state)).
			And(Types.Column("project_id").In("-1", projectId)).
			OrderBy(Types.Column("name"), true).
			List(page, rows)
	} else if auxState == "0" {
		return Types.Data(_select).
			Where(Types.Column("kind").Eq(kind)).
			And(Types.Column("_state").In("-1", state)).
			And(Types.Column("project_id").In("-1", projectId)).
			OrderBy(Types.Column("name"), true).
			List(page, rows)
	} else {
		return Types.Data(_select).
			Where(Types.Column("kind").Eq(kind)).
			And(Types.Column("_state").Eq(state)).
			And(Types.Column("project_id").In("-1", projectId)).
			OrderBy(Types.Column("name"), true).
			List(page, rows)
	}
}
