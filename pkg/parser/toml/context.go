package toml

import "strings"

type TomlContext struct {
	CurrentTable string
	KeyOnLine    string
	Prefix       string
	KeyPath      []string
}

func (doc *TomlDocument) ContextAt(pos Position) TomlContext {
	ctx := TomlContext{
		CurrentTable: "",
		KeyPath:      make([]string, 0),
	}

	for _, table := range doc.Tables {
		if table.Range.Start.Line < pos.Line {
			ctx.CurrentTable = table.Name
		} else {
			break
		}
	}

	for _, key := range doc.Keys {
		if key.Range.Start.Line == pos.Line {
			ctx.KeyOnLine = key.Name
			ctx.Prefix = key.Prefix
			ctx.KeyPath = key.Path
			break
		}
	}

	if ctx.CurrentTable != "" && ctx.KeyOnLine != "" {
		ctx.KeyPath = append(strings.Split(ctx.CurrentTable, "."), ctx.KeyOnLine)
	} else if ctx.KeyOnLine != "" {
		ctx.KeyPath = []string{ctx.KeyOnLine}
	} else if ctx.CurrentTable != "" {
		ctx.KeyPath = strings.Split(ctx.CurrentTable, ".")
	}

	return ctx
}
