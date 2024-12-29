package module

import (
	"github.com/cgalvisleon/elvis/aws"
	"github.com/cgalvisleon/elvis/cache"
	"github.com/cgalvisleon/elvis/console"
	"github.com/cgalvisleon/elvis/core"
	"github.com/cgalvisleon/elvis/envar"
	"github.com/cgalvisleon/elvis/et"
	"github.com/cgalvisleon/elvis/event"
	"github.com/cgalvisleon/elvis/linq"
	"github.com/cgalvisleon/elvis/msg"
	"github.com/cgalvisleon/elvis/strs"
	"github.com/cgalvisleon/elvis/utility"
)

var Users *linq.Model

func DefineUsers() error {
	if err := DefineSchemaModule(); err != nil {
		return console.Panic(err)
	}

	if Users != nil {
		return nil
	}

	Users = linq.NewModel(SchemaModule, "USERS", "Tabla de usuarios", 1)
	Users.DefineColum("date_make", "", "TIMESTAMP", "NOW()")
	Users.DefineColum("date_update", "", "TIMESTAMP", "NOW()")
	Users.DefineColum("_state", "", "VARCHAR(80)", utility.ACTIVE)
	Users.DefineColum("_id", "", "VARCHAR(80)", "-1")
	Users.DefineColum("name", "", "VARCHAR(250)", "")
	Users.DefineColum("password", "", "VARCHAR(250)", "")
	Users.DefineColum("_data", "", "JSONB", "{}")
	Users.DefineColum("index", "", "INTEGER", 0)
	Users.DefineAtrib("full_name", "", "text", "")
	Users.DefineAtrib("country", "", "text", "")
	Users.DefineAtrib("phone", "", "text", "")
	Users.DefineAtrib("email", "", "text", "")
	Users.DefineAtrib("avatar", "", "text", "")
	Users.DefinePrimaryKey([]string{"_id"})
	Users.DefineIndex([]string{
		"date_make",
		"date_update",
		"_state",
		"name",
		"index",
	})
	Users.DefineHidden([]string{"password"})
	Users.Details("last_use", "", "", func(col *linq.Column, data *et.Json) {
		id := data.Id()
		last_use, err := cache.HGetAtrib(id, "telemetry.token.last_use")
		if err != nil {
			return
		}

		data.Set(col.Low(), last_use)
	})
	Users.Details("projects", "", []et.Json{}, func(col *linq.Column, data *et.Json) {
		id := data.Id()
		projects, err := GetUserProjects(id)
		if err != nil {
			return
		}

		data.Set(col.Low(), projects)
	})
	Users.Details("modules", "", []et.Json{}, func(col *linq.Column, data *et.Json) {
		id := data.Id()
		modules, err := GetUserModules(id)
		if err != nil {
			return
		}

		data.Set(col.Low(), modules)
	})
	Users.Trigger(linq.AfterInsert, func(model *linq.Model, old, new *et.Json, data et.Json) error {
		id := new.Key("_id")
		if id == "USER.ADMIN" {
			fullName := new.Str("full_name")
			country := new.Str("country")
			phone := new.Str("phone")
			APP := envar.EnvarStr("", "APP")
			message := strs.Format(msg.MSG_ADMIN_WELCOME, fullName, APP)
			go aws.SendSMS(country, phone, message)
		}

		return nil
	})
	Users.OnListener = func(data et.Json) {
		option := data.Str("option")
		_idt := data.Str("_idt")
		if option == "insert" {
			item, err := GetUserByIdT(_idt)
			if err != nil {
				return
			}

			_id := item.Key("_id")
			event.WsPublish(_id, item.Result, "")
		} else if option == "update" {
			item, err := GetUserByIdT(_idt)
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

	if err := core.InitModel(Users); err != nil {
		return console.Panic(err)
	}

	return nil
}

/**
* User
*	Handler for CRUD data
**/
func GetUserByIdT(_idt string) (et.Item, error) {
	return Users.Data().
		Where(Users.Column("_idt").Eq(_idt)).
		First()
}

func GetUserByName(name string) (et.Item, error) {
	item, err := Users.Data().
		Where(Users.Column("name").Eq(name)).
		First()
	if err != nil {
		return et.Item{}, err
	}

	return item, nil
}

func GetUserById(id string) (et.Item, error) {
	item, err := Users.Data().
		Where(Users.Column("_id").Eq(id)).
		First()
	if err != nil {
		return et.Item{}, err
	}

	return item, nil
}

func InitAdmin(fullName, country, phone, email string) (et.Item, error) {
	if !utility.ValidStr(country, 0, []string{""}) {
		return et.Item{}, console.AlertF(msg.MSG_ATRIB_REQUIRED, "country")
	}

	if !utility.ValidStr(phone, 9, []string{""}) {
		return et.Item{}, console.AlertF(msg.MSG_ATRIB_REQUIRED, "phone")
	}

	if !utility.ValidStr(fullName, 0, []string{""}) {
		return et.Item{}, console.AlertF(msg.MSG_ATRIB_REQUIRED, "full_name")
	}

	id := "USER.ADMIN"
	current, err := GetUserById(id)
	if err != nil {
		return et.Item{}, err
	}

	if current.Ok {
		return current, nil
	}

	name := country + phone
	data := et.Json{}
	data["_id"] = id
	data["name"] = name
	data["full_name"] = fullName
	data["country"] = country
	data["phone"] = phone
	data["email"] = email
	data["avatar"] = ""
	item, err := Users.Insert(data).
		Where(Users.Column("_id").Eq(id)).
		CommandOne()
	if err != nil {
		return et.Item{}, err
	}

	return item, nil
}

func UpSetAdmin(fullName, country, phone, email string) (et.Item, error) {
	if !utility.ValidStr(country, 0, []string{""}) {
		return et.Item{}, console.AlertF(msg.MSG_ATRIB_REQUIRED, "country")
	}

	if !utility.ValidStr(phone, 9, []string{""}) {
		return et.Item{}, console.AlertF(msg.MSG_ATRIB_REQUIRED, "phone")
	}

	if !utility.ValidStr(fullName, 0, []string{""}) {
		return et.Item{}, console.AlertF(msg.MSG_ATRIB_REQUIRED, "full_name")
	}

	id := "USER.ADMIN"
	name := country + phone
	data := et.Json{}
	data["_id"] = id
	data["name"] = name
	data["full_name"] = fullName
	data["country"] = country
	data["phone"] = phone
	data["email"] = email
	data["avatar"] = ""
	item, err := Users.Upsert(data).
		Where(Users.Column("_id").Eq(id)).
		CommandOne()
	if err != nil {
		return et.Item{}, err
	}

	return item, nil
}

func InsertUser(fullName, country, phone, email, password string) (et.Item, error) {
	if !utility.ValidStr(country, 0, []string{""}) {
		return et.Item{}, console.AlertF(msg.MSG_ATRIB_REQUIRED, "country")
	}

	if !utility.ValidStr(phone, 9, []string{""}) {
		return et.Item{}, console.AlertF(msg.MSG_ATRIB_REQUIRED, "phone")
	}

	if !utility.ValidStr(fullName, 0, []string{""}) {
		return et.Item{}, console.AlertF(msg.MSG_ATRIB_REQUIRED, "full_name")
	}

	id := utility.NewId()
	data := et.Json{}
	name := country + phone
	data["_id"] = id
	data["_state"] = utility.ACTIVE
	data["name"] = name
	data["full_name"] = fullName
	data["country"] = country
	data["phone"] = phone
	data["email"] = email
	data["password"] = password
	data["avatar"] = ""
	_, err := Users.Insert(data).
		CommandOne()
	if err != nil {
		return et.Item{}, err
	}

	item, err := GetProfile(id)
	if err != nil {
		return et.Item{}, err
	}

	return item, nil
}

func SetUser(fullName, country, phone, email, password string) (et.Item, error) {
	if !utility.ValidStr(country, 0, []string{""}) {
		return et.Item{}, console.AlertF(msg.MSG_ATRIB_REQUIRED, "country")
	}

	if !utility.ValidStr(phone, 9, []string{""}) {
		return et.Item{}, console.AlertF(msg.MSG_ATRIB_REQUIRED, "phone")
	}

	if !utility.ValidStr(fullName, 0, []string{""}) {
		return et.Item{}, console.AlertF(msg.MSG_ATRIB_REQUIRED, "full_name")
	}

	name := country + phone
	current, err := GetUserByName(name)
	if err != nil {
		return et.Item{}, err
	}

	if current.Ok {
		return et.Item{}, console.Alert(msg.RECORD_FOUND)
	}

	result, err := InsertUser(fullName, country, phone, email, password)
	if err != nil {
		return et.Item{}, err
	}

	return result, nil
}

func UpdateUser(id, fullName, phone, email string, data et.Json) (et.Item, error) {
	if !utility.ValidStr(fullName, 3, []string{""}) {
		return et.Item{}, console.AlertF(msg.MSG_ATRIB_REQUIRED, "full_name")
	}

	if !utility.ValidStr(phone, 3, []string{""}) {
		return et.Item{}, console.AlertF(msg.MSG_ATRIB_REQUIRED, "phone")
	}

	current, err := GetUserById(id)
	if err != nil {
		return et.Item{}, err
	}

	if !current.Ok {
		return et.Item{}, console.ErrorM(msg.RECORD_NOT_FOUND)
	}

	name := strs.Format(`+57%s`, phone)
	data["_id"] = id
	data["full_name"] = fullName
	data["name"] = name
	data["phone"] = phone
	data["email"] = email
	data["avatar"] = ""
	_, err = Users.Insert(data).
		Where(Users.Column("_id").Eq(id)).
		And(Users.Column("_state").Eq(utility.ACTIVE)).
		CommandOne()
	if err != nil {
		return et.Item{}, err
	}

	item, err := GetProfile(id)
	if err != nil {
		return et.Item{}, err
	}

	return item, nil
}

func StateUser(id, state string) (et.Item, error) {
	if !utility.ValidId(state) {
		return et.Item{}, console.AlertF(msg.MSG_ATRIB_REQUIRED, "state")
	}

	return Users.Update(et.Json{
		"_state": state,
	}).
		Where(Users.Column("_id").Eq(id)).
		And(Users.Column("_state").Neg(state)).
		CommandOne()
}

func DeleteUser(id string) (et.Item, error) {
	return StateUser(id, utility.FOR_DELETE)
}

func AllUsers(state, search string, page, rows int, _select string) (et.List, error) {
	if state == "" {
		state = utility.ACTIVE
	}

	auxState := state

	if search != "" {
		return Users.Data(_select).
			Where(Users.Concat("NAME:", Users.Column("name"), ":DATA:", Users.Column("_data"), ":").Like("%"+search+"%")).
			OrderBy(Users.Column("name"), true).
			List(page, rows)
	} else if auxState == "*" {
		state = utility.FOR_DELETE

		return Users.Data(_select).
			Where(Users.Column("_state").Neg(state)).
			OrderBy(Users.Column("name"), true).
			List(page, rows)
	} else {
		return Users.Data(_select).
			Where(Users.Column("_state").Eq(state)).
			OrderBy(Users.Column("name"), true).
			List(page, rows)
	}
}

func GetProfile(userId string) (et.Item, error) {
	item, err := Users.Data().
		Where(Users.Column("_id").Eq(userId)).
		First()
	if err != nil {
		return et.Item{}, err
	}

	return item, nil
}
