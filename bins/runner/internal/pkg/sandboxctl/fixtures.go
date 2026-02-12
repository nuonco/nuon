package sandboxctl

import (
	"embed"
	"encoding/json"
	"fmt"
	"path"
)

//go:embed fixtures
var fixturesFS embed.FS

// Fixture holds the pre-loaded data for a sandbox response variant.
type Fixture struct {
	PlanContents        string         `json:"plan_contents"`
	PlanDisplayContents string         `json:"plan_display_contents"`
	StateJSON           string         `json:"state_json,omitempty"`
	Outputs             map[string]any `json:"outputs"`
}

// FixtureRegistry provides access to embedded fixture data.
type FixtureRegistry struct {
	fixtures map[JobCategory]map[ResponseVariant]*Fixture
}

// NewFixtureRegistry loads all embedded fixtures into memory.
func NewFixtureRegistry() (*FixtureRegistry, error) {
	r := &FixtureRegistry{
		fixtures: make(map[JobCategory]map[ResponseVariant]*Fixture),
	}

	for _, cat := range AllCategories() {
		r.fixtures[cat] = make(map[ResponseVariant]*Fixture)
		for _, variant := range []ResponseVariant{VariantDefault, VariantEmpty, VariantLarge} {
			f, err := loadFixture(cat, variant)
			if err != nil {
				return nil, fmt.Errorf("loading fixture %s/%s: %w", cat, variant, err)
			}
			r.fixtures[cat][variant] = f
		}
	}

	return r, nil
}

func loadFixture(cat JobCategory, variant ResponseVariant) (*Fixture, error) {
	p := path.Join("fixtures", string(cat), string(variant)+".json")
	data, err := fixturesFS.ReadFile(p)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", p, err)
	}

	var f Fixture
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", p, err)
	}
	return &f, nil
}

// Get returns the fixture for a category and variant. Returns nil if not found.
func (r *FixtureRegistry) Get(cat JobCategory, variant ResponseVariant) *Fixture {
	if cats, ok := r.fixtures[cat]; ok {
		if f, ok := cats[variant]; ok {
			return f
		}
	}
	return nil
}

// AvailableVariants returns the available variants for each category.
func (r *FixtureRegistry) AvailableVariants() map[JobCategory][]ResponseVariant {
	result := make(map[JobCategory][]ResponseVariant)
	for cat, variants := range r.fixtures {
		for variant := range variants {
			result[cat] = append(result[cat], variant)
		}
	}
	return result
}
