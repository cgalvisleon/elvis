package linq

import (
	e "github.com/cgalvisleon/elvis/json"
	"github.com/cgalvisleon/elvis/strs"
	"github.com/cgalvisleon/elvis/utility"
)

/**
*
**/
func (c *Model) Consolidate(current e.Json, linq *Linq) *Linq {
	var col *Column
	var source e.Json = e.Json{}
	var new e.Json = e.Json{}

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
			if col.PrimaryKey || col.ForeignKey {
				linq.references = append(linq.references, &ReferenceValue{c.Schema, c.Table, v, 1})
			}
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

func (c *Model) Changue(current e.Json, linq *Linq) *Linq {
	var change bool
	c.Consolidate(current, linq)
	new := linq.new

	for k, v := range *new {
		k = strs.Lowcase(k)
		idxCol := c.ColIdx(k)

		if idxCol != -1 {
			ch := current.Str(k) != new.Str(k)
			if !change {
				change = ch
			}
			if ch {
				col := c.Column(k)
				if col.PrimaryKey || col.ForeignKey {
					linq.references = append(linq.references, &ReferenceValue{c.Schema, c.Table, current.Str(k), -1})
					linq.references = append(linq.references, &ReferenceValue{c.Schema, c.Table, v, 1})
				}
			}
		}
	}

	linq.change = change

	return linq
}

/**
*	Prepare command data
**/
func (c *Linq) PrepareInsert() error {
	model := c.from[0].model
	model.Consolidate(model.Model(), c)
	for _, validate := range c.validates {
		if err := validate.Col.Valid(validate.Value); err != nil {
			return err
		}
	}

	now := utility.Now()

	if model.UseDateMake {
		c.new.Set(model.DateMakeField, now)
	}

	if model.UseDateUpdate {
		c.new.Set(model.DateUpdateField, now)
	}

	return nil
}

func (c *Linq) PrepareUpdate(current e.Json) bool {
	model := c.from[0].model
	model.Changue(current, c)

	if !c.change {
		return c.change
	}

	if model.UseDateMake {
		delete(*c.new, strs.Lowcase(model.DateMakeField))
	}

	now := utility.Now()
	if model.UseDateUpdate {
		c.new.Set(model.DateUpdateField, now)
	}

	return c.change
}

func (c *Linq) PrepareDelete(current e.Json) {
	model := c.from[0].model

	for k, v := range current {
		col := model.Column(k)
		if col.PrimaryKey || col.ForeignKey {
			c.references = append(c.references, &ReferenceValue{model.Schema, model.Table, v, -1})
		}
	}
}
