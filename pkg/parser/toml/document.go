package toml

type Position struct {
	Line      int
	Character int
}

type Range struct {
	Start Position
	End   Position
}

type Table struct {
	Name  string
	Path  []string
	Range Range
}

type Key struct {
	Name   string
	Path   []string
	Prefix string
	Range  Range
	Value  any
}

type TomlDocument struct {
	Tables       []Table
	Keys         []Key
	CurrentTable string
	Values       map[string]any
}

func NewTomlDocument() *TomlDocument {
	return &TomlDocument{
		Tables: make([]Table, 0),
		Keys:   make([]Key, 0),
		Values: make(map[string]any),
	}
}

func (doc *TomlDocument) SchemaPath(keyPath []string) []string {
	return keyPath
}

func (doc *TomlDocument) TableSchemaPath(tableName string) []string {
	for _, table := range doc.Tables {
		if table.Name == tableName {
			return table.Path
		}
	}
	return []string{tableName}
}
