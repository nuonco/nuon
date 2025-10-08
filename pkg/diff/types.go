package diff

// DiffEntryType represents the type of difference in a resource or property
type DiffEntryType int

const (
	// EntryUnchanged indicates no change
	EntryUnchanged DiffEntryType = iota
	// EntryRemoved indicates the resource or property was removed
	EntryRemoved
	// EntryAdded indicates the resource or property was added
	EntryAdded
	// EntryModified indicates the resource or property was modified
	EntryModified
	// EntryError indicates an error occurred during the diff
	EntryError
)

// String returns a human-readable string representation for DiffEntryType
func (t DiffEntryType) String() string {
	switch t {
	case EntryUnchanged:
		return "unchanged"
	case EntryRemoved:
		return "deleted"
	case EntryAdded:
		return "created"
	case EntryModified:
		return "modified"
	case EntryError:
		return "error"
	}
	return "unknown"
}

// Symbol returns a single character representation for DiffEntryType (for visual diff)
func (t DiffEntryType) Symbol() string {
	switch t {
	case EntryUnchanged:
		return " "
	case EntryRemoved:
		return "-"
	case EntryAdded:
		return "+"
	case EntryModified:
		return "~"
	case EntryError:
		return "!"
	}
	return "?"
}

// DiffEntry represents a single change in a resource
type DiffEntry struct {
	Path     string                 `json:"path,omitempty"`
	Original interface{}            `json:"original,omitempty"`
	Applied  interface{}            `json:"applied,omitempty"`
	Type     DiffEntryType          `json:"type"`
	Changes  map[string]interface{} `json:"changes,omitempty"` // For nested changes
	Payload  string                 `json:"payload,omitempty"` // For raw diff content
}

// ResourceDiff represents the diff for a Kubernetes resource
// Used by both Helm and Kubernetes manifest implementations
type ResourceDiff struct {
	// Version identifier for this diff format
	Version string `json:"_version"`

	// Resource identification
	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	Kind      string `json:"kind,omitempty"`
	ApiPath   string `json:"api,omitempty"`
	Resource  string `json:"resource,omitempty"`

	// Operation details
	Operation string        `json:"op,omitempty"`
	Type      DiffEntryType `json:"type"`
	ErrorMsg  string        `json:"error,omitempty"`
	DryRun    bool          `json:"dry_run,omitempty"`

	// Detailed changes
	Entries []DiffEntry `json:"entries"`
}

// PlanContents is a common structure for both Helm and Kubernetes manifest plan contents
type PlanContents struct {
	Plan        string         `json:"plan"`
	Op          string         `json:"op"`
	ContentDiff []ResourceDiff `json:"content_diff"`
}
