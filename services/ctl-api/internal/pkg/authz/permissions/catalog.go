package permissions

import (
	"cmp"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"slices"
)

const (
	OrgInspect       PermissionID = "org.inspect"
	AppInspect       PermissionID = "app.inspect"
	InstallInspect   PermissionID = "install.inspect"
	InstallUpdate    PermissionID = "install.update"
	RunnerJobInspect PermissionID = "runner_job.inspect"
	RunnerJobCancel  PermissionID = "runner_job.cancel"
	LogStreamAppend  PermissionID = "log_stream.append"
)

type PermissionDefinition struct {
	ID          PermissionID   `json:"id"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Scopes      []ResourceKind `json:"scopes"`
	Targets     []ResourceKind `json:"targets,omitempty"`
	Members     []PermissionID `json:"members,omitempty"`
}

type Catalog struct {
	definitions map[PermissionID]PermissionDefinition
	expanded    map[PermissionID][]PermissionID
	revision    string
}

func DefaultCatalog() (*Catalog, error) {
	org := []ResourceKind{KindOrg}
	app := []ResourceKind{KindOrg, KindApp}
	install := []ResourceKind{KindOrg, KindApp, KindInstall}
	job := []ResourceKind{KindOrg, KindApp, KindInstall, KindRunnerJob}
	stream := []ResourceKind{KindOrg, KindApp, KindInstall, KindRunnerJob, KindLogStream}
	return NewCatalog([]PermissionDefinition{
		{ID: OrgInspect, Title: "Inspect org", Description: "Read the org inspection response.", Scopes: org, Targets: []ResourceKind{KindOrg}},
		{ID: AppInspect, Title: "Inspect app", Description: "Read the app inspection response.", Scopes: app, Targets: []ResourceKind{KindApp}},
		{ID: InstallInspect, Title: "Inspect install", Description: "Read the install inspection response, including embedded metadata.", Scopes: install, Targets: []ResourceKind{KindInstall}},
		{ID: InstallUpdate, Title: "Update install", Description: "Update an existing install.", Scopes: install, Targets: []ResourceKind{KindInstall}},
		{ID: InstallDelete, Title: "Delete install", Description: "Delete an existing install.", Scopes: install, Targets: []ResourceKind{KindInstall}},
		{ID: RunnerJobInspect, Title: "Inspect job", Description: "Read the job inspection response.", Scopes: job, Targets: []ResourceKind{KindRunnerJob}},
		{ID: RunnerJobCancel, Title: "Cancel job", Description: "Cancel an existing job, subject to its state constraints.", Scopes: job, Targets: []ResourceKind{KindRunnerJob}},
		{ID: LogStreamRead, Title: "Read log stream", Description: "Read the whole stream, including shared or copied records.", Scopes: stream, Targets: []ResourceKind{KindLogStream}},
		{ID: LogStreamAppend, Title: "Append log stream", Description: "Append records with the existing parent-copy side effects.", Scopes: []ResourceKind{KindLogStream}, Targets: []ResourceKind{KindLogStream}},
		{ID: RunnerJobRead, Title: "Job Reader", Description: "Read jobs and their validated associated streams.", Scopes: []ResourceKind{KindRunnerJob}, Members: []PermissionID{RunnerJobInspect, LogStreamRead}},
		{ID: InstallRead, Title: "Install Reader", Description: "Read installs and their owned jobs and logs.", Scopes: []ResourceKind{KindInstall}, Members: []PermissionID{InstallInspect, RunnerJobRead, LogStreamRead}},
		{ID: InstallWrite, Title: "Install write", Description: "Broad install writes, including deletion and job cancellation; not an Editor preset.", Scopes: []ResourceKind{KindInstall}, Members: []PermissionID{InstallUpdate, InstallDelete, RunnerJobCancel}},
		{ID: AppRead, Title: "App Reader", Description: "Read apps and their supported owned installs, jobs and logs.", Scopes: []ResourceKind{KindApp}, Members: []PermissionID{AppInspect, InstallRead, RunnerJobRead, LogStreamRead}},
		{ID: OrgRead, Title: "Org Reader", Description: "Read the org and supported owned apps, installs, jobs and logs.", Scopes: org, Members: []PermissionID{OrgInspect, AppRead, InstallRead, RunnerJobRead, LogStreamRead}},
	})
}

func NewCatalog(definitions []PermissionDefinition) (*Catalog, error) {
	c := &Catalog{
		definitions: make(map[PermissionID]PermissionDefinition, len(definitions)),
		expanded:    make(map[PermissionID][]PermissionID, len(definitions)),
	}
	for _, definition := range definitions {
		if !validScopeID(string(definition.ID)) || definition.Title == "" || definition.Description == "" {
			return nil, fmt.Errorf("permission definition requires a stable ID, title and description")
		}
		if _, exists := c.definitions[definition.ID]; exists {
			return nil, fmt.Errorf("duplicate permission definition %q", definition.ID)
		}
		if len(definition.Scopes) == 0 || (len(definition.Targets) == 0) == (len(definition.Members) == 0) {
			return nil, fmt.Errorf("permission %q requires scopes and either targets or bundle members", definition.ID)
		}
		for _, kind := range append(slices.Clone(definition.Scopes), definition.Targets...) {
			if err := kind.Validate(); err != nil {
				return nil, fmt.Errorf("permission %q: %w", definition.ID, err)
			}
		}
		for _, scope := range definition.Scopes {
			for _, target := range definition.Targets {
				if !scopeSupports(scope, target, definition.ID) {
					return nil, fmt.Errorf("permission %q has no matcher from %q to %q", definition.ID, scope, target)
				}
			}
		}
		definition.Scopes = sortedUnique(definition.Scopes)
		definition.Targets = sortedUnique(definition.Targets)
		definition.Members = sortedUnique(definition.Members)
		c.definitions[definition.ID] = definition
	}
	for id := range c.definitions {
		if _, err := c.expand(id, make(map[PermissionID]bool)); err != nil {
			return nil, err
		}
	}
	data, err := json.Marshal(c.Definitions())
	if err != nil {
		return nil, err
	}
	hash := sha256.Sum256(data)
	c.revision = hex.EncodeToString(hash[:])
	return c, nil
}

func (c *Catalog) expand(id PermissionID, visiting map[PermissionID]bool) ([]PermissionID, error) {
	if expanded, ok := c.expanded[id]; ok {
		return expanded, nil
	}
	definition, ok := c.definitions[id]
	if !ok {
		return nil, fmt.Errorf("unknown permission reference %q", id)
	}
	if visiting[id] {
		return nil, fmt.Errorf("permission bundle cycle at %q", id)
	}
	visiting[id] = true
	operations := []PermissionID{id}
	if len(definition.Members) != 0 {
		operations = nil
		for _, member := range definition.Members {
			leaves, err := c.expand(member, visiting)
			if err != nil {
				return nil, err
			}
			operations = append(operations, leaves...)
		}
		for _, operation := range operations {
			for _, scope := range definition.Scopes {
				if !slices.Contains(c.definitions[operation].Scopes, scope) {
					return nil, fmt.Errorf("bundle %q at %q scope cannot include %q", id, scope, operation)
				}
			}
		}
	}
	delete(visiting, id)
	c.expanded[id] = sortedUnique(operations)
	return c.expanded[id], nil
}

func (c *Catalog) Revision() string {
	return c.revision
}

func (c *Catalog) Definitions() []PermissionDefinition {
	definitions := make([]PermissionDefinition, 0, len(c.definitions))
	for _, definition := range c.definitions {
		definition.Scopes = slices.Clone(definition.Scopes)
		definition.Targets = slices.Clone(definition.Targets)
		definition.Members = slices.Clone(definition.Members)
		definitions = append(definitions, definition)
	}
	slices.SortFunc(definitions, func(a, b PermissionDefinition) int {
		return cmp.Compare(a.ID, b.ID)
	})
	return definitions
}

func sortedUnique[T ~string](values []T) []T {
	values = slices.Clone(values)
	slices.Sort(values)
	return slices.Compact(values)
}
