package module

import (
	"github.com/cgalvisleon/elvis/console"
	"github.com/cgalvisleon/elvis/et"
	"github.com/cgalvisleon/elvis/utility"
)

var initDefine bool

func InitDefine() error {
	if initDefine {
		return nil
	}

	if err := DefineUsers(); err != nil {
		return console.Panic(err)
	}
	if err := DefineProjects(); err != nil {
		return console.Panic(err)
	}
	if err := DefineTypes(); err != nil {
		return console.Panic(err)
	}
	if err := DefineTokens(); err != nil {
		return console.Panic(err)
	}
	if err := DefineModules(); err != nil {
		return console.Panic(err)
	}
	if err := DefineFolders(); err != nil {
		return console.Panic(err)
	}
	if err := DefineProfiles(); err != nil {
		return console.Panic(err)
	}
	if err := DefineRoles(); err != nil {
		return console.Panic(err)
	}
	if err := DefineModuleFolders(); err != nil {
		return console.Panic(err)
	}
	if err := DefineProjectModules(); err != nil {
		return console.Panic(err)
	}
	if err := DefineProfileFolders(); err != nil {
		return console.Panic(err)
	}
	if err := initData(); err != nil {
		return console.Panic(err)
	}

	console.LogK("Module", "Init module")

	initDefine = true

	return nil
}

func initData() error {
	if _, err := Projects.Upsert(et.Json{
		"_id":  "-1",
		"name": "My project",
	}).
		Where(Projects.Column("_id").Eq("-1")).
		CommandOne(); err != nil {
		return err
	}

	if _, err := Modules.Upsert(et.Json{
		"_id":  "-1",
		"name": "Admin",
	}).
		Where(Modules.Column("_id").Eq("-1")).
		CommandOne(); err != nil {
		return err
	}

	if _, err := Types.Upsert(et.Json{
		"_id": "-1",
	}).
		Where(Types.Column("_id").Eq("-1")).
		CommandOne(); err != nil {
		return err
	}

	// Initial state types
	InitType("-1", utility.OF_SYSTEM, utility.OF_SYSTEM, "STATE", "System", "Record system")
	InitType("-1", utility.FOR_DELETE, utility.OF_SYSTEM, "STATE", "Delete", "To delete record")
	InitType("-1", utility.ACTIVE, utility.OF_SYSTEM, "STATE", "Activo", "")
	InitType("-1", utility.ARCHIVED, utility.OF_SYSTEM, "STATE", "Archivado", "")
	InitType("-1", utility.CANCELLED, utility.OF_SYSTEM, "STATE", "Cacnelado", "")
	InitType("-1", utility.IN_PROCESS, utility.OF_SYSTEM, "STATE", "En tramite", "")
	InitType("-1", utility.PENDING_APPROVAL, utility.OF_SYSTEM, "STATE", "Pendiente de aprobación", "")
	InitType("-1", utility.APPROVAL, utility.OF_SYSTEM, "STATE", "Aprobado", "")
	InitType("-1", utility.REFUSED, utility.OF_SYSTEM, "STATE", "Rechazado", "")
	// Initial profile types
	InitType("-1", "PROFILE.ADMIN", utility.OF_SYSTEM, "PROFILE", "Admin", "")
	InitType("-1", "PROFILE.DEV", utility.OF_SYSTEM, "PROFILE", "Develop", "")
	InitType("-1", "PROFILE.SUPORT", utility.OF_SYSTEM, "PROFILE", "Suport", "")

	if _, err := Folders.Upsert(et.Json{
		"_id": "-1",
	}).
		Where(Folders.Column("_id").Eq("-1")).
		CommandOne(); err != nil {
		return err
	}

	CheckProjectModule("-1", "-1", true)
	CheckRole("-1", "-1", "PROFILE.ADMIN", "USER.ADMIN", true)

	return nil
}
