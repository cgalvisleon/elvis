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

var Profiles *linq.Model
var ProfileFolders *linq.Model

func DefineProfiles() error {
	if err := DefineSchemaModule(); err != nil {
		return console.Panic(err)
	}

	if Profiles != nil {
		return nil
	}

	Profiles = linq.NewModel(SchemaModule, "PROFILES", "Tabla de perfiles", 1)
	Profiles.DefineColum("date_make", "", "TIMESTAMP", "NOW()")
	Profiles.DefineColum("date_update", "", "TIMESTAMP", "NOW()")
	Profiles.DefineColum("module_id", "", "VARCHAR(80)", "-1")
	Profiles.DefineColum("profile_tp", "", "VARCHAR(80)", "-1")
	Profiles.DefineColum("_data", "", "JSONB", "{}")
	Profiles.DefineColum("index", "", "INTEGER", 0)
	Profiles.DefinePrimaryKey([]string{"module_id", "profile_tp"})
	Profiles.DefineIndex([]string{
		"date_make",
		"date_update",
		"index",
	})
	Profiles.DefineForeignKey("module_id", Modules.Column("_id"))
	Profiles.OnListener = func(data et.Json) {
		option := data.Str("option")
		_idt := data.Str("_idt")
		if option == "insert" {
			item, err := GetProfileByIdT(_idt)
			if err != nil {
				return
			}

			_id := item.Key("module_id") + "/" + item.Key("profile_tp")
			event.WsPublish(_id, item.Result, "")
		} else if option == "update" {
			item, err := GetProfileByIdT(_idt)
			if err != nil {
				return
			}

			_id := item.Key("module_id") + "/" + item.Key("profile_tp")
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

	if err := core.InitModel(Profiles); err != nil {
		return console.Panic(err)
	}

	return nil
}

func DefineProfileFolders() error {
	if err := DefineSchemaModule(); err != nil {
		return console.Panic(err)
	}

	if ProfileFolders != nil {
		return nil
	}

	ProfileFolders = linq.NewModel(SchemaModule, "PROFILE_FOLDERS", "Tabla de carpetas por perfil", 1)
	ProfileFolders.DefineColum("date_make", "", "TIMESTAMP", "NOW()")
	ProfileFolders.DefineColum("module_id", "", "VARCHAR(80)", "-1")
	ProfileFolders.DefineColum("profile_tp", "", "VARCHAR(80)", "-1")
	ProfileFolders.DefineColum("folder_id", "", "VARCHAR(80)", "-1")
	ProfileFolders.DefineColum("index", "", "INTEGER", 0)
	ProfileFolders.DefinePrimaryKey([]string{"module_id", "profile_tp", "folder_id"})
	ProfileFolders.DefineIndex([]string{
		"date_make",
		"index",
	})
	ProfileFolders.DefineForeignKey("module_id", Modules.Column("_id"))
	ProfileFolders.DefineForeignKey("folder_id", Folders.Column("_id"))
	ProfileFolders.OnListener = func(data et.Json) {
		option := data.Str("option")
		_idt := data.Str("_idt")
		if option == "insert" {
			item, err := GetProfileFolderByIdT(_idt)
			if err != nil {
				return
			}

			_id := item.Key("module_id") + "/" + item.Key("profile_tp") + "/" + item.Key("folder_id")
			event.WsPublish(_id, item.Result, "")
		} else if option == "update" {
			item, err := GetProfileFolderByIdT(_idt)
			if err != nil {
				return
			}

			_id := item.Key("module_id") + "/" + item.Key("profile_tp") + "/" + item.Key("folder_id")
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

	if err := core.InitModel(ProfileFolders); err != nil {
		return console.Panic(err)
	}

	return nil
}

/**
* Profile
*	Handler for CRUD data
**/
func GetProfileByIdT(_idt string) (et.Item, error) {
	return Profiles.Data().
		Where(Profiles.Column("_idt").Eq(_idt)).
		First()
}

func GetProfileById(moduleId, profileTp string) (et.Item, error) {
	return Profiles.Data().
		Where(Profiles.Column("module_id").Eq(moduleId)).
		And(Profiles.Column("profile_tp").Eq(profileTp)).
		First()
}

func InitProfile(moduleId, profileTp string, data et.Json) (et.Item, error) {
	if !utility.ValidId(moduleId) {
		return et.Item{}, console.AlertF(msg.MSG_ATRIB_REQUIRED, "moduleId")
	}

	if !utility.ValidId(profileTp) {
		return et.Item{}, console.AlertF(msg.MSG_ATRIB_REQUIRED, "profile_tp")
	}

	module, err := GetModuleById(moduleId)
	if err != nil {
		return et.Item{}, err
	}

	if !module.Ok {
		return et.Item{}, console.ErrorM(msg.MODULE_NOT_FOUND)
	}

	current, err := GetProfileById(moduleId, profileTp)
	if err != nil {
		return et.Item{}, err
	}

	if current.Ok {
		return et.Item{
			Ok: current.Ok,
			Result: et.Json{
				"message": msg.RECORD_FOUND,
				"_id":     current.Id(),
				"index":   current.Index(),
			},
		}, nil
	}

	data["module_id"] = moduleId
	data["profile_tp"] = profileTp
	return Profiles.Insert(data).
		CommandOne()
}

func UpSetProfile(moduleId, profileTp string, data et.Json) (et.Item, error) {
	if !utility.ValidId(moduleId) {
		return et.Item{}, console.AlertF(msg.MSG_ATRIB_REQUIRED, "moduleId")
	}

	if !utility.ValidId(profileTp) {
		return et.Item{}, console.AlertF(msg.MSG_ATRIB_REQUIRED, "profile_tp")
	}

	module, err := GetModuleById(moduleId)
	if err != nil {
		return et.Item{}, err
	}

	if !module.Ok {
		return et.Item{}, console.ErrorM(msg.MODULE_NOT_FOUND)
	}

	data["module_id"] = moduleId
	data["profile_tp"] = profileTp
	return Profiles.Upsert(data).
		Where(Profiles.Column("module_id").Eq(moduleId)).
		And(Profiles.Column("profile_tp").Eq(profileTp)).
		CommandOne()
}

func UpSetProfileTp(projectId, moduleId, id, name, description string, data et.Json) (et.Item, error) {
	if !utility.ValidStr(name, 0, []string{""}) {
		return et.Item{}, console.AlertF(msg.MSG_ATRIB_REQUIRED, "name")
	}

	profile, err := UpSetType(projectId, id, "PROFILE", name, description)
	if err != nil {
		return et.Item{}, err
	}

	profileTp := profile.Id()
	_, err = UpSetProfile(moduleId, profileTp, data)
	if err != nil {
		return et.Item{}, err
	}

	profile.Set("project_id", projectId)
	profile.Set("module_id", moduleId)
	return profile, nil
}

func DeleteProfile(moduleId, profileTp string) (et.Item, error) {
	current, err := GetProfileById(moduleId, profileTp)
	if err != nil {
		return et.Item{}, err
	}

	if !current.Ok {
		return et.Item{}, nil
	}

	return Profiles.Delete().
		Where(Profiles.Column("module_id").Eq(moduleId)).
		And(Profiles.Column("profile_tp").Eq(profileTp)).
		CommandOne()
}

/**
* Profile Folder
**/
func GetProfileFolderByIdT(_idt string) (et.Item, error) {
	return ProfileFolders.Data().
		Where(ProfileFolders.Column("_idt").Eq(_idt)).
		First()
}

func GetProfileFolderById(moduleId, profileTp, folderId string) (et.Item, error) {
	return ProfileFolders.Data().
		Where(ProfileFolders.Column("module_id").Eq(moduleId)).
		And(ProfileFolders.Column("profile_tp").Eq(profileTp)).
		And(ProfileFolders.Column("folder_id").Eq(folderId)).
		First()
}

func CheckProfileFolder(moduleId, profileTp, folderId string, chk bool) (et.Item, error) {
	if !utility.ValidId(moduleId) {
		return et.Item{}, console.AlertF(msg.MSG_ATRIB_REQUIRED, "module_id")
	}

	if !utility.ValidId(profileTp) {
		return et.Item{}, console.AlertF(msg.MSG_ATRIB_REQUIRED, "profile_tp")
	}

	if !utility.ValidId(folderId) {
		return et.Item{}, console.AlertF(msg.MSG_ATRIB_REQUIRED, "folder_id")
	}

	profile, err := GetTypeById(profileTp)
	if err != nil {
		return et.Item{}, err
	}

	if !profile.Ok {
		return et.Item{}, console.AlertF(msg.PROFILE_NOT_FOUND, profileTp)
	}

	data := et.Json{}
	data.Set("module_id", moduleId)
	data.Set("profile_tp", profileTp)
	data.Set("folder_id", folderId)
	if chk {
		return ProfileFolders.Insert(data).
			Where(ProfileFolders.Column("module_id").Eq(moduleId)).
			And(ProfileFolders.Column("profile_tp").Eq(profileTp)).
			And(ProfileFolders.Column("folder_id").Eq(folderId)).
			Returns(ProfileFolders.Column("index")).
			CommandOne()
	} else {
		return ProfileFolders.Delete().
			Where(ProfileFolders.Column("module_id").Eq(moduleId)).
			And(ProfileFolders.Column("profile_tp").Eq(profileTp)).
			And(ProfileFolders.Column("folder_id").Eq(folderId)).
			CommandOne()
	}
}

func getProfileFolders(userId, projectId, mainId string) []et.Json {
	sql := `
	SELECT DISTINCT A._DATA||jsonb_build_object('date_make', A.DATE_MAKE,
	'date_update', A.DATE_UPDATE,
	'module_id', A.MODULE_ID,
	'_state', A._STATE,
	'_id', A._ID,
	'main_id', A.MAIN_ID,
	'name', A.NAME,
	'description', A.DESCRIPTION,
	'index', A.INDEX) AS _DATA,
	$2 AS PROJECT_ID,
	A.INDEX
	FROM module.FOLDERS AS A
	INNER JOIN module.PROFILE_FOLDERS AS B ON B.FOLDER_ID = A._ID
	INNER JOIN module.MODULE_FOLDERS AS C ON C.FOLDER_ID = A._ID
	WHERE A.MAIN_ID = $1
	AND C.MODULE_ID IN (SELECT C.MODULE_ID FROM module.PROJECT_MODULES AS C WHERE C.PROJECT_ID = $2)
	AND B.PROFILE_TP IN (SELECT D.PROFILE_TP FROM module.ROLES AS D WHERE D.PROJECT_ID = $2 AND D.USER_ID = $3)
	ORDER BY A.INDEX ASC;`

	items, err := linq.QueryData(sql, mainId, projectId, userId)
	if err != nil {
		return []et.Json{}
	}

	return items.Result
}

func GetProfileFolders(userId, projectId string) ([]et.Json, error) {
	if !utility.ValidId(userId) {
		return []et.Json{}, console.AlertF(msg.MSG_ATRIB_REQUIRED, "userId")
	}

	if !utility.ValidId(projectId) {
		return []et.Json{}, console.AlertF(msg.MSG_ATRIB_REQUIRED, "project_id")
	}

	mainId := "-1"
	result := getProfileFolders(userId, projectId, mainId)
	for _, item := range result {
		mainId = item.Id()
		item["folders"] = getProfileFolders(userId, projectId, mainId)
	}

	return result, nil
}

func AllModuleProfiles(projectId, moduleId, state, search string, page, rows int) (et.List, error) {
	if state == "" {
		state = utility.ACTIVE
	}

	auxState := state

	_select := Profiles.All()
	_select = append(_select, Types.Column("_state"), Types.Column("_id"), Types.Column("name"), Types.Column("description"))

	if search != "" {
		return linq.From(Profiles, "A").
			Join(Profiles.As("A"), Types.As("B"), Types.Col("_id").Eq(Profiles.Col("profile_tp"))).
			Where(Types.Column("project_id").In("-1", projectId)).
			And(Profiles.Column("module_id").Eq(moduleId)).
			And(Profiles.Concat("NAME:", Types.As("B").Col("name"), ":DESCRIPTION:", Types.As("B").Col("description"), ":DATA:", Profiles.As("A").Column("_data"), ":").Like("%"+search+"%")).
			OrderBy(Types.Column("name"), true).
			Data(_select).
			List(page, rows)
	} else if auxState == "*" {
		state = utility.FOR_DELETE

		return linq.From(Profiles, "A").
			Join(Profiles.As("A"), Types.As("B"), Types.Col("_id").Eq(Profiles.Col("profile_tp"))).
			Where(Types.Column("_state").Neg(state)).
			And(Types.Column("project_id").In("-1", projectId)).
			And(Profiles.Column("module_id").Eq(moduleId)).
			OrderBy(Types.Column("name"), true).
			Data(_select).
			List(page, rows)
	} else if auxState == "0" {
		return linq.From(Profiles, "A").
			Join(Profiles.As("A"), Types.As("B"), Types.Col("_id").Eq(Profiles.Col("profile_tp"))).
			Where(Types.Column("_state").In("-1", state)).
			And(Types.Column("project_id").In("-1", projectId)).
			And(Profiles.Column("module_id").Eq(moduleId)).
			OrderBy(Types.Column("name"), true).
			Data(_select).
			List(page, rows)
	} else {
		return linq.From(Profiles, "A").
			Join(Profiles.As("A"), Types.As("B"), Types.Col("_id").Eq(Profiles.Col("profile_tp"))).
			Where(Types.Column("_state").Eq(state)).
			And(Types.Column("project_id").In("-1", projectId)).
			And(Profiles.Column("module_id").Eq(moduleId)).
			OrderBy(Types.Column("name"), true).
			Data(_select).
			List(page, rows)
	}
}
