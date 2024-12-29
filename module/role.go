package module

import (
	"github.com/cgalvisleon/elvis/cache"
	"github.com/cgalvisleon/elvis/console"
	"github.com/cgalvisleon/elvis/core"
	"github.com/cgalvisleon/elvis/et"
	"github.com/cgalvisleon/elvis/event"
	"github.com/cgalvisleon/elvis/jdb"
	"github.com/cgalvisleon/elvis/linq"
	"github.com/cgalvisleon/elvis/msg"
	"github.com/cgalvisleon/elvis/utility"
)

var Roles *linq.Model

func DefineRoles() error {
	if err := DefineSchemaModule(); err != nil {
		return console.Panic(err)
	}

	if Roles != nil {
		return nil
	}

	Roles = linq.NewModel(SchemaModule, "ROLES", "Tabla de roles", 1)
	Roles.DefineColum("date_make", "", "TIMESTAMP", "NOW()")
	Roles.DefineColum("date_update", "", "TIMESTAMP", "NOW()")
	Roles.DefineColum("project_id", "", "VARCHAR(80)", "-1")
	Roles.DefineColum("module_id", "", "VARCHAR(80)", "-1")
	Roles.DefineColum("user_id", "", "VARCHAR(80)", "-1")
	Roles.DefineColum("profile_tp", "", "VARCHAR(80)", "-1")
	Roles.DefineColum("index", "", "INTEGER", 0)
	Roles.DefinePrimaryKey([]string{"project_id", "module_id", "user_id"})
	Roles.DefineIndex([]string{
		"date_make",
		"date_update",
		"profile_tp",
		"index",
	})
	Roles.OnListener = func(data et.Json) {
		option := data.Str("option")
		_idt := data.Str("_idt")
		if option == "insert" {
			item, err := GetRoleByIdT(_idt)
			if err != nil {
				return
			}

			_id := item.Key("project_id") + "/" + item.Key("module_id") + "/" + item.Key("user_id")
			event.WsPublish(_id, item.Result, "")
		} else if option == "update" {
			item, err := GetRoleByIdT(_idt)
			if err != nil {
				return
			}

			_id := item.Key("project_id") + "/" + item.Key("module_id") + "/" + item.Key("user_id")
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

	if err := core.InitModel(Roles); err != nil {
		return console.Panic(err)
	}

	return nil
}

/**
* Role
*	Handler for CRUD data
**/
func GetRoleByIdT(_idt string) (et.Item, error) {
	return Roles.Data().
		Where(Roles.Column("_idt").Eq(_idt)).
		First()
}

func GetRoleById(projectId, moduleId, userId, profileTp string) (et.Item, error) {
	return Roles.Data().
		Where(Roles.Column("project_id").Eq(projectId)).
		And(Roles.Column("module_id").Eq(moduleId)).
		And(Roles.Column("user_id").Eq(userId)).
		And(Roles.Column("profile_tp").Eq(profileTp)).
		First()
}

func GetUserRoleByIndex(idx int) (et.Item, error) {
	sql := `
	SELECT
	D._ID AS PROJECT_ID,
	D.NAME AS PROJECT,
	B._ID AS MODULE_ID,
	B.NAME AS MODULE,
	A.PROFILE_TP,
	C.NAME PROFILE,
	A.USER_ID,
	A.INDEX
	FROM module.ROLES A
	INNER JOIN module.MODULES B ON B._ID=A.MODULE_ID
	INNER JOIN module.TYPES C ON C._ID=A.PROFILE_TP
	INNER JOIN module.PROJECTS D ON D._ID=A.PROJECT_ID
	WHERE A.INDEX=$1
	LIMIT 1;`

	item, err := jdb.QueryOne(sql, idx)
	if err != nil {
		return et.Item{}, err
	}

	return item, nil
}

func GetUserProjects(userId string) ([]et.Json, error) {
	sql := `
	SELECT
	B._ID,
	B.NAME,
	MIN(A.INDEX) AS INDEX
	FROM module.ROLES A	
	INNER JOIN module.PROJECTS B ON B._ID=A.PROJECT_ID
	WHERE A.USER_ID=$1
	GROUP BY B._ID, B.NAME
	ORDER BY B.NAME;`

	modules, err := jdb.Query(sql, userId)
	if err != nil {
		return []et.Json{}, err
	}

	return modules.Result, nil
}

func GetUserModules(userId string) ([]et.Json, error) {
	sql := `
	SELECT
	D._ID AS PROJECT_ID,
	D.NAME AS PROJECT,
	B._ID AS MODULE_ID,
	B.NAME AS MODULE,
	A.PROFILE_TP,
	C.NAME PROFILE,
	A.USER_ID,
	A.INDEX
	FROM module.ROLES A
	INNER JOIN module.MODULES B ON B._ID=A.MODULE_ID
	INNER JOIN module.TYPES C ON C._ID=A.PROFILE_TP
	INNER JOIN module.PROJECTS D ON D._ID=A.PROJECT_ID
	WHERE A.USER_ID=$1
	GROUP BY D._ID, D.NAME, B._ID, B.NAME, A.PROFILE_TP, C.NAME, USER_ID, A.INDEX
	ORDER BY D.NAME, B.NAME, C.NAME;`

	modules, err := jdb.Query(sql, userId)
	if err != nil {
		return []et.Json{}, err
	}

	return modules.Result, nil
}

func CheckRole(projectId, moduleId, profileTp, userId string, chk bool) (et.Item, error) {
	if !utility.ValidId(projectId) {
		return et.Item{}, console.AlertF(msg.MSG_ATRIB_REQUIRED, "project_id")
	}

	if !utility.ValidId(moduleId) {
		return et.Item{}, console.AlertF(msg.MSG_ATRIB_REQUIRED, "module_id")
	}

	if !utility.ValidId(userId) {
		return et.Item{}, console.AlertF(msg.MSG_ATRIB_REQUIRED, "user_id")
	}

	if !utility.ValidId(profileTp) {
		return et.Item{}, console.AlertF(msg.MSG_ATRIB_REQUIRED, "profile_tp")
	}

	project, err := GetProjectById(projectId)
	if err != nil {
		return et.Item{}, err
	}

	if !project.Ok {
		return et.Item{}, console.AlertF(msg.PROJECT_NOT_FOUND, projectId)
	}

	module, err := GetModuleById(moduleId)
	if err != nil {
		return et.Item{}, err
	}

	if !module.Ok {
		return et.Item{}, console.Alert(msg.MODULE_NOT_FOUND)
	}

	profile, err := GetProfileById(moduleId, profileTp)
	if err != nil {
		return et.Item{}, err
	}

	if !profile.Ok {
		return et.Item{}, console.AlertF(msg.PROFILE_NOT_FOUND, profileTp)
	}

	if chk {
		current, err := GetRoleById(projectId, moduleId, userId, profileTp)
		if err != nil {
			return et.Item{}, err
		}

		now := utility.Now()
		if current.Ok {
			index := current.Index()
			sql := `
			UPDATE module.ROLES SET
			DATE_UPDATE=$3,
			PROFILE_TP=$2
			WHERE INDEX=$1;`

			item, err := jdb.QueryOne(sql, index, profileTp, now)
			if err != nil {
				return et.Item{}, err
			}

			item, err = GetUserRoleByIndex(index)
			if err != nil {
				return et.Item{}, err
			}

			return et.Item{
				Ok: item.Ok,
				Result: et.OkOrNotJson(item.Ok, item.Result, et.Json{
					"message": msg.RECORD_NOT_UPDATE,
					"index":   index,
				}),
			}, nil
		}

		index := core.NextSerie("module.ROLES")

		sql := `
		INSERT INTO module.ROLES(DATE_MAKE, DATE_UPDATE, PROJECT_ID, MODULE_ID, USER_ID, PROFILE_TP, INDEX)
		VALUES($1, $1, $2, $3, $4, $5, $6)
		RETURNING INDEX;`

		item, err := jdb.QueryOne(sql, now, projectId, moduleId, userId, profileTp, index)
		if err != nil {
			return et.Item{}, err
		}

		item, err = GetUserRoleByIndex(index)
		if err != nil {
			return et.Item{}, err
		}

		return et.Item{
			Ok: item.Ok,
			Result: et.OkOrNotJson(item.Ok, item.Result, et.Json{
				"message": msg.RECORD_NOT_UPDATE,
				"index":   index,
			}),
		}, nil
	} else {
		sql := `
		DELETE FROM module.ROLES
		WHERE PROJECT_ID=$1
		AND MODULE_ID=$2
		AND PROFILE_TP=$3
		AND USER_ID=$4
		RETURNING INDEX;`

		item, err := jdb.QueryOne(sql, projectId, moduleId, profileTp, userId)
		if err != nil {
			return et.Item{}, err
		}

		return et.Item{
			Ok: item.Ok,
			Result: et.Json{
				"message": utility.OkOrNot(item.Ok, msg.RECORD_DELETE, msg.RECORD_NOT_DELETE),
				"index":   item.Index(),
			},
		}, nil
	}
}
