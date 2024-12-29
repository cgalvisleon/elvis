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

var Folders *linq.Model

func DefineFolders() error {
	if err := DefineSchemaModule(); err != nil {
		return console.Panic(err)
	}

	if Folders != nil {
		return nil
	}

	Folders = linq.NewModel(SchemaModule, "FOLDERS", "Tabla de carpetas", 1)
	Folders.DefineColum("date_make", "", "TIMESTAMP", "NOW()")
	Folders.DefineColum("date_update", "", "TIMESTAMP", "NOW()")
	Folders.DefineColum("module_id", "", "VARCHAR(80)", "-1")
	Folders.DefineColum("_state", "", "VARCHAR(80)", utility.ACTIVE)
	Folders.DefineColum("_id", "", "VARCHAR(80)", "-1")
	Folders.DefineColum("main_id", "", "VARCHAR(80)", "-1")
	Folders.DefineColum("name", "", "VARCHAR(250)", "")
	Folders.DefineColum("description", "", "VARCHAR(250)", "")
	Folders.DefineColum("_data", "", "JSONB", "{}")
	Folders.DefineColum("index", "", "INTEGER", 0)
	Folders.DefinePrimaryKey([]string{"_id"})
	Folders.DefineForeignKey("module_id", Modules.Col("_id"))
	Folders.DefineIndex([]string{
		"date_make",
		"date_update",
		"_state",
		"main_id",
		"name",
		"index",
	})
	Folders.Trigger(linq.AfterInsert, func(model *linq.Model, old, new *et.Json, data et.Json) error {
		id := new.Id()
		if id != "-1" {
			moduleId := new.Key("module_id")
			CheckProfileFolder(moduleId, "PROFILE.ADMIN", id, true)
			CheckProfileFolder(moduleId, "PROFILE.DEV", id, true)
			CheckProfileFolder(moduleId, "PROFILE.SUPORT", id, true)
			CheckModuleFolder(moduleId, id, true)
		}

		return nil
	})
	Folders.Trigger(linq.AfterUpdate, func(model *linq.Model, old, new *et.Json, data et.Json) error {
		event.Action("folder/update", *new)
		oldState := old.Key("_state")
		newState := old.Key("_state")
		if oldState != newState {
			event.Action("folder/state", *new)
		}

		return nil
	})
	Folders.Trigger(linq.AfterDelete, func(model *linq.Model, old, new *et.Json, data et.Json) error {
		event.Action("folder/delete", *old)

		return nil
	})
	Folders.OnListener = func(data et.Json) {
		option := data.Str("option")
		_idt := data.Str("_idt")
		if option == "insert" {
			item, err := GetFolderByIdT(_idt)
			if err != nil {
				return
			}

			_id := item.Key("_id")
			event.WsPublish(_id, item.Result, "")
		} else if option == "update" {
			item, err := GetFolderByIdT(_idt)
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

	if err := core.InitModel(Folders); err != nil {
		return console.Panic(err)
	}

	return nil
}

/**
*	Folder
*	Handler for CRUD data
**/
func GetFolderByIdT(_idt string) (et.Item, error) {
	return Folders.Data().
		Where(Folders.Column("_idt").Eq(_idt)).
		First()
}

func GetFolderByName(moduleId, mainId, name string) (et.Item, error) {
	return Folders.Data().
		Where(Folders.Column("module_id").Eq(moduleId)).
		And(Folders.Column("main_id").Eq(mainId)).
		And(Folders.Column("name").Eq(name)).
		First()
}

func InitFolder(moduleId, mainId, id, name, description string, data et.Json) (et.Item, error) {
	if !utility.ValidId(moduleId) {
		return et.Item{}, console.AlertF(msg.MSG_ATRIB_REQUIRED, "module_id")
	}

	if !utility.ValidId(mainId) {
		return et.Item{}, console.AlertF(msg.MSG_ATRIB_REQUIRED, "main_id")
	}

	if !utility.ValidStr(name, 0, []string{""}) {
		return et.Item{}, console.AlertF(msg.MSG_ATRIB_REQUIRED, "name")
	}

	module, err := GetModuleById(moduleId)
	if err != nil {
		return et.Item{}, err
	}

	if !module.Ok {
		return et.Item{}, console.ErrorM(msg.MODULE_NOT_FOUND)
	}

	current, err := GetFolderByName(moduleId, mainId, name)
	if err != nil {
		return et.Item{}, err
	}

	if current.Ok && current.Id() != id {
		return et.Item{
			Ok: current.Ok,
			Result: et.Json{
				"message": msg.RECORD_FOUND,
				"_id":     id,
			},
		}, nil
	}

	id = utility.GenId(id)
	data["module_id"] = moduleId
	data["main_id"] = mainId
	data["_id"] = id
	data["name"] = name
	data["description"] = description
	item, err := Folders.Upsert(data).
		Where(Folders.Column("_id").Eq(id)).
		And(Folders.Column("_state").Eq(utility.ACTIVE)).
		CommandOne()
	if err != nil {
		return et.Item{}, err
	}

	CheckModuleFolder(moduleId, id, true)

	return item, nil
}

func UpSetFolder(moduleId, mainId, name, description string, data et.Json) (et.Item, error) {
	if !utility.ValidId(moduleId) {
		return et.Item{}, console.AlertF(msg.MSG_ATRIB_REQUIRED, "module_id")
	}

	if !utility.ValidId(mainId) {
		return et.Item{}, console.AlertF(msg.MSG_ATRIB_REQUIRED, "main_id")
	}

	if !utility.ValidStr(name, 0, []string{""}) {
		return et.Item{}, console.AlertF(msg.MSG_ATRIB_REQUIRED, "name")
	}

	module, err := GetModuleById(moduleId)
	if err != nil {
		return et.Item{}, err
	}

	if !module.Ok {
		return et.Item{}, console.ErrorM(msg.MODULE_NOT_FOUND)
	}

	current, err := Folders.Data(Folders.Column("_id")).
		Where(Folders.Column("module_id").Eq(moduleId)).
		And(Folders.Column("main_id").Eq(mainId)).
		And(Folders.Column("name").Eq(name)).
		First()
	if err != nil {
		return et.Item{}, err
	}

	id := current.Id()
	id = utility.GenId(id)
	data["module_id"] = moduleId
	data["main_id"] = mainId
	data["_id"] = id
	data["name"] = name
	data["description"] = description
	item, err := Folders.Upsert(data).
		Where(Folders.Column("_id").Eq(id)).
		And(Folders.Column("_state").Eq(utility.ACTIVE)).
		CommandOne()
	if err != nil {
		return et.Item{}, err
	}

	return item, nil
}

func GetFolderById(id string) (et.Item, error) {
	return Folders.Data().
		Where(Folders.Column("_id").Eq(id)).
		First()
}

func StateFolder(id, state string) (et.Item, error) {
	if !utility.ValidId(state) {
		return et.Item{}, console.AlertF(msg.MSG_ATRIB_REQUIRED, "state")
	}

	item, err := Folders.Update(et.Json{
		"_state": state,
	}).
		Where(Folders.Column("_id").Eq(id)).
		And(Folders.Column("_state").Neg(state)).
		CommandOne()
	if err != nil {
		return et.Item{}, err
	}

	return item, nil
}

func DeleteFolder(id string) (et.Item, error) {
	item, err := Folders.Delete().
		Where(Folders.Column("_id").Eq(id)).
		CommandOne()
	if err != nil {
		return et.Item{}, err
	}

	return item, nil
}

func AllFolders(state, search string, page, rows int) (et.List, error) {
	if state == "" {
		state = utility.ACTIVE
	}

	auxState := state

	if search != "" {
		return Folders.Data().
			Where(Folders.Concat("NAME:", Folders.Column("name"), ":DESCRIPTION", Folders.Column("description"), ":DATA:", Folders.Column("_data"), ":").Like("%"+search+"%")).
			OrderBy(Folders.Column("name"), true).
			List(page, rows)
	} else if auxState == "*" {
		state = utility.FOR_DELETE

		return Folders.Data().
			Where(Folders.Column("_state").Neg(state)).
			OrderBy(Folders.Column("name"), true).
			List(page, rows)
	} else if auxState == "0" {
		return Folders.Data().
			Where(Folders.Column("_state").In("-1", state)).
			OrderBy(Folders.Column("name"), true).
			List(page, rows)
	} else {
		return Folders.Data().
			Where(Folders.Column("_state").Eq(state)).
			OrderBy(Folders.Column("name"), true).
			List(page, rows)
	}
}
