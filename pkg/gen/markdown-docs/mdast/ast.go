package mdast

import (
	"fmt"
	"sort"
	"strings"
)

type Node interface {
	Render() string
}

type Document struct {
	frontmatter map[string]string
	nodes       []Node
}

func NewDocument() *Document {
	return &Document{
		frontmatter: make(map[string]string),
		nodes:       []Node{},
	}
}

func (d *Document) AddFrontmatter(data map[string]string) {
	for k, v := range data {
		d.frontmatter[k] = v
	}
}

func (d *Document) AddNode(node Node) {
	d.nodes = append(d.nodes, node)
}

func (d *Document) AddHeading(level int, text string) {
	d.nodes = append(d.nodes, &Heading{Level: level, Text: text})
}

func (d *Document) AddParagraph(text string) {
	d.nodes = append(d.nodes, &Paragraph{Text: text})
}

func (d *Document) AddTable(table *Table) {
	d.nodes = append(d.nodes, table)
}

func (d *Document) AddListItem(text string) {
	d.nodes = append(d.nodes, &ListItem{Text: text})
}

func (d *Document) AddCodeBlock(language, code string) {
	d.nodes = append(d.nodes, &CodeBlock{Language: language, Code: code})
}

func (d *Document) AddRaw(text string) {
	d.nodes = append(d.nodes, &RawText{Text: text})
}

func (d *Document) Render() string {
	var sb strings.Builder

	if len(d.frontmatter) > 0 {
		sb.WriteString("---\n")
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

	for _, node := range d.nodes {
		sb.WriteString(node.Render())
	}

	return sb.String()
}

type Heading struct {
	Level int
	Text  string
}

func (h *Heading) Render() string {
	return fmt.Sprintf("%s %s\n\n", strings.Repeat("#", h.Level), h.Text)
}

type Paragraph struct {
	Text string
}

func (p *Paragraph) Render() string {
	return p.Text + "\n\n"
}

type Table struct {
	Headers []string
	Rows    [][]string
}

func NewTable(headers []string) *Table {
	return &Table{
		Headers: headers,
		Rows:    [][]string{},
	}
}

func (t *Table) AddRow(cells []string) {
	t.Rows = append(t.Rows, cells)
}

func (t *Table) Render() string {
	var sb strings.Builder

	sb.WriteString("|")
	for _, header := range t.Headers {
		sb.WriteString(" ")
		sb.WriteString(header)
		sb.WriteString(" |")
	}
	sb.WriteString("\n")

	sb.WriteString("|")
	for range t.Headers {
		sb.WriteString("----------|")
	}
	sb.WriteString("\n")

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

type ListItem struct {
	Text string
}

func (l *ListItem) Render() string {
	return fmt.Sprintf("- %s\n", l.Text)
}

type CodeBlock struct {
	Language string
	Code     string
}

func (c *CodeBlock) Render() string {
	return fmt.Sprintf("```%s\n%s\n```\n\n", c.Language, c.Code)
}

type RawText struct {
	Text string
}

func (r *RawText) Render() string {
	return r.Text
}

type Section struct {
	nodes []Node
}

func NewSection() *Section {
	return &Section{
		nodes: []Node{},
	}
}

func (s *Section) AddHeading(level int, text string) {
	s.nodes = append(s.nodes, &Heading{Level: level, Text: text})
}

func (s *Section) AddParagraph(text string) {
	s.nodes = append(s.nodes, &Paragraph{Text: text})
}

func (s *Section) AddListItem(text string) {
	s.nodes = append(s.nodes, &ListItem{Text: text})
}

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

func Code(text string) string {
	return fmt.Sprintf("`%s`", text)
}

func EscapeMDX(s string) string {
	s = strings.ReplaceAll(s, "{", "\\{")
	s = strings.ReplaceAll(s, "}", "\\}")
	s = strings.ReplaceAll(s, "<", "\\<")
	s = strings.ReplaceAll(s, ">", "\\>")
	s = strings.ReplaceAll(s, "|", "\\|")
	return s
}
