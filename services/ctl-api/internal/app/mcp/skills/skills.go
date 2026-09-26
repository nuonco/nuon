// Package skills serves agent skills (focused, markdown how-to guides) over MCP.
// Skills are grouped into domains (skill:///{domain}/{skill}.md) and kept small
// so a client can browse the index without filling its context, then load only
// the skill it needs.
package skills

import (
	"embed"
	"fmt"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed apps/*.md
var files embed.FS

// Instructions is included in MCP server instructions so clients call
// list_skills for tasks a skill covers (e.g. writing or updating a README)
// instead of guessing at conventions from scratch.
const Instructions = "Before writing, reviewing, or suggesting changes to an app config, app/runbook README, Rego/OPA policy under policies/, or other app-config artifact, call list_skills to check for a relevant skill, then load_skill for the full guide."

// Skill is one focused how-to guide, scoped to a domain.
type Skill struct {
	Domain      string
	Name        string
	Description string
	Body        string
}

// URI is this skill's resource URI: skill:///{domain}/{name}.md
func (s Skill) URI() string {
	return fmt.Sprintf("skill:///%s/%s.md", s.Domain, s.Name)
}

// frontmatter is the YAML header on each skill file.
type frontmatter struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Domain      string `yaml:"domain"`
}

var all = mustLoad()

func mustLoad() []Skill {
	entries, err := files.ReadDir("apps")
	if err != nil {
		panic(fmt.Sprintf("skills: reading embedded apps dir: %v", err))
	}

	var loaded []Skill
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		raw, err := files.ReadFile("apps/" + entry.Name())
		if err != nil {
			panic(fmt.Sprintf("skills: reading %s: %v", entry.Name(), err))
		}

		skill, err := parse(raw)
		if err != nil {
			panic(fmt.Sprintf("skills: parsing %s: %v", entry.Name(), err))
		}
		loaded = append(loaded, skill)
	}

	sort.Slice(loaded, func(i, j int) bool {
		if loaded[i].Domain != loaded[j].Domain {
			return loaded[i].Domain < loaded[j].Domain
		}
		return loaded[i].Name < loaded[j].Name
	})

	return loaded
}

// parse splits a skill file into its YAML frontmatter and markdown body.
func parse(raw []byte) (Skill, error) {
	const delim = "---\n"

	content := string(raw)
	if !strings.HasPrefix(content, delim) {
		return Skill{}, fmt.Errorf("missing frontmatter delimiter")
	}
	rest := content[len(delim):]

	end := strings.Index(rest, delim)
	if end < 0 {
		return Skill{}, fmt.Errorf("missing closing frontmatter delimiter")
	}

	var fm frontmatter
	if err := yaml.Unmarshal([]byte(rest[:end]), &fm); err != nil {
		return Skill{}, fmt.Errorf("invalid frontmatter: %w", err)
	}
	if fm.Name == "" || fm.Domain == "" || fm.Description == "" {
		return Skill{}, fmt.Errorf("frontmatter must set name, domain, and description")
	}

	return Skill{
		Domain:      fm.Domain,
		Name:        fm.Name,
		Description: fm.Description,
		Body:        strings.TrimSpace(rest[end+len(delim):]),
	}, nil
}

// All returns every registered skill, sorted by domain then name.
func All() []Skill {
	return all
}

// Find looks up a skill by domain and name.
func Find(domain, name string) (Skill, bool) {
	for _, s := range all {
		if s.Domain == domain && s.Name == name {
			return s, true
		}
	}
	return Skill{}, false
}

// Index renders the one-line-per-skill catalog served at skill:///index.md.
func Index() string {
	var b strings.Builder
	b.WriteString("# Skill index\n\n")
	b.WriteString("Call load_skill(domain, name) or read skill:///{domain}/{name}.md for the full guide.\n\n")
	for _, s := range all {
		fmt.Fprintf(&b, "- `%s/%s`: %s\n", s.Domain, s.Name, s.Description)
	}
	return b.String()
}
