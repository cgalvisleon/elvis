package linq

import (
	"github.com/cgalvisleon/elvis/et"
	"github.com/cgalvisleon/elvis/strs"
	"github.com/cgalvisleon/elvis/utility"
)

/**
*
**/
func (c *Model) Consolidate(linq *Linq) *Linq {
	var col *Column
	var source et.Json = et.Json{}
	var new et.Json = et.Json{}

	setValue := func(key string, val interface{}) {
		new.Set(key, val)
	}

	for k, v := range linq.data {
		k = strs.Lowcase(k)
		idxCol := c.ColIdx(k)

		if idxCol == -1 {
			idx := c.TitleIdx(k)
			if idx != -1 && utility.ContainsInt([]int{TpReference}, c.Definition[idx].Tp) {
				col = c.Definition[idx]
				linq.AddValidate(col, v)
				reference := linq.data.Json(k)
				setValue(col.name, reference.Key(col.Reference.Key))
				continue
			}
		}

		if idxCol == -1 && !c.integrityAtrib {
			source.Set(k, v)
			continue
		} else if idxCol == -1 {
			continue
		} else {
			col = c.Definition[idxCol]
			linq.AddValidate(col, v)
		}

		if utility.ContainsInt([]int{TpField, TpFunction, TpDetail}, col.Tp) {
			continue
		} else if k == strs.Lowcase(c.SourceField) {
			atribs := linq.data.Json(k)

			if c.integrityAtrib {
				for ak, av := range atribs {
					ak = strs.Lowcase(ak)
					if idx := c.AtribIdx(ak); idx != -1 {
						atrib := c.Definition[idx]
						linq.AddValidate(atrib, av)
						source[ak] = av
					}
				}
			} else {
				source = atribs
			}
		} else if utility.ContainsInt([]int{TpColumn}, col.Tp) {
			delete(source, k)
			setValue(k, v)
			col := c.Column(k)
			linq.AddValidate(col, v)
		} else if utility.ContainsInt([]int{TpAtrib}, col.Tp) {
			source.Set(k, v)
		}
	}

	if c.UseSource && len(source) > 0 {
		setValue(c.SourceField, source)
	}

	linq.new = &new

	return linq
}

func (c *Model) Changue(current et.Json, linq *Linq) *Linq {
	var change bool
	new := linq.new

	for k := range *new {
		k = strs.Lowcase(k)
		idxCol := c.ColIdx(k)

		if idxCol != -1 {
			ch := current.Str(k) != new.Str(k)
			if !change {
				change = ch
			}
		}
	}

	linq.change = change

	return linq
}

/**
*	Prepare command data
**/
func (c *Linq) PrepareInsert() (et.Items, error) {
	model := c.from[0].model
	model.Consolidate(c)
	for _, validate := range c.validates {
		if err := validate.Col.Valid(validate.Value); err != nil {
			return et.Items{}, err
		}
	}

	result, err := c.Current()
	if err != nil {
		return et.Items{}, err
	}

	return result, nil
}

func (c *Linq) PrepareUpdate() (et.Items, error) {
	model := c.from[0].model
	model.Consolidate(c)

	result, err := c.Current()
	if err != nil {
		return et.Items{}, err
	}

	if !result.Ok {
		return et.Items{}, nil
	}

	return result, nil
}

func (c *Linq) PrepareDelete() (et.Items, error) {
	return c.PrepareUpdate()
}

func (c *Linq) PrepareUpsert() (et.Items, error) {
	model := c.from[0].model
	model.Consolidate(c)

	current, err := c.Current()
	if err != nil {
		return et.Items{}, err
	}

	return current, nil
}
