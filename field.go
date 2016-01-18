package gorma

import (
	"fmt"
	"strings"

	"github.com/raphael/goa/design"
	"github.com/raphael/goa/goagen/codegen"
)

// NewRelationalFieldDefinition creates and parses a field from the design attributes
func NewRelationalFieldDefinition(name string, a *design.AttributeDefinition) (*RelationalFieldDefinition, error) {
	f := &RelationalFieldDefinition{}
	f.a = a
	t := a.Definition().Type
	f.Datatype = codegen.GoNativeType(t)
	f.Name = codegen.Goify(name, true)
	if strings.HasSuffix(f.Name, "Id") {
		f.Name = f.Name[:len(f.Name)-2] + "ID"
	}
	f.DatabaseFieldName = f.Name
	err := f.Parse()
	return f, err

}

// Generating fields

// Definition returns the field's struct definition
func (f *RelationalFieldDefinition) Definition() string {

	var desc, fieldType, fieldName, pointer string
	fieldType = f.Datatype
	if f.HasOne != "" {
		fieldType = fmt.Sprintf("%s.%s", strings.ToLower(deModel(f.HasOne)),
			deModel(f.HasOne))
	}
	if f.HasMany != "" {
		fieldType = fmt.Sprintf("[]%s.%s", strings.ToLower(deModel(f.HasMany)), deModel(f.HasMany))
	}
	fieldName = f.Name
	if f.Nullable {
		pointer = "*"
	}
	if f.Description != "" {
		desc = fmt.Sprintf("//%s", f.Description)
	}
	return fmt.Sprintf("%s \t %s%s %s %s\n", fieldName, pointer, fieldType, f.Tags(), desc)

}

// Tags returns teh sql and gorm struct tags for the Definition
func (f *RelationalFieldDefinition) Tags() string {
	var sqltags, gormtags, jsontags string
	var dirty bool
	if f.SQLTag != "" {
		sqltags = fmt.Sprintf("sql:\"%s\"", f.SQLTag)
		dirty = true
	}
	if f.PrimaryKey {
		if f.Aliased {
			gormtags = fmt.Sprintf("gorm:\"%s,column:%s\"", "primary_key", f.DatabaseFieldName)
		} else {
			gormtags = fmt.Sprintf("gorm:\"%s\"", "primary_key")
		}
		dirty = true
	} else {
		if f.Aliased {
			gormtags = fmt.Sprintf("gorm:\"column:%s\"", f.DatabaseFieldName)
			dirty = true
		}
	}
	if dirty {
		tags := strings.TrimSpace(strings.Join([]string{jsontags, sqltags, gormtags}, " "))
		return fmt.Sprintf("`%s`", tags)
	}
	return ""
}

// Parsing Methods

// Parse populates all the attributes of the Field
func (f *RelationalFieldDefinition) Parse() error {
	if err := f.ParsePrimaryKey(); err != nil {
		return err
	}
	if err := f.ParseSQLTag(); err != nil {
		return err
	}
	if err := f.ParseTimestamps(); err != nil {
		return err
	}
	if err := f.ParseAlias(); err != nil {
		return err
	}
	if err := f.ParseBelongsTo(); err != nil {
		return err
	}
	if err := f.ParseHasOne(); err != nil {
		return err
	}
	if err := f.ParseHasMany(); err != nil {
		return err
	}
	if err := f.ParseManyToMany(); err != nil {
		return err
	}
	if err := f.ParseDescription(); err != nil {
		return err
	}
	return nil
}

func (f *RelationalFieldDefinition) ParseDescription() error {
	if f.a.Description != "" {
		f.Description = f.a.Description
	}
	return nil
}

//ParseTimestamps populates the timestamps field
func (f *RelationalFieldDefinition) ParseTimestamps() error {
	if _, ok := metaLookup(f.a.Metadata, gengorma.MetaTimestampCreated); ok {
		f.Timestamp = true
		f.Datatype = "time.Time"
		f.Nullable = false
	}
	if _, ok := metaLookup(f.a.Metadata, gengorma.MetaTimestampUpdated); ok {
		f.Timestamp = true
		f.Datatype = "time.Time"
		f.Nullable = false
	}
	if _, ok := metaLookup(f.a.Metadata, gengorma.MetaTimestampDeleted); ok {
		f.Timestamp = true
		f.Datatype = "time.Time"
		f.Nullable = true
	}

	return nil

}

//ParseSQLTag populates the SQLTag field
func (f *RelationalFieldDefinition) ParseSQLTag() error {
	// is it a primary key?
	if gt, ok := metaLookup(f.a.Metadata, gengorma.MetaSQLTag); ok {
		f.SQLTag = gt
	}
	return nil

}

//ParseBelongsTo populates the SQLTag field
func (f *RelationalFieldDefinition) ParseBelongsTo() error {
	if gt, ok := metaLookup(f.a.Metadata, gengorma.MetaBelongsTo); ok {
		f.BelongsTo = gt
	}
	return nil

}

//ParseManyToMany populates the ManyToMany relationships
func (f *RelationalFieldDefinition) ParseManyToMany() error {
	if gt, ok := metaLookup(f.a.Metadata, gengorma.MetaManyToMany); ok {
		f.Many2Many = gt
	}
	return nil
}

//ParseHasOne populates the SQLTag field
func (f *RelationalFieldDefinition) ParseHasOne() error {
	if gt, ok := metaLookup(f.a.Metadata, gengorma.MetaHasOne); ok {
		f.HasOne = gt
	}
	return nil

}

//ParseHasMany populates the SQLTag field
func (f *RelationalFieldDefinition) ParseHasMany() error {
	if gt, ok := metaLookup(f.a.Metadata, gengorma.MetaHasMany); ok {
		f.HasMany = gt
	}
	return nil

}

//ParseAlias populates the DatabaseFieldName field
func (f *RelationalFieldDefinition) ParseAlias() error {

	if gt, ok := metaLookup(f.a.Metadata, gengorma.MetaGormTag); ok {
		f.Aliased = true
		f.DatabaseFieldName = gt
	}
	return nil

}

//ParsePrimaryKey populates the primary key tag
func (f *RelationalFieldDefinition) ParsePrimaryKey() error {
	// is it a primary key?
	if gt, ok := metaLookup(f.a.Metadata, gengorma.MetaPrimaryKey); ok {
		if strings.Contains(gt, "primary_key") {
			f.PrimaryKey = true
		}
	}
	if f.Name == "ID" || f.Name == "Id" || f.Name == "id" {
		f.PrimaryKey = true
	}
	return nil
}
