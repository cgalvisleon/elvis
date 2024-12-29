package linq

import (
	"github.com/cgalvisleon/elvis/et"
	"github.com/cgalvisleon/elvis/jdb"
	"github.com/cgalvisleon/elvis/strs"
)

var schemas []*Schema = []*Schema{}

type Schema struct {
	Db              int
	Database        *jdb.Db
	Name            string
	Description     string
	Define          string
	UseSync         bool
	UseRecycle      bool
	UseSerie        bool
	SourceField     string
	DateMakeField   string
	DateUpdateField string
	SerieField      string
	CodeField       string
	ProjectField    string
	StateField      string
	Models          []*Model
}

func NewSchema(db int, name string, sync, recycle, serie bool) *Schema {
	result := &Schema{
		Db:              db,
		Name:            strs.Lowcase(name),
		Database:        jdb.DB(db),
		UseSync:         sync,
		UseRecycle:      recycle,
		UseSerie:        serie,
		SourceField:     "_DATA",
		DateMakeField:   "DATE_MAKE",
		DateUpdateField: "DATE_UPDATE",
		SerieField:      "INDEX",
		CodeField:       "CODE",
		ProjectField:    "PROJECT_ID",
		StateField:      "_STATE",
		Models:          []*Model{},
	}

	result.Init()
	schemas = append(schemas, result)

	return result
}

func GetSchema(name string) *Schema {
	for _, item := range schemas {
		if strs.Uppcase(item.Name) == strs.Uppcase(name) {
			return item
		}
	}

	return nil
}

/**
*
**/
func (c *Schema) Describe() et.Json {
	var models []et.Json = []et.Json{}
	for _, model := range c.Models {
		models = append(models, model.Describe())
	}

	return et.Json{
		"name":            c.Name,
		"description":     c.Description,
		"database":        c.Database.Dbname,
		"useSync":         c.UseSync,
		"useRecycle":      c.UseRecycle,
		"useSerie":        c.UseSerie,
		"source_field":    c.SourceField,
		"dateMakeField":   c.DateMakeField,
		"dateUpdateField": c.DateUpdateField,
		"serieField":      c.SerieField,
		"codeField":       c.CodeField,
		"projectField":    c.ProjectField,
		"models":          models,
	}
}

func (c *Schema) Init() error {
	_, err := jdb.CreateSchema(c.Db, c.Name)
	if err != nil {
		return err
	}

	c.Define = strs.Format(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"; CREATE SCHEMA IF NOT EXISTS "%s";`, c.Name)

	return nil
}

/**
*
**/
func (c *Schema) Model(name string) *Model {
	for _, item := range c.Models {
		if strs.Uppcase(item.Name) == strs.Uppcase(name) {
			return item
		}
	}

	return nil
}
