package linq

import (
	"strings"

	e "github.com/cgalvisleon/elvis/json"
	"github.com/cgalvisleon/elvis/strs"
)

type Where struct {
	linq      *Linq
	connector string
	where     string
	val1      any
	operator  string
	val2      any
}

func (c *Where) Str1() string {
	result := ""
	switch v := c.val1.(type) {
	case []any:
		for _, vl := range v {
			def := c.Def(vl)
			result = strs.Append(result, def, ",")
		}
		result = strs.Format(`(%s)`, result)
	default:
		result = c.Def(v)
	}

	return result
}

func (c *Where) Str2() string {
	result := ""
	switch v := c.val2.(type) {
	case []any:
		for _, vl := range v {
			def := c.Def(vl)
			result = strs.Append(result, def, ",")
		}
		result = strs.Format(`(%s)`, result)
	default:
		result = c.Def(v)
	}

	return result
}

func StrToCols(str string) []string {
	str = strs.ReplaceAll(str, []string{" "}, "")
	cols := strings.Split(str, ",")

	return cols
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

func (c *Where) Def(val any) string {
	switch v := val.(type) {
	case Column:
		as := v.As(c.linq)
		return strs.Append(as, v.cast, "::")
	case *Column:
		as := v.As(c.linq)
		return strs.Append(as, v.cast, "::")
	case Col:
		as := v.from
		as = strs.Append(as, v.name, ".")
		return strs.Append(as, v.cast, "::")
	case *Col:
		as := v.from
		as = strs.Append(as, v.name, ".")
		return strs.Append(as, v.cast, "::")
	case SQL:
		return strs.Format(`%v`, v.val)
	default:
		return strs.Format(`%v`, e.Quoted(v))
	}
}

func (c *Where) Define(linq *Linq) *Where {
	var where string

	result := c.Str1()
	where = strs.Format(`%s %s`, result, c.operator)
	result = c.Str2()
	where = strs.Format(`%s %s`, where, result)

	c.where = where

	return c
}

func NewWhere(val1 any, operator string, val2 any) *Where {
	return &Where{val1: val1, operator: operator, val2: val2}
}
