package workflow

import (
	"path/filepath"
	"strings"

	"mvdan.cc/sh/v3/expand"
	"mvdan.cc/sh/v3/syntax"
)

// Shell expressions retain their existing expansion inside the container.
func directContainerCommand(command string) []string {
	file, err := syntax.NewParser(syntax.Variant(syntax.LangPOSIX)).Parse(strings.NewReader(command), "")
	if err != nil || len(file.Stmts) != 1 {
		return nil
	}
	stmt := file.Stmts[0]
	if stmt.Negated || stmt.Background || stmt.Coprocess || len(stmt.Redirs) != 0 {
		return nil
	}
	call, ok := stmt.Cmd.(*syntax.CallExpr)
	if !ok || len(call.Assigns) != 0 || len(call.Args) == 0 {
		return nil
	}
	args := make([]string, 0, len(call.Args))
	for _, word := range call.Args {
		if !literalParts(word.Parts, false) {
			return nil
		}
		value, err := expand.Literal(&expand.Config{}, word)
		if err != nil {
			return nil
		}
		args = append(args, value)
	}
	if !filepath.IsAbs(args[0]) {
		return nil
	}
	return args
}

func literalParts(parts []syntax.WordPart, quoted bool) bool {
	for _, part := range parts {
		switch part := part.(type) {
		case *syntax.Lit:
			if !quoted && strings.ContainsAny(part.Value, `\*?[]~{}`) {
				return false
			}
		case *syntax.SglQuoted:
			if part.Dollar {
				return false
			}
		case *syntax.DblQuoted:
			if part.Dollar || !literalParts(part.Parts, true) {
				return false
			}
		default:
			return false
		}
	}
	return true
}
