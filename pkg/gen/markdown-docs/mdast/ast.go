package mdast

import (
	"fmt"
	"sort"
	"strings"
)

// Node represents a node in the markdown AST
type Node interface {
	Render() string
}

// Document represents a complete markdown document
type Document struct {
	frontmatter map[string]string
	nodes       []Node
}

// NewDocument creates a new markdown document
func NewDocument() *Document {
	return &Document{
		frontmatter: make(map[string]string),
		nodes:       []Node{},
	}
}

// AddFrontmatter adds YAML frontmatter to the document
func (d *Document) AddFrontmatter(data map[string]string) {
	for k, v := range data {
		d.frontmatter[k] = v
	}
}

// AddNode adds a node to the document
func (d *Document) AddNode(node Node) {
	d.nodes = append(d.nodes, node)
}

// AddHeading adds a heading node
func (d *Document) AddHeading(level int, text string) {
	d.nodes = append(d.nodes, &Heading{Level: level, Text: text})
}

// AddParagraph adds a paragraph node
func (d *Document) AddParagraph(text string) {
	d.nodes = append(d.nodes, &Paragraph{Text: text})
}

// AddTable adds a table node
func (d *Document) AddTable(table *Table) {
	d.nodes = append(d.nodes, table)
}

// AddListItem adds a list item node
func (d *Document) AddListItem(text string) {
	d.nodes = append(d.nodes, &ListItem{Text: text})
}

// AddCodeBlock adds a code block node
func (d *Document) AddCodeBlock(language, code string) {
	d.nodes = append(d.nodes, &CodeBlock{Language: language, Code: code})
}

// AddRaw adds raw markdown text
func (d *Document) AddRaw(text string) {
	d.nodes = append(d.nodes, &RawText{Text: text})
}

// Render renders the entire document to markdown
func (d *Document) Render() string {
	var sb strings.Builder

	// Render frontmatter
	if len(d.frontmatter) > 0 {
		sb.WriteString("---\n")
		// Sort keys for deterministic output
		keys := make([]string, 0, len(d.frontmatter))
		for k := range d.frontmatter {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			sb.WriteString(fmt.Sprintf("%s: '%s'\n", k, d.frontmatter[k]))
		}
		sb.WriteString("---\n\n")
	}

	// Render nodes
	for _, node := range d.nodes {
		sb.WriteString(node.Render())
	}

	return sb.String()
}

// Heading represents a markdown heading
type Heading struct {
	Level int
	Text  string
}

func (h *Heading) Render() string {
	return fmt.Sprintf("%s %s\n\n", strings.Repeat("#", h.Level), h.Text)
}

// Paragraph represents a markdown paragraph
type Paragraph struct {
	Text string
}

func (p *Paragraph) Render() string {
	return p.Text + "\n\n"
}

// Table represents a markdown table
type Table struct {
	Headers []string
	Rows    [][]string
}

// NewTable creates a new table with headers
func NewTable(headers []string) *Table {
	return &Table{
		Headers: headers,
		Rows:    [][]string{},
	}
}

// AddRow adds a row to the table
func (t *Table) AddRow(cells []string) {
	t.Rows = append(t.Rows, cells)
}

func (t *Table) Render() string {
	var sb strings.Builder

	// Render header row
	sb.WriteString("|")
	for _, header := range t.Headers {
		sb.WriteString(" ")
		sb.WriteString(header)
		sb.WriteString(" |")
	}
	sb.WriteString("\n")

	// Render separator row
	sb.WriteString("|")
	for range t.Headers {
		sb.WriteString("----------|")
	}
	sb.WriteString("\n")

	// Render data rows
	for _, row := range t.Rows {
		sb.WriteString("|")
		for _, cell := range row {
			sb.WriteString(" ")
			sb.WriteString(cell)
			sb.WriteString(" |")
		}
		sb.WriteString("\n")
	}
	sb.WriteString("\n")

	return sb.String()
}

// ListItem represents a list item
type ListItem struct {
	Text string
}

func (l *ListItem) Render() string {
	return fmt.Sprintf("- %s\n", l.Text)
}

// CodeBlock represents a fenced code block
type CodeBlock struct {
	Language string
	Code     string
}

func (c *CodeBlock) Render() string {
	return fmt.Sprintf("```%s\n%s\n```\n\n", c.Language, c.Code)
}

// RawText represents raw markdown text
type RawText struct {
	Text string
}

func (r *RawText) Render() string {
	return r.Text
}

// Section represents a logical section with multiple nodes
type Section struct {
	nodes []Node
}

// NewSection creates a new section
func NewSection() *Section {
	return &Section{
		nodes: []Node{},
	}
}

// AddHeading adds a heading to the section
func (s *Section) AddHeading(level int, text string) {
	s.nodes = append(s.nodes, &Heading{Level: level, Text: text})
}

// AddParagraph adds a paragraph to the section
func (s *Section) AddParagraph(text string) {
	s.nodes = append(s.nodes, &Paragraph{Text: text})
}

// AddListItem adds a list item to the section
func (s *Section) AddListItem(text string) {
	s.nodes = append(s.nodes, &ListItem{Text: text})
}

// AddCodeBlock adds a code block to the section
func (s *Section) AddCodeBlock(language, code string) {
	s.nodes = append(s.nodes, &CodeBlock{Language: language, Code: code})
}

func (s *Section) Render() string {
	var sb strings.Builder
	for _, node := range s.nodes {
		sb.WriteString(node.Render())
	}
	return sb.String()
}

// Helper functions

// Code wraps text in backticks
func Code(text string) string {
	return fmt.Sprintf("`%s`", text)
}

// EscapeMDX escapes special characters that could be interpreted as MDX/JSX
func EscapeMDX(s string) string {
	s = strings.ReplaceAll(s, "{", "\\{")
	s = strings.ReplaceAll(s, "}", "\\}")
	s = strings.ReplaceAll(s, "<", "\\<")
	s = strings.ReplaceAll(s, ">", "\\>")
	s = strings.ReplaceAll(s, "|", "\\|")
	return s
}
