package module

import (
	"github.com/cgalvisleon/elvis/linq"
)

var SchemaModule *linq.Schema

func DefineSchemaModule() error {
	if SchemaModule != nil {
		return nil
	}

	SchemaModule = linq.NewSchema(0, "module", true, true, true)

	return nil
}
