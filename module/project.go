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

var Projects *linq.Model
var ProjectModules *linq.Model

func DefineProjects() error {
	if err := DefineSchemaModule(); err != nil {
		return console.Panic(err)
	}

	if Projects != nil {
		return nil
	}

	Projects = linq.NewModel(SchemaModule, "PROJECTS", "Tabla de projectos", 1)
	Projects.DefineColum("date_make", "", "TIMESTAMP", "NOW()")
	Projects.DefineColum("date_update", "", "TIMESTAMP", "NOW()")
	Projects.DefineColum("_state", "", "VARCHAR(80)", utility.ACTIVE)
	Projects.DefineColum("_id", "", "VARCHAR(80)", "-1")
	Projects.DefineColum("name", "", "VARCHAR(250)", "")
	Projects.DefineColum("description", "", "VARCHAR(250)", "")
	Projects.DefineColum("_data", "", "JSONB", "{}")
	Projects.DefineColum("index", "", "INTEGER", 0)
	Projects.DefinePrimaryKey([]string{"_id"})
	Projects.DefineIndex([]string{
		"date_make",
		"date_update",
		"_state",
		"name",
		"index",
	})
	Projects.Trigger(linq.AfterInsert, func(model *linq.Model, old, new *et.Json, data et.Json) error {
		moduleId := data.Key("module_id")
		id := new.Id()
		if id != "-1" {
			CheckProjectModule(id, moduleId, true)
			CheckRole(id, moduleId, "PROFILE.ADMIN", "USER.ADMIN", true)
		}

		return nil
	})
	Projects.OnListener = func(data et.Json) {
		option := data.Str("option")
		_idt := data.Str("_idt")
		if option == "insert" {
			item, err := GetProjectByIdT(_idt)
			if err != nil {
				return
			}

			_id := item.Key("_id")
			event.WsPublish(_id, item.Result, "")
		} else if option == "update" {
			item, err := GetProjectByIdT(_idt)
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

	if err := core.InitModel(Projects); err != nil {
		return console.Panic(err)
	}

	return nil
}

func DefineProjectModules() error {
	if err := DefineSchemaModule(); err != nil {
		return console.Panic(err)
	}

	if ProjectModules != nil {
		return nil
	}

	ProjectModules = linq.NewModel(SchemaModule, "PROJECT_MODULES", "Tabla de moduloes por projecto", 1)
	ProjectModules.DefineColum("date_make", "", "TIMESTAMP", "NOW()")
	ProjectModules.DefineColum("project_id", "", "VARCHAR(80)", "-1")
	ProjectModules.DefineColum("module_id", "", "VARCHAR(80)", "-1")
	ProjectModules.DefineColum("index", "", "INTEGER", 0)
	ProjectModules.DefinePrimaryKey([]string{"project_id", "module_id"})
	ProjectModules.DefineForeignKey("project_id", Projects.Col("_id"))
	ProjectModules.DefineForeignKey("module_id", Modules.Col("_id"))
	ProjectModules.DefineIndex([]string{
		"date_make",
		"index",
	})
	ProjectModules.OnListener = func(data et.Json) {
		option := data.Str("option")
		_idt := data.Str("_idt")
		if option == "insert" {
			item, err := GetProjectModulesByIdT(_idt)
			if err != nil {
				return
			}

			_id := item.Key("project_id") + "/" + item.Key("module_id")
			event.WsPublish(_id, item.Result, "")
		} else if option == "update" {
			item, err := GetProjectModulesByIdT(_idt)
			if err != nil {
				return
			}

			_id := item.Key("project_id") + "/" + item.Key("module_id")
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

	if err := core.InitModel(ProjectModules); err != nil {
		return console.Panic(err)
	}

	return nil
}

/**
* Project
*	Handler for CRUD data
**/
func GetProjectByIdT(_idt string) (et.Item, error) {
	return Projects.Data().
		Where(Projects.Column("_idt").Eq(_idt)).
		First()
}

func GetProjectById(id string) (et.Item, error) {
	return Projects.Data().
		Where(Projects.Column("_id").Eq(id)).
		First()
}

func GetProjectName(name string) (et.Item, error) {
	return Projects.Data().
		Where(Projects.Column("name").Eq(name)).
		First()
}

/**
* ProjectModules
**/
func GetProjectModulesByIdT(_idt string) (et.Item, error) {
	return ProjectModules.Data().
		Where(ProjectModules.Column("_idt").Eq(_idt)).
		First()
}

func GetProjectByModule(projectId, moduleId string) (et.Item, error) {
	return ProjectModules.Data(ProjectModules.Column("index")).
		Where(ProjectModules.Column("project_id").Eq(projectId)).
		And(ProjectModules.Column("module_id").Eq(moduleId)).
		First()
}

func InitProject(id, name, description string, data et.Json) (et.Item, error) {
	if !utility.ValidStr(name, 0, []string{""}) {
		return et.Item{}, console.AlertF(msg.MSG_ATRIB_REQUIRED, "name")
	}

	id = utility.GenId(id)
	data.Set("_id", id)
	data.Set("name", name)
	data.Set("description", description)
	item, err := Projects.Upsert(data).
		Where(Projects.Column("_id").Eq(id)).
		CommandOne()
	if err != nil {
		return et.Item{}, err
	}

	return item, nil
}

func UpSetProject(id, moduleId, name, description string, data et.Json) (et.Item, error) {
	if !utility.ValidId(moduleId) {
		return et.Item{}, console.AlertF(msg.MSG_ATRIB_REQUIRED, "module_id")
	}

	if !utility.ValidStr(name, 0, []string{""}) {
		return et.Item{}, console.AlertF(msg.MSG_ATRIB_REQUIRED, "name")
	}

	current, err := GetProjectName(name)
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
	data.Set("module_id", moduleId)
	item, err := Projects.Upsert(data).
		Where(Projects.Column("_id").Eq(id)).
		CommandOne()
	if err != nil {
		return et.Item{}, err
	}

	return item, nil
}

func StateProject(id, state string) (et.Item, error) {
	if !utility.ValidId(state) {
		return et.Item{}, console.AlertF(msg.MSG_ATRIB_REQUIRED, "state")
	}

	return Projects.Update(et.Json{
		"_state": state,
	}).
		Where(Projects.Column("_id").Eq(id)).
		And(Projects.Column("_state").Neg(state)).
		CommandOne()
}

func DeleteProject(id string) (et.Item, error) {
	return StateProject(id, utility.FOR_DELETE)
}

func AllProjects(state, search string, page, rows int, _select string) (et.List, error) {
	if state == "" {
		state = utility.ACTIVE
	}

	auxState := state

	if search != "" {
		return Projects.Data(_select).
			Where(Projects.Concat("NAME:", Projects.Column("name"), ":DESCRIPTION:", Projects.Column("description"), ":DATA:", Projects.Column("_data"), ":").Like("%"+search+"%")).
			OrderBy(Projects.Column("name"), true).
			List(page, rows)
	} else if auxState == "*" {
		state = utility.FOR_DELETE

		return Projects.Data(_select).
			Where(Projects.Column("_state").Neg(state)).
			OrderBy(Projects.Column("name"), true).
			List(page, rows)
	} else if auxState == "0" {
		return Projects.Data(_select).
			Where(Projects.Column("_state").In("-1", state)).
			OrderBy(Projects.Column("name"), true).
			List(page, rows)
	} else {
		return Projects.Data(_select).
			Where(Projects.Column("_state").Eq(state)).
			OrderBy(Projects.Column("name"), true).
			List(page, rows)
	}
}

func GetProjectModules(projectId, state, search string, page, rows int) (et.List, error) {
	if state == "" {
		state = utility.ACTIVE
	}

	auxState := state

	if auxState == "*" {
		state = utility.FOR_DELETE

		return linq.From(Modules, "A").
			Join(Modules.As("A"), ProjectModules.As("B"), ProjectModules.Col("module_id").Eq(Modules.Col("_id"))).
			Where(Modules.Column("_state").Neg(state)).
			And(ProjectModules.Column("project_id").Eq(projectId)).
			And(Modules.Concat("NAME:", Modules.Column("name"), ":DESCRIPTION", Modules.Column("description"), ":DATA:", Modules.Column("_data"), ":").Like("%"+search+"%")).
			OrderBy(Modules.Column("name"), true).
			Data().
			List(page, rows)
	} else if auxState == "0" {
		return linq.From(Modules, "A").
			Join(Modules.As("A"), ProjectModules.As("B"), ProjectModules.Col("module_id").Eq(Modules.Col("_id"))).
			Where(Modules.Column("_state").In("-1", state)).
			And(ProjectModules.Column("project_id").Eq(projectId)).
			And(Modules.Concat("NAME:", Modules.Column("name"), ":DESCRIPTION", Modules.Column("description"), ":DATA:", Modules.Column("_data"), ":").Like("%"+search+"%")).
			OrderBy(Modules.Column("name"), true).
			Data().
			List(page, rows)
	} else {
		return linq.From(Modules, "A").
			Join(Modules.As("A"), ProjectModules.As("B"), ProjectModules.Col("module_id").Eq(Modules.Col("_id"))).
			Where(Modules.Column("_state").Eq(state)).
			And(ProjectModules.Column("project_id").Eq(projectId)).
			And(Modules.Concat("NAME:", Modules.Column("name"), ":DESCRIPTION", Modules.Column("description"), ":DATA:", Modules.Column("_data"), ":").Like("%"+search+"%")).
			OrderBy(Modules.Column("name"), true).
			Data().
			List(page, rows)
	}
}

func CheckProjectModule(project_id, module_id string, chk bool) (et.Item, error) {
	if !utility.ValidId(project_id) {
		return et.Item{}, console.AlertF(msg.MSG_ATRIB_REQUIRED, "project_id")
	}

	if !utility.ValidId(module_id) {
		return et.Item{}, console.AlertF(msg.MSG_ATRIB_REQUIRED, "module_id")
	}

	data := et.Json{}
	data.Set("project_id", project_id)
	data.Set("module_id", module_id)
	if chk {
		current, err := GetProjectByModule(project_id, module_id)
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

		return ProjectModules.Insert(data).
			CommandOne()
	} else {
		return ProjectModules.Delete().
			Where(ProjectModules.Column("project_id").Eq(project_id)).
			And(ProjectModules.Column("module_id").Eq(module_id)).
			CommandOne()
	}
}
