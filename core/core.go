package core

import "github.com/cgalvisleon/elvis/linq"

func InitModel(model *linq.Model) error {
	if err := defineSchemaCore(); err != nil {
		return err
	}

	err := model.Init()
	if err != nil {
		return err
	}

	err = SetStruct("TABLE", model.Schema, model.Table, model.Define)
	if err != nil {
		return err
	}

	if model.UseSync() {
		SetSyncTrigger(model)
	} else {
		SetListenTrigger(model)
	}

	if model.UseRecycle() && model.UseState {
		SetRecycligTrigger(model)
	}

	if model.UseSerie {
		DefineSerie(model)
	}

	return nil
}
