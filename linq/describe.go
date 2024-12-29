package linq

import (
	"github.com/cgalvisleon/elvis/et"
	"github.com/cgalvisleon/elvis/jdb"
)

func Describe(db int, schema, model, filter string) et.Json {
	if len(model) > 0 {
		_schema := GetSchema(schema)
		if _schema == nil {
			return et.Json{}
		}

		_model := _schema.Model(model)
		if _model == nil {
			return et.Json{}
		}

		result := _model.Describe()

		if len(filter) > 0 {
			return et.Json{
				filter: result.Get(filter),
			}
		}

		return result
	}

	if len(schema) > 0 {
		_schema := GetSchema(schema)
		if _schema == nil {
			return et.Json{}
		}

		return _schema.Describe()
	}

	var describes []et.Json = []et.Json{}
	for _, schema := range schemas {
		describes = append(describes, schema.Describe())
	}

	result := jdb.DB(db).Describe()
	result.Set("schemas", describes)

	return result
}
