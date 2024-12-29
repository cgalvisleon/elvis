package linq

import (
	"github.com/cgalvisleon/elvis/et"
	"github.com/cgalvisleon/elvis/strs"
)

func (c *Linq) Sql() string {
	if c.Act == ActInsert {
		c.PrepareInsert()
		return c.SqlInsert()
	} else if c.Act == ActUpdate {
		c.PrepareUpdate()
		return c.SqlUpdate()
	} else if c.Act == ActDelete {
		c.PrepareDelete()
		return c.SqlDelete()
	} else {
		return c.SqlSelect()
	}
}

func (c *Linq) SQL() SQL {
	c.Sql()
	return SQL{
		val: strs.Format(`(%s)`, c.sql),
	}
}

/**
*
**/
func (c *Linq) SqlColumDef(cols ...*Column) string {
	var result string

	if c.Tp == TpData {
		atribs := make(map[string]string)
		data := ""
		objects := ""
		json := ""
		n := 0

		for _, col := range cols {
			n++
			def := col.Def(c)

			if col.name == col.Model.SourceField {
				data = col.As(c)
			} else if _, exist := atribs[col.name]; !exist {
				atribs[col.name] = col.name
				objects = strs.Append(objects, def, ",\n")
			}

			if n == 20 {
				def := strs.Format(`jsonb_build_object(%s)`, objects)
				json = strs.Append(json, def, "||")
				data = strs.Append(data, json, "||")
				objects = ""
				n = 0
			}
		}

		if n > 0 {
			def := strs.Format(`jsonb_build_object(%s)`, objects)
			json = strs.Append(json, def, "||")
			data = strs.Append(data, json, "||")
		}

		return data
	}

	for _, col := range cols {
		def := col.Def(c)

		result = strs.Append(result, def, ",\n")
	}

	return result
}

func (c *Linq) SqlColums(cols ...*Column) string {
	var result string
	n := len(cols)

	if c.Tp == TpData {
		if n == 0 {
			for _, from := range c.from {
				for _, col := range from.model.Definition {
					if col.Tp != TpAtrib && !col.Hidden {
						cols = append(cols, col)
					}
					if col.Tp == TpDetail {
						c.details = append(c.details, col)
					}
				}

				res := c.SqlColumDef(cols...)
				result = strs.Append(result, res, ",")
			}

			result = strs.Format(`%s AS %s`, result, c.from[0].model.SourceField)
		} else {
			for _, col := range cols {
				if col.Tp == TpDetail {
					c.details = append(c.details, col)
				}
			}

			result = c.SqlColumDef(cols...)
			result = strs.Format(`%s AS %s`, result, c.from[0].model.SourceField)
		}
	} else if n > 0 {
		for _, col := range cols {
			if col.Tp == TpDetail {
				c.details = append(c.details, col)
			}
		}

		result = c.SqlColumDef(cols...)
	} else {
		for _, from := range c.from {
			for _, col := range from.model.Definition {
				if col.Tp == TpDetail {
					c.details = append(c.details, col)
				}
			}
		}

		result = "*"
	}

	return result
}

/**
*
**/
func (c *Linq) SqlSelect() string {
	result := c.SqlColums(c._select...)

	c.sql = strs.Format(`SELECT %s`, result)

	c.SqlFrom()

	c.SqlJoin()

	c.SqlWhere()

	c.SqlGroupBy()

	c.SqlOrderBy()

	return c.sql
}

func (c *Linq) SqlReturn() string {
	result := c.SqlColums(c._return...)

	if len(result) > 0 {
		result = strs.Format(`RETURNING %s`, result)
	}

	c.sql = strs.Append(c.sql, result, "\n")

	return result
}

func (c *Linq) SqlCurrent() string {
	var result string
	var cols []*Column
	model := c.from[0].model

	for _, col := range model.Definition {
		if col.Tp == TpColumn {
			cols = append(cols, col)
		}
	}

	n := len(cols)

	if n > 0 {
		result = c.SqlColumDef(cols...)
		if c.Tp == TpData {
			result = strs.Format(`%s AS %s`, result, c.from[0].model.SourceField)
		}
	} else {
		result = "*"
	}

	c.sql = strs.Format(`SELECT %s`, result)

	c.SqlFrom()

	c.SqlKeys()

	return c.sql
}

func (c *Linq) SqlCount() string {
	c.sql = "SELECT COUNT(*) AS COUNT"

	c.SqlFrom()

	c.SqlJoin()

	c.SqlWhere()

	c.SqlGroupBy()

	return c.sql
}

func (c *Linq) SqlFrom() string {
	var result string
	for _, from := range c.from {
		result = strs.Append(result, from.NameAs(), ", ")
	}

	result = strs.Format(`FROM %s`, result)

	c.sql = strs.Append(c.sql, result, "\n")

	return result
}

func (c *Linq) SqlJoin() string {
	var result string
	for _, join := range c._join {
		where := join.where.Define(c).where
		def := strs.Append(join.join.model.Name, join.join.as, " AS ")
		def = strs.Format(`%s %s ON %s`, join.kind, def, where)
		result = strs.Append(result, def, "\n")
	}

	c.sql = strs.Append(c.sql, result, "\n")

	return result
}

func (c *Linq) SqlWhere() string {
	var result string
	var wh string
	for _, where := range c.where {
		def := where.Define(c)
		if len(result) == 0 {
			wh = def.where
		} else {
			wh = strs.Append(def.connector, def.where, " ")
		}
		result = strs.Append(result, wh, "\n")
	}

	if len(result) > 0 {
		result = strs.Format(`WHERE %s`, result)
	}

	c.sql = strs.Append(c.sql, result, "\n")

	return result
}

func (c *Linq) SqlGroupBy() string {
	var result string
	for _, col := range c.groupBy {
		def := col.As(c)
		result = strs.Append(result, def, ", ")
	}

	if len(result) > 0 {
		result = strs.Format(`GROUP BY %s`, result)
	}

	c.sql = strs.Append(c.sql, result, "\n")

	return result
}

func (c *Linq) SqlOrderBy() string {
	var result string
	var group string
	for _, order := range c.orderBy {
		if order.sorted {
			group = strs.Format(`%s ASC`, order.colum.As(c))
		} else {
			group = strs.Format(`%s DESC`, order.colum.As(c))
		}

		result = strs.Append(result, group, ", ")
	}

	if len(result) > 0 {
		result = strs.Format(`ORDER BY %s`, result)
	}

	c.sql = strs.Append(c.sql, result, "\n")

	return result
}

func (c *Linq) SqlLimit(limit int) string {
	c.SqlSelect()

	result := strs.Format(`LIMIT %d;`, limit)

	c.sql = strs.Append(c.sql, result, "\n")

	return c.sql
}

func (c *Linq) SqlOffset(limit, offset int) string {
	c.SqlSelect()

	result := strs.Format(`LIMIT %d OFFSET %d;`, limit, offset)

	c.sql = strs.Append(c.sql, result, "\n")

	return c.sql
}

func (c *Linq) SqlIndex() string {
	var result string
	var cols []*Column = []*Column{}
	from := c.from[0].model
	if from.UseSerie {
		col := from.Col(from.SerieField)
		cols = append(cols, col)
	} else {
		for _, key := range from.PrimaryKeys {
			col := from.Col(key)
			cols = append(cols, col)
		}
	}

	result = c.SqlColumDef(cols...)
	if c.Tp == TpData {
		result = strs.Format(`%s AS %s`, result, c.from[0].model.SourceField)
	}

	if len(result) > 0 {
		result = strs.Format(`RETURNING %s`, result)
	}

	c.sql = strs.Append(c.sql, result, "\n")

	return result
}

func (c *Linq) SqlKeys() string {
	result := c.SqlWhere()

	if len(result) > 0 {
		return result
	}

	var wh string
	for _, obj := range c.keys {
		if len(result) == 0 {
			wh = strs.Format(`%s=%v`, strs.Uppcase(obj.Col.name), et.Unquote(obj.Value))
		} else {
			wh = strs.Format(`AND %s=%v`, strs.Uppcase(obj.Col.name), et.Unquote(obj.Value))
		}
		result = strs.Append(result, wh, "\n")
	}

	if len(result) == 0 {
		result = `LIMIT 0`
	} else {
		result = strs.Format(`WHERE %s`, result)
	}

	c.sql = strs.Append(c.sql, result, "\n")

	return result
}

/**
*
**/
func (c *Linq) SqlInsert() string {
	model := c.from[0].model
	var fields string
	var values string

	for key, val := range *c.new {
		field := strs.Uppcase(key)
		value := et.Unquote(val)

		fields = strs.Append(fields, field, ", ")
		values = strs.Append(values, strs.Format(`%v`, value), ", ")
	}

	c.sql = strs.Format("INSERT INTO %s(%s)\nVALUES (%s)", model.Name, fields, values)

	c.SqlReturn()

	c.sql = strs.Format(`%s;`, c.sql)

	return c.sql
}

func (c *Linq) SqlUpdate() string {
	model := c.from[0].model
	var fieldValues string

	for key, val := range *c.new {
		field := strs.Uppcase(key)
		value := et.Unquote(val)

		if model.UseSource && field == strs.Uppcase(model.SourceField) {
			vals := strs.Uppcase(model.SourceField)
			atribs := c.new.Json(strs.Lowcase(field))

			for ak, av := range atribs {
				ak = strs.Lowcase(ak)
				av = et.Quote(av)

				vals = strs.Format(`jsonb_set(%s, '{%s}', '%v', true)`, vals, ak, av)
			}
			value = vals
		}

		fieldValue := strs.Format(`%s=%v`, field, value)
		fieldValues = strs.Append(fieldValues, fieldValue, ",\n")
	}

	c.sql = strs.Format(`UPDATE %s AS A SET %s`, model.Name, fieldValues)

	c.SqlWhere()

	c.SetAs(model, "A")

	c.SqlReturn()

	c.sql = strs.Format(`%s;`, c.sql)

	return c.sql
}

func (c *Linq) SqlDelete() string {
	model := c.from[0].model

	c.sql = strs.Format(`DELETE FROM %s`, model.Name)

	c.SqlWhere()

	c.SqlIndex()

	c.sql = strs.Format(`%s;`, c.sql)

	return c.sql
}
