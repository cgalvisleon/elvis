package linq

import (
	"strings"

	"github.com/cgalvisleon/elvis/console"
	"github.com/cgalvisleon/elvis/et"
	"github.com/cgalvisleon/elvis/jdb"
	"github.com/cgalvisleon/elvis/strs"
)

const TpRow = 1
const TpData = 2

const ActSelect = 3
const ActInsert = 4
const ActDelete = 5
const ActUpdate = 6
const ActUpsert = 7

/**
*
**/
type SQL struct {
	val string
}

/**
*
**/
type FRom struct {
	model *Model
	as    string
}

func (c *FRom) As() string {
	return c.as
}

func (c *FRom) NameAs() string {
	return strs.Append(c.model.Name, c.as, " AS ")
}

func (c *FRom) Col(name string, cast ...string) *Col {
	_cast := ""
	if len(cast) > 0 {
		_cast = cast[0]
	}

	return &Col{
		from: c.as,
		name: name,
		cast: _cast,
	}
}

func (c *FRom) Column(name string, cast ...string) *Col {
	return c.Col(name, cast...)
}

/**
*
**/
type Join struct {
	kind  string
	from  *FRom
	join  *FRom
	where *Where
}

/**
*
**/
type OrderBy struct {
	colum  *Column
	sorted bool
}

/**
*
**/
type Validate struct {
	Col   *Column
	Value any
}

/**
*
**/
type Linq struct {
	Tp        int
	Act       int
	db        int
	_select   []*Column
	from      []*FRom
	where     []*Where
	_join     []*Join
	orderBy   []*OrderBy
	groupBy   []*Column
	_return   []*Column
	fromAs    []*FRom
	as        int
	details   []*Column
	validates []*Validate
	keys      []*Validate
	data      et.Json
	new       *et.Json
	change    bool
	debug     int
	sql       string
}

/**
*
**/
func NewLinq(act int, model *Model, as ...string) *Linq {
	if len(as) == 0 && act == ActSelect {
		as = []string{GetAs(0)}
	} else if len(as) == 0 {
		as = []string{""}
	}
	from := &FRom{model: model, as: strs.Uppcase(as[0])}
	return &Linq{
		Tp:        TpRow,
		Act:       act,
		db:        model.Db,
		from:      []*FRom{from},
		fromAs:    []*FRom{from},
		where:     []*Where{},
		_join:     []*Join{},
		orderBy:   []*OrderBy{},
		groupBy:   []*Column{},
		details:   []*Column{},
		validates: []*Validate{},
		keys:      []*Validate{},
		data:      et.Json{},
		new:       &et.Json{},
		as:        1,
	}
}

/**
*
**/
func GetAs(n int) string {
	limit := 18251
	base := 26
	as := ""
	a := n % base
	b := n / base
	c := b / base

	if n >= limit {
		n = n - limit + 702
		a = n % base
		b = n / base
		c = b / base
		b = b / base
		a = 65 + a
		b = 65 + b - 1
		c = 65 + c - 1
		as = strs.Format(`A%c%c%c`, rune(c), rune(b), rune(a))
	} else if b > base {
		b = b / base
		a = 65 + a
		b = 65 + b - 1
		c = 65 + c - 1
		as = strs.Format(`%c%c%c`, rune(c), rune(b), rune(a))
	} else if b > 0 {
		a = 65 + a
		b = 65 + b - 1
		as = strs.Format(`%c%c`, rune(b), rune(a))
	} else {
		a = 65 + a
		as = strs.Format(`%c`, rune(a))
	}

	return as
}

/**
*
**/
func (c *Linq) SetTp(tp int) *Linq {
	c.Tp = tp

	return c
}

func (c *Linq) GetAs() string {
	result := GetAs(c.as)
	c.as++

	return result
}

func (c *Linq) Details(data *et.Json) *et.Json {
	for _, col := range c.details {
		col.Details(col, data)
	}

	return data
}

func (c *Linq) GetFrom(col *Column) *FRom {
	model := col.Model
	for _, item := range c.fromAs {
		if item.model.Up() == model.Up() {
			return item
		}
	}

	as := c.GetAs()
	result := &FRom{
		model: model,
		as:    as,
	}
	c.fromAs = append(c.fromAs, result)

	return result
}

func (c *Linq) SetFromAs(from *FRom) *FRom {
	model := from.model
	for _, item := range c.fromAs {
		if item.model.Up() == model.Up() {
			item.as = from.as
			return item
		}
	}

	as := c.GetAs()
	result := &FRom{
		model: model,
		as:    as,
	}
	c.fromAs = append(c.fromAs, result)

	return result
}

func (c *Linq) SetAs(model *Model, as string) string {
	for _, item := range c.fromAs {
		if item.model.Up() == model.Up() {
			item.as = as
			return as
		}
	}

	result := &FRom{
		model: model,
		as:    as,
	}
	c.fromAs = append(c.fromAs, result)

	return as
}

func (c *Linq) As(val any) string {
	switch v := val.(type) {
	case Column:
		col := &v
		return c.GetFrom(col).as
	case *Column:
		col := v
		return c.GetFrom(col).as
	case Col:
		col := &v
		return col.from
	case *Col:
		col := v
		return col.from
	default:
		return c.GetAs()
	}
}

func (c *Linq) Col(val any) *Column {
	switch v := val.(type) {
	case Column:
		return &v
	case *Column:
		return v
	default:
		return &Column{}
	}
}

func (c *Linq) GetCol(name string) *Column {
	if f := strs.ReplaceAll(name, []string{" "}, ""); len(f) == 0 {
		return nil
	}

	pars := strings.Split(name, ".")

	if len(pars) == 2 {
		modelN := pars[0]
		colN := pars[0]
		for _, item := range c.fromAs {
			if item.model.Up() == strs.Uppcase(modelN) {
				return item.model.Column(colN)
			}
		}
	} else if len(pars) == 1 {
		colN := pars[0]
		if len(c.fromAs) == 0 {
			return nil
		}

		return c.fromAs[0].model.Column(colN)
	}

	return nil
}

func (c *Linq) ToCols(sel ...any) []*Column {
	var cols []*Column
	for _, col := range sel {
		switch v := col.(type) {
		case Column:
			cols = append(cols, &v)
		case *Column:
			cols = append(cols, v)
		case string:
			c := c.GetCol(v)
			if c != nil {
				cols = append(cols, c)
			}
		}
	}

	return cols
}

func (c *Linq) AddPrimaryKey(col *Column, val any) {
	if !col.Required {
		return
	}

	ok := false
	if col.PrimaryKey {
		ok = false
		for _, key := range c.keys {
			if key.Col.ColName() == col.ColName() {
				ok = true
				break
			}
		}

		if !ok {
			c.keys = append(c.keys, &Validate{
				Col:   col,
				Value: val,
			})
		}
	}
}

func (c *Linq) AddValidate(col *Column, val any) {
	if !col.Required {
		return
	}

	ok := false
	for _, validate := range c.validates {
		if validate.Col.ColName() == col.ColName() {
			ok = true
			break
		}
	}

	if !ok {
		c.validates = append(c.validates, &Validate{
			Col:   col,
			Value: val,
		})
	}

	c.AddPrimaryKey(col, val)
}

/**
* Query
**/
func (c *Linq) Query() (et.Items, error) {
	if c.debug == 1 {
		console.Log(c.sql)
	}

	if c.Tp == TpData {
		result, err := jdb.DBQueryData(c.db, c.sql)
		if err != nil {
			return et.Items{}, err
		}

		return result, nil
	}

	result, err := jdb.DBQuery(c.db, c.sql)
	if err != nil {
		return et.Items{}, err
	}

	return result, nil
}

func (c *Linq) QueryOne() (et.Item, error) {
	if c.debug == 1 {
		console.Log(c.sql)
	}

	if c.Tp == TpData {
		result, err := jdb.DBQueryDataOne(c.db, c.sql)
		if err != nil {
			return et.Item{}, err
		}

		return result, nil
	}

	result, err := jdb.DBQueryOne(c.db, c.sql)
	if err != nil {
		return et.Item{}, err
	}

	return result, nil
}

func (c *Linq) QueryCount() int {
	if c.debug == 1 {
		console.Log(c.sql)
	}

	result := jdb.DBQueryCount(c.db, c.sql)

	return result
}
