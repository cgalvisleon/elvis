package linq

import (
	"github.com/cgalvisleon/elvis/et"
	"github.com/cgalvisleon/elvis/jdb"
	"github.com/cgalvisleon/elvis/strs"
)

/**
*
**/

func Query(sql string, args ...any) (et.Items, error) {
	result, err := jdb.DBQuery(0, sql, args...)
	if err != nil {
		return et.Items{}, err
	}

	return result, nil
}

func QueryOne(sql string, args ...any) (et.Item, error) {
	result, err := jdb.DBQueryOne(0, sql, args...)
	if err != nil {
		return et.Item{}, err
	}

	return result, nil
}

func QueryData(sql string, args ...any) (et.Items, error) {
	result, err := jdb.DBQueryData(0, sql, args...)
	if err != nil {
		return et.Items{}, err
	}

	return result, nil
}

func QueryDataOne(sql string, args ...any) (et.Item, error) {
	result, err := jdb.DBQueryDataOne(0, sql, args...)
	if err != nil {
		return et.Item{}, err
	}

	return result, nil
}

func From(model *Model, as ...string) *Linq {
	result := NewLinq(ActSelect, model, as...)

	return result
}

func (c *Linq) From(model *Model, as ...string) *Linq {
	if len(as) == 0 {
		as = []string{""}
	}
	from := &FRom{model: model, as: strs.Uppcase(as[0])}
	c.from = append(c.from, from)
	c.fromAs = append(c.fromAs, from)

	return c
}

func (c *Linq) Where(where *Where) *Linq {
	where.linq = c
	c.where = append(c.where, where)

	return c
}

func (c *Linq) And(where *Where) *Linq {
	where.linq = c
	where.connector = "AND"
	c.where = append(c.where, where)

	return c
}

func (c *Linq) Or(where *Where) *Linq {
	where.linq = c
	where.connector = "OR"
	c.where = append(c.where, where)

	return c
}

func (c *Linq) OrderBy(col *Column, sorted bool) *Linq {
	c.orderBy = append(c.orderBy, &OrderBy{colum: col, sorted: sorted})

	return c
}

func (c *Linq) GroupBy(cols ...any) *Linq {
	c.groupBy = c.ToCols(cols...)

	return c
}

func (c *Linq) Returns(cols ...any) *Linq {
	c._return = c.ToCols(cols...)

	return c
}
