package toml

// Position represents a line/column location in a TOML document
type Position struct {
	Line      int
	Character int
}

// Range represents a span in a TOML document
type Range struct {
	Start Position
	End   Position
}

// Table represents a TOML table or array-of-tables
type Table struct {
	Name  string   // e.g. "connected_repo"
	Path  []string // split path for nested tables e.g. ["parent", "child"]
	Range Range    // line/column of the table header
}

// Key represents a TOML key (complete or partial)
type Key struct {
	Name   string   // e.g. "directory"
	Path   []string // fully qualified path e.g. ["connected_repo", "directory"]
	Prefix string   // prefix typed so far, if partial
	Range  Range    // line/column of the key occurrence
	Value  any      // parsed value if available
}

// TomlDocument represents a parsed TOML document
type TomlDocument struct {
	Tables       []Table
	Keys         []Key
	CurrentTable string         // current table context
	Values       map[string]any // map of key-path → value
}

// NewTomlDocument creates a new empty TomlDocument
func NewTomlDocument() *TomlDocument {
	return &TomlDocument{
		Tables: make([]Table, 0),
		Keys:   make([]Key, 0),
		Values: make(map[string]any),
	}
}

// SchemaPath returns the schema path for a given key path
func (doc *TomlDocument) SchemaPath(keyPath []string) []string {
	return keyPath
}

// TableSchemaPath returns the schema path for a table
func (doc *TomlDocument) TableSchemaPath(tableName string) []string {
	for _, table := range doc.Tables {
		if table.Name == tableName {
			return table.Path
		}
	}
	return []string{tableName}
}
