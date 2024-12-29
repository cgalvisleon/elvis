package linq

import (
	"strings"

	"github.com/cgalvisleon/elvis/et"
	"github.com/cgalvisleon/elvis/msg"
	"github.com/cgalvisleon/elvis/strs"
)

func (c *Model) DefineColum(name, description, _type string, _default any) *Model {
	NewColumn(c, name, description, _type, _default)

	return c
}

func (c *Model) DefineAtrib(name, description, _type string, _default any) *Model {
	source := c.Col(c.SourceField)
	result := NewColumn(c, name, description, _type, _default)
	result.Tp = TpAtrib
	result.Column = source
	result.name = strs.Lowcase(name)
	source.Atribs = append(source.Atribs, result)

	return c
}

func (c *Model) DefineIndex(index []string) *Model {
	for _, name := range index {
		idx := c.ColIdx(name)
		if idx != -1 {
			c.Definition[idx].Indexed = true
			c.IndexAdd(name)
		}
	}

	return c
}

func (c *Model) DefineUniqueIndex(index []string) *Model {
	for _, name := range c.Index {
		col := c.Col(name)
		if col != nil {
			col.Indexed = true
			col.Unique = true
			c.IndexAdd(name)
		}
	}

	return c
}

func (c *Model) DefineHidden(hiddens []string) *Model {
	for _, key := range hiddens {
		col := c.Col(key)
		if col != nil {
			col.Hidden = true
		}
	}

	return c
}

func (c *Model) DefinePrimaryKey(keys []string) *Model {
	for _, name := range keys {
		col := c.Col(name)
		if col != nil {
			col.Unique = true
			col.Required = true
			col.PrimaryKey = true
			c.PrimaryKeys = append(c.PrimaryKeys, name)
		}
	}

	return c
}

func (c *Model) DefineForeignKey(thisKey string, otherKey *Column) *Model {
	col := c.Col(thisKey)
	if col != nil {
		col.ForeignKey = true
		col.Reference = NewForeignKey(thisKey, otherKey)
		c.ForeignKey = append(c.ForeignKey, col.Reference)
		c.IndexAdd(thisKey)
		otherKey.ReferencesAdd(col)
	}

	return c
}

func (c *Model) DefineReference(thisKey, name, otherKey string, column *Column, showThisKey bool) *Model {
	if name == "" {
		name = thisKey
	}
	idxName := c.ColIdx(name)
	if idxName == -1 {
		col := NewColumn(c, name, "", "REFERENCE", et.Json{"_id": "", "name": ""})
		col.Tp = TpReference
		col.Title = name
		col.Reference = &Reference{thisKey, name, otherKey, column}
		idxThisKey := c.ColIdx(thisKey)
		if idxThisKey != -1 {
			c.Definition[idxThisKey].Hidden = !showThisKey
			c.Definition[idxThisKey].Indexed = true
			c.Definition[idxThisKey].Model.IndexAdd(c.Definition[idxThisKey].name)
			_otherKey := column.Model.Col(otherKey)
			if _otherKey != nil {
				_otherKey.ReferencesAdd(c.Definition[idxThisKey])
			}
		}
	}

	return c
}

func (c *Model) DefineCaption(thisKey, name, otherKey string, column *Column, _default any) *Model {
	if name == "" {
		name = thisKey
	}
	idx := c.ColIdx(name)
	if idx == -1 {
		col := NewColumn(c, name, "", "CAPTION", _default)
		col.Tp = TpCaption
		col.Title = name
		col.Reference = &Reference{thisKey, name, otherKey, column}
		idx := c.ColIdx(thisKey)
		if idx != -1 {
			c.Definition[idx].Indexed = true
			c.Definition[idx].Model.IndexAdd(c.Definition[idx].name)
			_otherKey := column.Model.Col(otherKey)
			if _otherKey != nil {
				_otherKey.ReferencesAdd(c.Definition[idx])
			}
		}
	}

	return c
}

func (c *Model) DefineField(name, description string, _default any, definition string) *Model {
	result := NewColumn(c, name, "", "FIELD", _default)
	result.Tp = TpField
	result.Definition = definition

	return c
}

func (c *Model) DefineRequired(names []string) *Model {
	for _, name := range names {
		list := strings.Split(name, ":")
		key := list[0]
		col := c.Col(key)
		if col != nil {
			col.Required = true
		}

		if len(list) > 1 {
			msg := list[1]
			if msg == "" {
				col.RequiredMsg = msg
			}
		} else {
			col.RequiredMsg = strs.Format(msg.MSG_ATRIB_REQUIRED, col.name)
		}
	}

	return c
}
