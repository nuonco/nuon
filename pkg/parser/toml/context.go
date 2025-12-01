package toml

import "strings"

// TomlContext represents the context at a specific position in a TOML document
type TomlContext struct {
	CurrentTable string   // current table name
	KeyOnLine    string   // key on this line if any
	Prefix       string   // prefix before cursor (for completion)
	KeyPath      []string // fully qualified key path
}

// ContextAt returns the TOML context at a specific position
func (doc *TomlDocument) ContextAt(pos Position) TomlContext {
	ctx := TomlContext{
		CurrentTable: "", // Start with root context
		KeyPath:      make([]string, 0),
	}

	// Find the last table header BEFORE the cursor position
	// This ensures we only consider tables that actually precede the cursor
	for _, table := range doc.Tables {
		if table.Range.Start.Line < pos.Line {
			ctx.CurrentTable = table.Name
		} else {
			break // Stop when we reach tables after cursor
		}
	}

	// Find key on the current line
	for _, key := range doc.Keys {
		if key.Range.Start.Line == pos.Line {
			ctx.KeyOnLine = key.Name
			ctx.Prefix = key.Prefix
			ctx.KeyPath = key.Path
			break
		}
	}

	// Build fully qualified key path
	if ctx.CurrentTable != "" && ctx.KeyOnLine != "" {
		ctx.KeyPath = append(strings.Split(ctx.CurrentTable, "."), ctx.KeyOnLine)
	} else if ctx.KeyOnLine != "" {
		ctx.KeyPath = []string{ctx.KeyOnLine}
	} else if ctx.CurrentTable != "" {
		ctx.KeyPath = strings.Split(ctx.CurrentTable, ".")
	}

	return ctx
}
