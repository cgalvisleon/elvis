package master

import (
	"github.com/cgalvisleon/elvis/et"
	"github.com/cgalvisleon/elvis/strs"
)

func (c *Node) InsertValues(data et.Json) (fields, values string) {
	for k, v := range data {
		k = strs.Uppcase(k)
		v = et.Unquote(v)

		if len(fields) == 0 {
			fields = k
			values = strs.Format(`%v`, v)
		} else {
			fields = strs.Format(`%s, %s`, fields, k)
			values = strs.Format(`%s, %v`, values, v)
		}
	}

	return
}

func (c *Node) UpsertValues(data et.Json) (fields, values, fieldValue string) {
	for k, v := range data {
		k = strs.Uppcase(k)
		v = et.Unquote(v)

		if len(fieldValue) == 0 {
			fields = k
			values = strs.Format(`%v`, v)
			if k == "_IDT" {
				v = "-1"
			}
			fieldValue = strs.Format(`%s=%v`, k, v)
		} else {
			fields = strs.Format(`%s, %s`, fields, k)
			values = strs.Format(`%s, %v`, values, v)
			if k == "_IDT" {
				v = "-1"
			}
			fieldValue = strs.Append(fieldValue, strs.Format(`%s=%v`, k, v), ",\n")
		}
	}

	return
}

func (c *Node) SqlField(schema, table string, data et.Json) string {
	fields, _ := c.InsertValues(data)
	result := strs.Format(`INSERT INTO %s.%s (%s)`, strs.Lowcase(schema), strs.Uppcase(table), fields)
	result = strs.Append(result, "VALUES", "\n")
	return result
}

func (c *Node) ToSql(schema, table, idT string, data et.Json, action string) (string, bool) {
	var result string
	var ok bool
	if action == "INSERT" {
		_, values := c.InsertValues(data)
		result = strs.Format(`(%s)`, values)
	} else if action == "UPDATE" {
		fields, values, fieldValue := c.UpsertValues(data)
		result = strs.Format(`INSERT INTO %s.%s (%s)`, strs.Lowcase(schema), strs.Uppcase(table), fields)
		result = strs.Append(result, strs.Format(`VALUES (%s)`, values), "\n")
		result = strs.Append(result, "ON CONFLICT (_IDT) DO UPDATE SET", "\n")
		result = strs.Append(result, fieldValue, "\n")
		result = strs.Format(`%s;`, result)
	} else if action == "DELETE" {
		result = strs.Format(`DELETE FROM %s.%s WHERE _IDT=%s`, strs.Lowcase(schema), strs.Uppcase(table), idT)
	}

	return result, ok
}
