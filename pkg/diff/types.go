package diff

type DiffEntryType int

const (
	EntryUnchanged DiffEntryType = iota
	EntryRemoved
	EntryAdded
	EntryModified
	EntryError
)

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

type DiffEntry struct {
	Path     string                 `json:"path,omitempty"`
	Original interface{}            `json:"original,omitempty"`
	Applied  interface{}            `json:"applied,omitempty"`
	Type     DiffEntryType          `json:"type"`
	Changes  map[string]interface{} `json:"changes,omitempty"`
	Payload  string                 `json:"payload,omitempty"`
}

type ResourceDiff struct {
	Version string `json:"_version"`

	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	Kind      string `json:"kind,omitempty"`
	ApiPath   string `json:"api,omitempty"`
	Resource  string `json:"resource,omitempty"`

	Operation string        `json:"op,omitempty"`
	Type      DiffEntryType `json:"type"`
	ErrorMsg  string        `json:"error,omitempty"`
	DryRun    bool          `json:"dry_run,omitempty"`

	Entries []DiffEntry `json:"entries"`
}

type PlanContents struct {
	Plan        string         `json:"plan"`
	Op          string         `json:"op"`
	ContentDiff []ResourceDiff `json:"content_diff"`
}
