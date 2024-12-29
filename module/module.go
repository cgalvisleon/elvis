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

var Modules *linq.Model
var ModelFolders *linq.Model

func DefineModules() error {
	if err := DefineSchemaModule(); err != nil {
		return console.Panic(err)
	}

	if Modules != nil {
		return nil
	}

	Modules = linq.NewModel(SchemaModule, "MODULES", "Tabla de modulos", 1)
	Modules.DefineColum("date_make", "", "TIMESTAMP", "NOW()")
	Modules.DefineColum("date_update", "", "TIMESTAMP", "NOW()")
	Modules.DefineColum("_state", "", "VARCHAR(80)", utility.ACTIVE)
	Modules.DefineColum("_id", "", "VARCHAR(80)", "-1")
	Modules.DefineColum("name", "", "VARCHAR(250)", "")
	Modules.DefineColum("description", "", "VARCHAR(250)", "")
	Modules.DefineColum("_data", "", "JSONB", "{}")
	Modules.DefineColum("index", "", "INTEGER", 0)
	Modules.DefinePrimaryKey([]string{"_id"})
	Modules.DefineIndex([]string{
		"date_make",
		"date_update",
		"_state",
		"name",
		"index",
	})
	Modules.Trigger(linq.AfterInsert, func(model *linq.Model, old, new *et.Json, data et.Json) error {
		id := new.Id()
		InitProfile(id, "PROFILE.ADMIN", et.Json{})
		InitProfile(id, "PROFILE.DEV", et.Json{})
		InitProfile(id, "PROFILE.SUPORT", et.Json{})
		CheckProjectModule("-1", id, true)
		CheckRole("-1", id, "PROFILE.ADMIN", "USER.ADMIN", true)

		return nil
	})
	Modules.OnListener = func(data et.Json) {
		option := data.Str("option")
		_idt := data.Str("_idt")
		if option == "insert" {
			item, err := GetModuleByIdT(_idt)
			if err != nil {
				return
			}

			_id := item.Key("_id")
			event.WsPublish(_id, item.Result, "")
		} else if option == "update" {
			item, err := GetModuleByIdT(_idt)
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

	if err := core.InitModel(Modules); err != nil {
		return console.Panic(err)
	}

	return nil
}

func DefineModuleFolders() error {
	if err := DefineSchemaModule(); err != nil {
		return console.Panic(err)
	}

	if ModelFolders != nil {
		return nil
	}

	ModelFolders = linq.NewModel(SchemaModule, "MODULE_FOLDERS", "Tabla de folders por modulo", 1)
	ModelFolders.DefineColum("date_make", "", "TIMESTAMP", "NOW()")
	ModelFolders.DefineColum("module_id", "", "VARCHAR(80)", "-1")
	ModelFolders.DefineColum("folder_id", "", "VARCHAR(80)", "-1")
	ModelFolders.DefineColum("index", "", "INTEGER", 0)
	ModelFolders.DefinePrimaryKey([]string{"module_id", "folder_id"})
	ModelFolders.DefineForeignKey("module_id", Modules.Col("_id"))
	ModelFolders.DefineForeignKey("folder_id", Folders.Col("_id"))
	Modules.DefineIndex([]string{
		"date_make",
		"index",
	})
	Modules.OnListener = func(data et.Json) {
		option := data.Str("option")
		_idt := data.Str("_idt")
		if option == "insert" {
			item, err := GetModuleFolderByIdT(_idt)
			if err != nil {
				return
			}

			_id := item.Key("module_id") + "/" + item.Key("folder_id")
			event.WsPublish(_id, item.Result, "")
		} else if option == "update" {
			item, err := GetModuleFolderByIdT(_idt)
			if err != nil {
				return
			}

			_id := item.Key("module_id") + "/" + item.Key("folder_id")
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

	if err := core.InitModel(ModelFolders); err != nil {
		return console.Panic(err)
	}

	return nil
}

/**
* Module
*	Handler for CRUD data
**/
func GetModuleByIdT(_idt string) (et.Item, error) {
	return Modules.Data().
		Where(Modules.Column("_idt").Eq(_idt)).
		First()
}

func GetModuleByName(name string) (et.Item, error) {
	return Modules.Data().
		Where(Modules.Column("name").Eq(name)).
		First()
}

func GetModuleById(id string) (et.Item, error) {
	return Modules.Data().
		Where(Modules.Column("_id").Eq(id)).
		First()
}

func IsInit() (et.Item, error) {
	count := Users.Data().
		Count()

	return et.Item{
		Ok: count > 0,
		Result: et.Json{
			"message": utility.OkOrNot(count > 0, msg.SYSTEM_HAVE_ADMIN, msg.SYSTEM_NOT_HAVE_ADMIN),
		},
	}, nil
}

func InitModule(id, name, description string, data et.Json) (et.Item, error) {
	if !utility.ValidStr(name, 0, []string{""}) {
		return et.Item{}, console.AlertF(msg.MSG_ATRIB_REQUIRED, "name")
	}

	current, err := GetModuleByName(name)
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
	data.Set("_id", id)
	data.Set("name", name)
	data.Set("description", description)
	item, err := Modules.Upsert(data).
		Where(Modules.Column("_id").Eq(id)).
		CommandOne()
	if err != nil {
		return et.Item{}, err
	}

	return item, nil
}

func UpSetModule(id, name, description string, data et.Json) (et.Item, error) {
	if !utility.ValidStr(name, 0, []string{""}) {
		return et.Item{}, console.AlertF(msg.MSG_ATRIB_REQUIRED, "name")
	}

	current, err := GetModuleByName(name)
	if err != nil {
		return et.Item{}, err
	}

	if current.Ok && current.Id() != id {
		return et.Item{
			Ok: current.Ok,
			Result: et.Json{
				"message": msg.RECORD_FOUND,
				"_id":     current.Id(),
				"index":   current.Index(),
			},
		}, nil
	}

	id = utility.GenId(id)
	data.Set("_id", id)
	data.Set("name", name)
	data.Set("description", description)
	item, err := Modules.Upsert(data).
		Where(Modules.Column("_id").Eq(id)).
		CommandOne()
	if err != nil {
		return et.Item{}, err
	}

	return item, nil
}

func StateModule(id, state string) (et.Item, error) {
	if !utility.ValidId(state) {
		return et.Item{}, console.AlertF(msg.MSG_ATRIB_REQUIRED, "state")
	}

	return Modules.Update(et.Json{
		"_state": state,
	}).
		Where(Modules.Column("_id").Eq(id)).
		And(Modules.Column("_state").Neg(state)).
		CommandOne()
}

func DeleteModule(id string) (et.Item, error) {
	return StateModule(id, utility.FOR_DELETE)
}

func AllModules(state, search string, page, rows int, _select string) (et.List, error) {
	if state == "" {
		state = utility.ACTIVE
	}

	auxState := state

	if search != "" {
		return Modules.Data(_select).
			Where(Modules.Concat("NAME:", Modules.Column("name"), ":DESCRIPTION", Modules.Column("description"), ":DATA:", Modules.Column("_data"), ":").Like("%"+search+"%")).
			OrderBy(Modules.Column("name"), true).
			List(page, rows)
	} else if auxState == "*" {
		state = utility.FOR_DELETE

		return Modules.Data(_select).
			Where(Modules.Column("_state").Neg(state)).
			OrderBy(Modules.Column("name"), true).
			List(page, rows)
	} else if auxState == "0" {
		return Modules.Data(_select).
			Where(Modules.Column("_state").In("-1", state)).
			OrderBy(Modules.Column("name"), true).
			List(page, rows)
	} else {
		return Modules.Data(_select).
			Where(Modules.Column("_state").Eq(state)).
			OrderBy(Modules.Column("name"), true).
			List(page, rows)
	}
}

// Module Folder
func GetModuleFolderByIdT(_idt string) (et.Item, error) {
	return ModelFolders.Data().
		Where(ModelFolders.Column("_idt").Eq(_idt)).
		First()
}

func GetModuleFolderById(module_id, folder_id string) (et.Item, error) {
	return ModelFolders.Data().
		Where(ModelFolders.Column("module_id").Eq(module_id)).
		And(ModelFolders.Column("folder_id").Eq(folder_id)).
		First()
}

// Check folder that module
func CheckModuleFolder(module_id, folder_id string, chk bool) (et.Item, error) {
	if !utility.ValidId(module_id) {
		return et.Item{}, console.AlertF(msg.MSG_ATRIB_REQUIRED, "module_id")
	}

	if !utility.ValidId(folder_id) {
		return et.Item{}, console.AlertF(msg.MSG_ATRIB_REQUIRED, "folder_id")
	}

	data := et.Json{}
	data.Set("module_id", module_id)
	data.Set("folder_id", folder_id)
	if chk {
		current, err := GetModuleFolderById(module_id, folder_id)
		if err != nil {
			return et.Item{}, err
		}

		if current.Ok {
			return et.Item{
				Ok: current.Ok,
				Result: et.Json{
					"message": msg.RECORD_NOT_UPDATE,
					"index":   current.Index(),
				},
			}, nil
		}

		return ModelFolders.Insert(data).
			CommandOne()
	} else {
		return ModelFolders.Delete().
			Where(ModelFolders.Column("module_id").Eq(module_id)).
			And(ModelFolders.Column("folder_id").Eq(folder_id)).
			CommandOne()
	}
}
