package linq

import (
	"errors"

	"github.com/cgalvisleon/elvis/et"
	"github.com/cgalvisleon/elvis/strs"
	"github.com/cgalvisleon/elvis/utility"
)

const TpColumn = 0
const TpAtrib = 1
const TpDetail = 2
const TpReference = 3
const TpCaption = 4
const TpFunction = 5
const TpClone = 6
const TpField = 7

/**
*
**/
type Col struct {
	from string
	name string
	cast string
	as   string
}

func (c *Col) Up() string {
	return strs.Uppcase(c.name)
}

func (c *Col) Low() string {
	return strs.Lowcase(c.name)
}

func (c *Col) Cast(cast string) *Col {
	c.cast = cast

	return c
}

func (c *Col) As() string {
	if len(c.as) == 0 {
		return c.name
	}

	return c.as
}

func (c *Col) AsUp() string {
	return strs.Uppcase(c.As())
}

func (c *Col) AsLow() string {
	return strs.Lowcase(c.As())
}

/**
*
**/
type Details func(col *Column, data *et.Json)

type Column struct {
	Model       *Model
	Tp          int
	Column      *Column
	name        string
	Title       string
	Description string
	Type        string
	Default     any
	Atribs      []*Column
	Reference   *Reference
	Definition  interface{}
	Function    string
	Details     Details
	References  []*Column
	Indexed     bool
	Unique      bool
	Required    bool
	RequiredMsg string
	PrimaryKey  bool
	ForeignKey  bool
	Hidden      bool
	from        string
	cast        string
}

func (c *Column) Driver() string {
	return c.Model.Driver()
}

func (c *Column) describe() et.Json {
	return et.Json{
		"name":        c.name,
		"description": c.Description,
		"type":        c.Type,
		"default":     c.Default,
		"tp":          c.Tp,
		"indexed":     c.Indexed,
		"unique":      c.Unique,
		"required":    c.Required,
		"primaryKey":  c.PrimaryKey,
		"foreignKey":  c.ForeignKey,
		"hidden":      c.Hidden,
		"references":  c.References,
	}
}

func (c *Column) Describe() et.Json {
	var atribs []et.Json = []et.Json{}
	for _, atrib := range c.Atribs {
		atribs = append(atribs, atrib.describe())
	}

	reference := et.Json{}
	if c.Reference != nil {
		reference = c.Reference.Describe()
	}

	return et.Json{
		"name":        c.name,
		"description": c.Description,
		"type":        c.Type,
		"default":     c.Default,
		"atribs":      atribs,
		"reference":   reference,
		"tp":          c.Tp,
		"indexed":     c.Indexed,
		"unique":      c.Unique,
		"required":    c.Required,
		"hidden":      c.Hidden,
		"references":  c.References,
	}
}

func (c *Column) Valid(val any) error {
	if c.Required {
		switch strs.Uppcase(c.Type) {
		case "BOOLEAN":
			if !utility.ValidIn(val.(string), 0, []string{"TRUE", "FALSE", "true", "false", "1", "0"}) {
				return errors.New(c.RequiredMsg)
			}
		default:
			if !utility.ValidStr(val.(string), 0, []string{""}) {
				return errors.New(c.RequiredMsg)
			}
		}
	}

	return nil
}

func NewColumn(model *Model, name, description, _type string, _default any) *Column {
	result := &Column{
		Model:       model,
		Tp:          TpColumn,
		name:        strs.Uppcase(name),
		Description: description,
		Type:        _type,
		Default:     _default,
		Atribs:      []*Column{},
		References:  []*Column{},
		Indexed:     false,
		Hidden:      false,
	}

	if !model.UseDateMake {
		model.UseDateMake = strs.Uppcase(result.name) == strs.Uppcase(model.DateMakeField)
	}

	if !model.UseDateUpdate {
		model.UseDateUpdate = strs.Uppcase(result.name) == strs.Uppcase(model.DateUpdateField)
	}

	if !model.UseState {
		model.UseState = strs.Uppcase(result.name) == strs.Uppcase(model.StateField)
	}

	if !model.UseProject {
		model.UseProject = strs.Uppcase(result.name) == strs.Uppcase(model.ProjectField)
	}

	if !model.UseSerie && model.schema.UseSerie {
		model.UseSerie = strs.Uppcase(result.name) == strs.Uppcase(model.SerieField)
	}

	if !model.UseSource {
		model.UseSource = strs.Uppcase(result.name) == strs.Uppcase(model.SourceField)
	}

	model.Definition = append(model.Definition, result)
	return result
}

func NewVirtualAtrib(model *Model, name, description, _type string, _default any) *Column {
	result := &Column{
		Model:       model,
		Tp:          TpColumn,
		name:        strs.Uppcase(name),
		Description: description,
		Type:        _type,
		Default:     _default,
		Atribs:      []*Column{},
		Indexed:     false,
	}

	return result
}

/**
* DDL
**/
func (c *Column) DDL() string {
	return DDLColumn(c)
}

func (c *Column) DDLIndex() string {
	return DDLIndex(c)
}

func (c *Column) DDLUniqueIndex() string {
	return DDLUniqueIndex(c)
}

/**
* References
**/

func (c *Column) ReferencesIdx(col *Column) int {
	for i, _col := range c.References {
		if _col.name == col.name {
			return i
		}
	}

	return -1
}

func (c *Column) ReferencesAdd(col *Column) int {
	idx := c.ReferencesIdx(col)
	if idx == -1 {
		c.Indexed = true
		c.Model.IndexAdd(c.name)
		c.References = append(c.References, col)
		idx = len(c.References) - 1
	}

	return idx
}

/**
*
**/
func (c *Column) Up() string {
	return strs.Uppcase(c.name)
}

func (c *Column) Low() string {
	return strs.Lowcase(c.name)
}

func (c *Column) ColName() string {
	result := strs.Uppcase(c.name)
	result = strs.Append(c.Model.Name, result, ".")
	result = strs.Append(c.Model.Schema, result, ".")
	return result
}

/**
* This function not use in Select
**/
func (c *Column) As(linq *Linq) string {
	switch c.Tp {
	case TpColumn:
		from := linq.GetFrom(c)
		return strs.Append(from.As(), c.Up(), ".")
	case TpAtrib:
		from := linq.GetFrom(c)
		col := strs.Append(from.As(), c.Column.Up(), ".")
		return strs.Format(`%s#>>'{%s}'`, col, c.Low())
	case TpClone:
		return strs.Append(strs.Uppcase(c.from), c.Up(), ".")
	case TpReference:
		from := linq.GetFrom(c)
		as := linq.GetAs()
		as = strs.Format(`A%s`, as)
		fn := strs.Format(`%s.%s`, as, c.Reference.Reference.Up())
		fm := strs.Format(`%s AS %s`, c.Reference.Reference.Model.Name, as)
		key := strs.Format(`%s.%s`, as, c.Reference.Key)
		Fkey := strs.Append(from.As(), c.Reference.Fkey, ".")
		return strs.Format(`(SELECT %s FROM %s WHERE %s=%v LIMIT 1)`, fn, fm, key, Fkey)
	case TpCaption:
		from := linq.GetFrom(c)
		as := linq.GetAs()
		as = strs.Format(`A%s`, as)
		fn := strs.Format(`%s.%s`, as, c.Reference.Reference.Up())
		fm := strs.Format(`%s AS %s`, c.Reference.Reference.Model.Name, as)
		key := strs.Format(`%s.%s`, as, c.Reference.Key)
		Fkey := strs.Append(from.As(), c.Reference.Fkey, ".")
		return strs.Format(`(SELECT %s FROM %s WHERE %s=%v LIMIT 1)`, fn, fm, key, Fkey)
	case TpDetail:
		return strs.Format(`%v`, et.Unquote(c.Default))
	case TpFunction:
		def := FunctionDef(linq, c)
		return strs.Append(def, c.Up(), " AS ")
	case TpField:
		as := linq.As(c)
		if len(as) > 0 {
			as = strs.Format(`%s.`, as)
		}
		def := strs.Format(`(%v)`, c.Definition)
		return strs.ReplaceAll(def, []string{"{AS}.", "{as}.", "{AS}", "{as}"}, as)
	default:
		return strs.Format(`%s`, c.Up())
	}
}

/**
* This function use in Select
**/
func (c *Column) Def(linq *Linq) string {
	from := linq.GetFrom(c)

	if linq.Tp == TpData {
		switch c.Tp {
		case TpColumn:
			def := c.As(linq)
			return strs.Format(`'%s', %s`, c.Low(), def)
		case TpAtrib:
			def := c.As(linq)
			def = strs.Format(`COALESCE(%s, %v)`, def, et.Unquote(c.Default))
			return strs.Format(`'%s', %s`, c.Low(), def)
		case TpClone:
			return c.As(linq)
		case TpReference:
			Fkey := strs.Append(from.As(), c.Reference.Fkey, ".")
			def := c.As(linq)
			def = strs.Format(`jsonb_build_object('_id', %s, 'name', %s)`, Fkey, def)
			return strs.Format(`'%s', %s`, c.Title, def)
		case TpCaption:
			def := c.As(linq)
			return strs.Format(`'%s', %s`, c.Title, def)
		case TpDetail:
			def := et.Unquote(c.Default)
			return strs.Format(`'%s', %s`, c.Low(), def)
		case TpFunction:
			def := FunctionDef(linq, c)
			return strs.Format(`'%s', %s`, c.Low(), def)
		case TpField:
			def := c.As(linq)
			return strs.Format(`'%s', %s`, c.Low(), def)
		default:
			def := et.Unquote(c.Default)
			return strs.Format(`'%s', %s`, c.Low(), def)
		}
	}

	switch c.Tp {
	case TpColumn:
		return strs.Append(from.As(), c.Up(), ".")
	case TpAtrib:
		col := strs.Append(from.As(), c.Column.Up(), ".")
		def := strs.Format(`%s#>>'{%s}'`, col, c.Low())
		def = strs.Format(`COALESCE(%s, %v)`, def, et.Unquote(c.Default))
		return strs.Format(`%s AS %s`, def, c.Up())
	case TpClone:
		return strs.Append(strs.Uppcase(c.from), c.Up(), ".")
	case TpReference:
		def := c.As(linq)
		return strs.Format(`%s AS %s`, def, strs.Uppcase(c.Title))
	case TpCaption:
		def := c.As(linq)
		return strs.Format(`%s AS %s`, def, strs.Uppcase(c.Title))
	case TpDetail:
		def := et.Unquote(c.Default)
		return strs.Format(`%v AS %s`, def, c.Up())
	case TpFunction:
		def := FunctionDef(linq, c)
		return strs.Format(`%s AS %s`, def, c.Up())
	case TpField:
		def := c.As(linq)
		return strs.Format(`%s AS %s`, def, c.Up())
	default:
		return strs.Format(`%s`, c.Up())
	}
}

/**
* This function use in Select
**/
func (c *Column) From(from string) *Column {
	result := &Column{
		Model:   c.Model,
		Tp:      TpClone,
		name:    c.name,
		Type:    c.Type,
		Default: c.Default,
		Atribs:  c.Atribs,
		from:    from,
	}

	return result
}

func (c *Column) Name(name string) *Column {
	c.name = name

	return c
}

func (c *Column) Cast(cast string) *Column {
	c.cast = cast

	return c
}

/**
*
**/
func (c *Column) Eq(val any) *Where {
	result := NewWhere(c, "=", val)
	result.SetPrimaryKey(c, val)
	return result
}

func (c *Column) Neg(val any) *Where {
	result := NewWhere(c, "!=", val)
	result.SetPrimaryKey(c, val)
	return result
}

func (c *Column) In(vals ...any) *Where {
	result := NewWhere(c, "IN", vals)
	return result
}

func (c *Column) Like(val any) *Where {
	if c.Tp == TpFunction {
		c.Cast("TEXT")
	}

	if val == "%"+"%" {
		val = "%"
	}

	result := NewWhere(c, "ILIKE", val)
	result.SetPrimaryKey(c, val)
	return result
}

func (c *Column) More(val any) *Where {
	result := NewWhere(c, ">", val)
	result.SetPrimaryKey(c, val)
	return result
}

func (c *Column) Less(val any) *Where {
	result := NewWhere(c, "<", val)
	result.SetPrimaryKey(c, val)
	return result
}

func (c *Column) MoreEq(val any) *Where {
	result := NewWhere(c, ">=", val)
	result.SetPrimaryKey(c, val)
	return result
}

func (c *Column) LessEq(val any) *Where {
	result := NewWhere(c, "<=", val)
	result.SetPrimaryKey(c, val)
	return result
}
