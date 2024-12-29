package module

import (
	"time"

	"github.com/cgalvisleon/elvis/cache"
	"github.com/cgalvisleon/elvis/claim"
	"github.com/cgalvisleon/elvis/console"
	"github.com/cgalvisleon/elvis/core"
	"github.com/cgalvisleon/elvis/et"
	"github.com/cgalvisleon/elvis/event"
	"github.com/cgalvisleon/elvis/jdb"
	"github.com/cgalvisleon/elvis/linq"
	"github.com/cgalvisleon/elvis/msg"
	"github.com/cgalvisleon/elvis/strs"
	"github.com/cgalvisleon/elvis/utility"
)

type Token struct {
	Date_make   time.Time `json:"date_make"`
	Date_update time.Time `json:"date_update"`
	Id          string    `json:"_id"`
	Name        string    `json:"name"`
	App         string    `json:"app"`
	Device      string    `json:"device"`
	Token       string    `json:"token"`
	Index       int       `json:"index"`
}

func (n *Token) Scan(js *et.Json) error {
	n.Date_make = js.Time("date_make")
	n.Date_update = js.Time("date_update")
	n.Id = js.Str("_id")
	n.Name = js.Str("name")
	n.App = js.Str("app")
	n.Device = js.Str("device")
	n.Token = js.Str("token")
	n.Index = js.Int("index")

	return nil
}

var Tokens *linq.Model

func DefineTokens() error {
	if err := DefineSchemaModule(); err != nil {
		return console.Panic(err)
	}

	if Tokens != nil {
		return nil
	}

	Tokens = linq.NewModel(SchemaModule, "TOKENS", "Tabla de tokens", 1)
	Tokens.DefineColum("date_make", "", "TIMESTAMP", "NOW()")
	Tokens.DefineColum("date_update", "", "TIMESTAMP", "NOW()")
	Tokens.DefineColum("_id", "", "VARCHAR(80)", "-1")
	Tokens.DefineColum("project_id", "", "VARCHAR(80)", "-1")
	Tokens.DefineColum("user_id", "", "VARCHAR(80)", "-1")
	Tokens.DefineColum("name", "", "VARCHAR(80)", "")
	Tokens.DefineColum("app", "", "VARCHAR(80)", "")
	Tokens.DefineColum("device", "", "VARCHAR(80)", "")
	Tokens.DefineColum("token", "", "TEXT", "")
	Tokens.DefineColum("index", "", "INTEGER", 0)
	Tokens.DefinePrimaryKey([]string{"_id"})
	Tokens.DefineForeignKey("project_id", Projects.Col("_id"))
	Tokens.DefineIndex([]string{
		"date_make",
		"date_update",
		"name",
		"app",
		"device",
		"index",
	})
	Tokens.Details("last_use", "", "", func(col *linq.Column, data *et.Json) {
		id := data.Id()
		last_use, err := cache.HGetAtrib(id, "telemetry.token.last_use")
		if err != nil {
			return
		}

		data.Set(col.Low(), last_use)
	})
	Tokens.Details("token", "", "", func(col *linq.Column, data *et.Json) {
		token := data.Str("token")
		newToken := token
		if len(token) > 6 {
			newToken = token[0:6] + "..." + token[len(token)-6:]
		}

		data.Set(col.Low(), newToken)
		data.Set("long_token", token)
	})
	Tokens.OnListener = func(data et.Json) {
		option := data.Str("option")
		_idt := data.Str("_idt")
		if option == "insert" {
			item, err := GetTokenByIdT(_idt)
			if err != nil {
				return
			}

			_id := item.Key("_id")
			event.WsPublish(_id, item.Result, "")
		} else if option == "update" {
			item, err := GetTokenByIdT(_idt)
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

	if err := core.InitModel(Tokens); err != nil {
		return console.Panic(err)
	}

	go LoadTokens()

	return nil
}

func loadToken(token *Token) error {
	key := claim.TokenKey(token.App, token.Device, token.Id)
	err := cache.Set(key, token.Token, 0)
	if err != nil {
		return err
	}

	return nil
}

func unLoadTokenById(app, device, id string) error {
	key := claim.TokenKey(app, device, id)
	_, err := cache.Del(key)
	if err != nil {
		return err
	}

	return nil
}

/**
* Token
*	Handler for CRUD data
**/
func GetTokenByIdT(_idt string) (et.Item, error) {
	return Tokens.Data().
		Where(Tokens.Col("_idt").Eq(_idt)).
		First()
}

func GetTokenById(id string) (et.Item, error) {
	item, err := Tokens.Data().
		Where(Tokens.Col("_id").Eq(id)).
		First()
	if err != nil {
		return et.Item{}, err
	}

	return item, nil
}

func UpSetToken(projeectId, id, app, device, name, userId string) (et.Item, error) {
	user, err := GetUserById(userId)
	if err != nil {
		return et.Item{}, err
	}

	if !user.Ok {
		return et.Item{}, console.ErrorM(msg.USER_NOT_FONUND)
	}

	id = utility.GenId(id)
	current, err := GetTokenById(id)
	if err != nil {
		return et.Item{}, err
	}

	if current.Ok {
		id := current.Id()
		data := et.Json{
			"name": name,
		}

		item, err := Tokens.Update(data).
			Where(Tokens.Col("_id").Eq(id)).
			CommandOne()
		if err != nil {
			return et.Item{}, err
		}

		return et.Item{
			Ok: item.Ok,
			Result: et.OkOrNotJson(item.Ok, item.Result, et.Json{
				"message": msg.RECORD_NOT_UPDATE,
				"_id":     id,
			}),
		}, nil
	} else {
		id := utility.NewId()
		token, err := claim.GenToken(id, app, name, "token", app, device, 0)
		if err != nil {
			return et.Item{}, console.Error(err)
		}

		data := et.Json{}
		data.Set("project_id", projeectId)
		data.Set("_id", id)
		data.Set("user_id", userId)
		data.Set("app", app)
		data.Set("device", device)
		data.Set("name", name)
		data.Set("token", token)

		item, err := Tokens.Insert(data).
			CommandOne()
		if err != nil {
			return et.Item{}, console.Error(err)
		}

		err = loadToken(&Token{
			Date_make:   item.Time("date_make"),
			Date_update: item.Time("date_update"),
			Id:          id,
			Name:        name,
			App:         app,
			Device:      device,
			Token:       token,
			Index:       item.Index(),
		})
		if err != nil {
			return et.Item{}, console.Error(err)
		}

		return et.Item{
			Ok: item.Ok,
			Result: et.OkOrNotJson(item.Ok, item.Result, et.Json{
				"message": msg.RECORD_NOT_CREATE,
				"_id":     id,
			}),
		}, nil
	}
}

func LoadTokens() error {
	var ok bool = true
	var rows int = 30
	var page int = 1
	for ok {
		ok = false

		offset := (page - 1) * rows
		sql := strs.Format(`
		SELECT *
		FROM module.TOKENS
		ORDER BY INDEX
		LIMIT %d OFFSET %d;`, rows, offset)

		items, err := jdb.Query(sql)
		if err != nil {
			return console.Error(err)
		}

		for _, item := range items.Result {
			var result Token
			err = result.Scan(&item)
			if err != nil {
				return console.Error(err)
			}

			err = loadToken(&result)
			if err != nil {
				return console.Error(err)
			}

			ok = true
		}

		page++
	}

	return nil
}

func UnLoadTokens() error {
	var ok bool = true
	var rows int = 30
	var page int = 1
	for ok {
		ok = false

		offset := (page - 1) * rows
		sql := strs.Format(`
		SELECT APP, DEVICE, _ID
		FROM module.TOKENS
		ORDER BY INDEX
		LIMIT %d OFFSET %d;`, rows, offset)

		items, err := jdb.Query(sql)
		if err != nil {
			return console.Error(err)
		}

		for _, item := range items.Result {
			app := item.Str("app")
			device := item.Str("device")
			id := item.Id()
			err = unLoadTokenById(app, device, id)
			if err != nil {
				return console.Error(err)
			}

			ok = true
		}

		page++
	}

	return nil
}

func GetTokensByUserId(userId, search string, page, rows int) (et.List, error) {
	sql := `
  SELECT COUNT(*) AS COUNT
  FROM module.TOKENS A
  WHERE A.USER_ID=$1
	AND CONCAT('NAME:', A.NAME, ':APP:', A.APP, ':DEVICE:', A.DEVICE, ':') ILIKE CONCAT('%', $2, '%');`

	all := jdb.QueryCount(sql, userId, search)

	offset := (page - 1) * rows
	sql = `
  SELECT A.*
  FROM module.TOKENS A
	WHERE A.USER_ID=$1
  AND CONCAT('NAME:', A.NAME, ':APP:', A.APP, ':DEVICE:', A.DEVICE, ':') ILIKE CONCAT('%', $2, '%')
	ORDER BY A.APP, A.DEVICE, A.NAME
  LIMIT $3 OFFSET $4;`

	items, err := jdb.Query(sql, userId, search, rows, offset)
	if err != nil {
		return et.List{}, err
	}

	for _, item := range items.Result {
		id := item.Id()
		last_use, err := cache.HGetAtrib(id, "telemetry.token.last_use")
		if err != nil {
			return et.List{}, err
		}

		token := item["token"].(string)
		newToken := token
		if len(token) > 6 {
			newToken = token[0:6] + "..." + token[len(token)-6:]
		}
		item["token"] = newToken
		item["last_use"] = last_use
	}

	return items.ToList(all, page, rows), nil
}

func DeleteToken(id string) (et.Item, error) {
	current, err := GetTokenById(id)
	if err != nil {
		return et.Item{}, err
	}

	if !current.Ok {
		return et.Item{}, nil
	}

	sql := `
  DELETE FROM module.TOKENS
  WHERE _ID=$1
  RETURNING *;`

	item, err := jdb.QueryOne(sql, id)
	if err != nil {
		return et.Item{}, err
	}

	app := item.Str("app")
	device := item.Str("device")
	err = unLoadTokenById(app, device, id)
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
