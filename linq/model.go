package linq

import (
	"strings"

	"github.com/cgalvisleon/elvis/et"
	"github.com/cgalvisleon/elvis/jdb"
	"github.com/cgalvisleon/elvis/strs"
	"github.com/cgalvisleon/elvis/utility"
)

const BeforeInsert = 1
const AfterInsert = 2
const BeforeUpdate = 3
const AfterUpdate = 4
const BeforeDelete = 5
const AfterDelete = 6

type Trigger func(model *Model, old, new *et.Json, data et.Json) error

type Listener func(data et.Json)

type Model struct {
	Db                 int
	Database           *jdb.Db
	Name               string
	Description        string
	Define             string
	schema             *Schema
	Schema             string
	Table              string
	Definition         []*Column
	PrimaryKeys        []string
	ForeignKey         []*Reference
	Index              []string
	SourceField        string
	DateMakeField      string
	DateUpdateField    string
	SerieField         string
	CodeField          string
	ProjectField       string
	StateField         string
	Ddl                string
	integrityAtrib     bool
	integrityReference bool
	UseDateMake        bool
	UseDateUpdate      bool
	UseState           bool
	UseProject         bool
	UseSource          bool
	UseSerie           bool
	BeforeInsert       []Trigger
	AfterInsert        []Trigger
	BeforeUpdate       []Trigger
	AfterUpdate        []Trigger
	BeforeDelete       []Trigger
	AfterDelete        []Trigger
	OnListener         Listener
	Version            int
}

func (c *Model) Driver() string {
	return c.Database.Driver
}

func (c *Model) Describe() et.Json {
	var colums []et.Json = []et.Json{}
	for _, atrib := range c.Definition {
		colums = append(colums, atrib.Describe())
	}

	var foreignKey []et.Json = []et.Json{}
	for _, atrib := range c.ForeignKey {
		foreignKey = append(foreignKey, atrib.Describe())
	}

	var primaryKeys []string = append([]string{}, c.PrimaryKeys...)
	var index []string = append([]string{}, c.Index...)

	return et.Json{
		"name":               c.Name,
		"description":        c.Description,
		"schema":             c.Schema,
		"table":              c.Table,
		"colums":             colums,
		"primaryKeys":        primaryKeys,
		"foreignKeys":        foreignKey,
		"index":              index,
		"sourceField":        c.SourceField,
		"dateMakeField":      c.DateMakeField,
		"dateUpdateField":    c.DateUpdateField,
		"serieField":         c.SerieField,
		"codeField":          c.CodeField,
		"projectField":       c.ProjectField,
		"integrityAtrib":     c.integrityAtrib,
		"integrityReference": c.integrityReference,
		"useDateMake":        c.UseDateMake,
		"useDateUpdate":      c.UseDateUpdate,
		"useState":           c.UseState,
		"useProject":         c.UseProject,
		"useReciclig":        c.UseRecycle(),
		"useSerie":           c.UseSerie,
		"useSync":            c.UseSync(),
		"useListener":        !c.UseSync(),
		"model":              c.Model(),
	}
}

func (c *Model) Model() et.Json {
	var result et.Json = et.Json{}
	for _, col := range c.Definition {
		if !utility.ContainsInt([]int{TpColumn, TpAtrib, TpDetail}, col.Tp) {
			continue
		}

		if len(col.Atribs) > 0 {
			for _, atr := range col.Atribs {
				result.Set(atr.name, atr.Default)
			}
		} else if col.name == c.SourceField {
			continue
		} else if col.Type == "JSON" && col.Default == "[]" {
			result.Set(col.name, []et.Json{})
		} else if col.Type == "JSON" {
			result.Set(col.name, et.Json{})
		} else if col.Type == "JSONB" && col.Default == "[]" {
			result.Set(col.name, []et.Json{})
		} else if col.Type == "JSONB" {
			result.Set(col.name, et.Json{})
		} else if col.Type == "TIMESTAMP" && col.Default == "NOW()" {
			result.Set(col.name, utility.Now())
		} else if col.Type == "TIMESTAMP" && col.Default == "NULL" {
			result.Set(col.name, nil)
		} else if col.Type == "BOOLEAN" && col.Default == "TRUE" {
			result.Set(col.name, true)
		} else if col.Type == "BOOLEAN" && col.Default == "FALSE" {
			result.Set(col.name, false)
		} else {
			result.Set(col.name, col.Default)
		}
	}

	return result
}

func (c *Model) UseSync() bool {
	return c.schema.UseSync
}

func (c *Model) UseRecycle() bool {
	return c.schema.UseSync
}

/**
*
**/
func NewModel(schema *Schema, table, description string, version int) *Model {
	result := &Model{
		Db:                 schema.Db,
		Database:           schema.Database,
		schema:             schema,
		Schema:             schema.Name,
		Name:               strs.Append(strs.Lowcase(schema.Name), strs.Uppcase(table), "."),
		Description:        description,
		Table:              strs.Uppcase(table),
		Version:            version,
		SourceField:        schema.SourceField,
		DateMakeField:      schema.DateMakeField,
		DateUpdateField:    schema.DateUpdateField,
		SerieField:         schema.SerieField,
		CodeField:          schema.CodeField,
		ProjectField:       schema.ProjectField,
		StateField:         schema.StateField,
		integrityReference: true,
	}

	result.BeforeInsert = append(result.BeforeInsert, beforeInsert)
	result.AfterInsert = append(result.AfterInsert, afterInsert)
	result.BeforeUpdate = append(result.BeforeUpdate, beforeUpdate)
	result.AfterUpdate = append(result.AfterUpdate, afterUpdate)
	result.BeforeDelete = append(result.BeforeDelete, beforeDelete)
	result.AfterDelete = append(result.AfterDelete, afterDelete)

	schema.Models = append(schema.Models, result)

	return result
}

/**
* DDL
**/
func (c *Model) Init() error {
	sql := c.DDL()

	_, err := jdb.DBQDDL(c.Db, sql)
	if err != nil {
		return err
	}

	c.Define = sql

	return nil
}

func (c *Model) DDL() string {
	var result string
	var fields []string
	var index []string

	for _, column := range c.Definition {
		if column.Tp == TpColumn {
			fields = append(fields, column.DDL())
			if column.Indexed {
				if column.Unique {
					index = append(index, column.DDLUniqueIndex())
				} else {
					index = append(index, column.DDLIndex())
				}
			}
		}
	}

	// Definition create table with fields
	_fields := ""
	for i, def := range fields {
		if i == 0 {
			def = strs.Format("\n%s", def)
		}
		_fields = strs.Append(_fields, def, ",\n")
	}

	def := strs.Format(`CREATE TABLE IF NOT EXISTS %s(%s);`, c.Name, _fields) + "\n"
	result = strs.Append(result, def, "\n")

	// Definition create index
	for _, def := range index {
		result = strs.Append(result, def, "\n")
	}

	if len(c.Index) > 0 {
		result = result + "\n"
	}

	// Definition create primary key
	if len(c.PrimaryKeys) > 0 {
		pkey := strs.Replace(c.Table, ".", "_")
		pkey = strs.Replace(pkey, "-", "_") + "_pkey"
		pkey = strs.Lowcase(pkey)
		def = strs.Format(`ALTER TABLE IF EXISTS %s ADD CONSTRAINT %s PRIMARY KEY (%s);`, c.Name, pkey, strings.Join(c.PrimaryKeys, ", "))
		def = strs.Format(`SELECT core.create_constraint_if_not_exists('%s', '%s', '%s', '%s');`, c.Schema, c.Table, pkey, def)
		result = strs.Append(result, def, "\n")
	}

	// Definition create foreign key
	for _, ref := range c.ForeignKey {
		fkey := strs.Replace(c.Table, ".", "_")
		fkey = strs.Replace(fkey, "-", "_") + "_" + ref.Fkey + "_fkey"
		fkey = strs.Lowcase(fkey)
		def = strs.Format(`ALTER TABLE IF EXISTS %s ADD CONSTRAINT %s FOREIGN KEY (%s) %s;`, c.Name, fkey, ref.Fkey, ref.DDL())
		def = strs.Format(`SELECT core.create_constraint_if_not_exists('%s', '%s', '%s', '%s');`, c.Schema, c.Table, fkey, def)
		result = strs.Append(result, def, "\n")
	}

	if len(c.ForeignKey) > 0 {
		result = result + "\n"
	}

	c.Ddl = result

	return result
}

func (c *Model) DDLMigration() string {
	var fields []string

	table := c.Name
	c.Table = "NEW_TABLE"
	c.Name = strs.Append(c.Schema, c.Table, ",")
	ddl := c.DDL()

	for _, column := range c.Definition {
		fields = append(fields, column.name)
	}

	insert := strs.Format(`INSERT INTO %s(%s) SELECT %s FROM %s;`, c.Name, strings.Join(fields, ", "), strings.Join(fields, ", "), table)

	drop := strs.Format(`DROP TABLE %s CASCADE;`, c.Name)

	alter := strs.Format(`ALTER TABLE %s RENAME TO %s;`, c.Name, table)

	result := strs.Format(`%s %s %s %s`, ddl, insert, drop, alter)

	return result
}

func (c *Model) DropDDL() string {
	return strs.Format(`DROP TABLE IF EXISTS %s CASCADE;`, c.Name)
}

/**
*
**/
func (c *Model) Trigger(event int, trigger Trigger) {
	if event == BeforeInsert {
		c.BeforeInsert = append(c.BeforeInsert, trigger)
	} else if event == AfterInsert {
		c.AfterInsert = append(c.AfterInsert, trigger)
	} else if event == BeforeUpdate {
		c.BeforeUpdate = append(c.BeforeUpdate, trigger)
	} else if event == AfterUpdate {
		c.AfterUpdate = append(c.AfterUpdate, trigger)
	} else if event == BeforeDelete {
		c.BeforeDelete = append(c.BeforeDelete, trigger)
	} else if event == AfterDelete {
		c.AfterDelete = append(c.BeforeDelete, trigger)
	}
}

/**
*
**/
func (c *Model) Details(name, description string, _default any, details Details) {
	col := NewColumn(c, name, "", "DETAIL", _default)
	col.Tp = TpDetail
	col.Hidden = true
	col.Details = details
}

/**
*
**/
func (c *Model) Up() string {
	return strs.Uppcase(c.Name)
}

func (c *Model) Low() string {
	return strs.Lowcase(c.Name)
}

func (c *Model) ColIdx(name string) int {
	for i, item := range c.Definition {
		if item.Up() == strs.Uppcase(name) {
			return i
		}
	}

	return -1
}

func (c *Model) Col(name string) *Column {
	idx := c.ColIdx(name)
	if idx == -1 && !c.integrityAtrib {
		return NewVirtualAtrib(c, name, "", "text", "")
	} else if idx == -1 {
		return nil
	}

	return c.Definition[idx]
}

func (c *Model) As(as string) *FRom {
	return &FRom{
		model: c,
		as:    as,
	}
}

func (c *Model) Column(name string) *Column {
	return c.Col(name)
}

func (c *Model) TitleIdx(name string) int {
	for i, item := range c.Definition {
		if strs.Uppcase(item.Title) == strs.Uppcase(name) {
			return i
		}
	}

	return -1
}

func (c *Model) AtribIdx(name string) int {
	source := c.Col(c.SourceField)
	if source == nil {
		return -1
	}

	for i, item := range source.Atribs {
		if strs.Lowcase(item.name) == strs.Lowcase(name) {
			return i
		}
	}

	return -1
}

func (c *Model) Atrib(name string) *Column {
	idx := c.ColIdx(name)
	if idx == -1 {
		return nil
	}

	return c.Definition[idx]
}

func (c *Model) IndexIdx(name string) int {
	for i, _name := range c.Index {
		if strs.Uppcase(_name) == strs.Uppcase(name) {
			return i
		}
	}

	return -1
}

func (c *Model) IndexAdd(name string) int {
	idx := c.IndexIdx(name)
	if idx == -1 {
		c.Index = append(c.Index, name)
		idx = len(c.Index) - 1
	}

	return idx
}

func (c *Model) All() []*Column {
	result := c.Definition

	return result
}

func (c *Model) IntegrityAtrib(ok bool) *Model {
	c.integrityAtrib = ok

	return c
}

func (c *Model) IntegrityReference(ok bool) *Model {
	c.integrityReference = ok

	return c
}

/**
*
**/
func (c *Model) From() *Linq {
	return From(c)
}

func (c *Model) Data(sel ...any) *Linq {
	result := From(c)
	if !c.UseSource {
		result.Select(sel...)
	} else {
		result.Data(sel...)
	}

	return result
}

func (c *Model) Select(sel ...any) *Linq {
	result := From(c)
	result.Select(sel...)

	return result
}

/**
*
**/
func (c *Model) Insert(data et.Json) *Linq {
	tp := TpRow
	if c.UseSource {
		tp = TpData
	}

	result := NewLinq(ActInsert, c)
	result.SetTp(tp)
	result.data = data

	return result
}

func (c *Model) Update(data et.Json) *Linq {
	tp := TpRow
	if c.UseSource {
		tp = TpData
	}

	result := NewLinq(ActUpdate, c)
	result.SetTp(tp)
	result.data = data

	return result
}

func (c *Model) Delete() *Linq {
	tp := TpRow
	if c.UseSource {
		tp = TpData
	}

	result := NewLinq(ActDelete, c)
	result.SetTp(tp)

	return result
}

func (c *Model) Upsert(data et.Json) *Linq {
	tp := TpRow
	if c.UseSource {
		tp = TpData
	}

	result := NewLinq(ActUpsert, c)
	result.SetTp(tp)
	result.data = data

	return result
}

/**
* Row
**/
func (c *Model) InsertRow(data et.Json) *Linq {
	tp := TpRow

	result := NewLinq(ActInsert, c)
	result.SetTp(tp)
	result.data = data

	return result
}

func (c *Model) UpdateRow(data et.Json) *Linq {
	tp := TpRow

	result := NewLinq(ActUpdate, c)
	result.SetTp(tp)
	result.data = data

	return result
}

func (c *Model) DeleteRow() *Linq {
	tp := TpRow

	result := NewLinq(ActDelete, c)
	result.SetTp(tp)

	return result
}

func (c *Model) UpsertRow(data et.Json) *Linq {
	tp := TpRow

	result := NewLinq(ActUpsert, c)
	result.SetTp(tp)
	result.data = data

	return result
}
