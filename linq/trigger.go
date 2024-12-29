package linq

import (
	"github.com/cgalvisleon/elvis/et"
	"github.com/cgalvisleon/elvis/event"

	"github.com/cgalvisleon/elvis/utility"
)

func beforeInsert(model *Model, old, new *et.Json, data et.Json) error {
	now := utility.Now()

	if model.UseDateMake {
		new.Set(model.DateMakeField, now)
	}

	if model.UseDateUpdate {
		new.Set(model.DateUpdateField, now)
	}

	return nil
}

func afterInsert(model *Model, old, new *et.Json, data et.Json) error {
	event.Action("model/insert", et.Json{
		"table": model.Name,
		"old":   old,
		"new":   new,
	})

	return nil
}

func beforeUpdate(model *Model, old, new *et.Json, data et.Json) error {
	now := utility.Now()

	if model.UseDateUpdate {
		new.Set(model.DateUpdateField, now)
	}

	return nil
}

func afterUpdate(model *Model, old, new *et.Json, data et.Json) error {
	event.Action("model/update", et.Json{
		"table": model.Name,
		"old":   old,
		"new":   new,
	})

	return nil
}

func beforeDelete(model *Model, old, new *et.Json, data et.Json) error {
	return nil
}

func afterDelete(model *Model, old, new *et.Json, data et.Json) error {
	event.Action("model/delete", et.Json{
		"table": model.Name,
		"old":   old,
		"new":   new,
	})

	return nil
}
