package diff

import "fmt"

type Op string

const (
	OpAdd     Op = "add"
	OpRemove  Op = "remove"
	OpChange  Op = "change"
	OpNoop    Op = "noop"
	OpUnknown Op = ""
)

type Diffable interface {
	Diff() (string, Op)
}

type DiffKey struct {
	Op   Op     `json:"op"`
	Diff string `json:"diff"`
}

type Diff struct {
	Key      string   `json:"key"`
	Diff     *DiffKey `json:"diff,omitempty"`
	Children []*Diff  `json:"children,omitempty"`
}

// String returns the full diff tree with +/-/~ prefixes indicating
// added, removed, changed, and unchanged fields.
// Empty unchanged fields (both old and new are "") are suppressed.
func (d *Diff) String(indent string) string {
	if d == nil {
		return ""
	}

	if d.Diff != nil {
		if d.Diff.Op == OpNoop && d.Diff.Diff == "'' (unchanged)" {
			return ""
		}
		return fmt.Sprintf(indent+"%s: %s\n", d.Key, d.Diff.Diff)
	}

	diff := indent + d.Key + ":\n"
	for _, child := range d.Children {
		diff = diff + child.String(indent+"\t")
	}
	return diff
}

// FormatChanged returns only the parts of the diff tree that have changes
// (added, removed, or changed). Unchanged fields and sections are omitted entirely.
func (d *Diff) FormatChanged(indent string) string {
	if d == nil {
		return ""
	}

	if d.Diff != nil {
		if d.Diff.Op == OpNoop {
			return ""
		}
		prefix := opPrefix(d.Diff.Op)
		return fmt.Sprintf("%s%s%s: %s\n", prefix, indent, d.Key, d.Diff.Diff)
	}

	var childOutput string
	for _, child := range d.Children {
		childOutput += child.FormatChanged(indent + "\t")
	}
	if childOutput == "" {
		return ""
	}
	return fmt.Sprintf("%s%s:\n%s", indent, d.Key, childOutput)
}

func opPrefix(op Op) string {
	switch op {
	case OpAdd:
		return "+ "
	case OpRemove:
		return "- "
	case OpChange:
		return "~ "
	default:
		return "  "
	}
}

type DiffSummary struct {
	HasChanged bool `json:"has_changed"`
	Added      int  `json:"added"`
	Removed    int  `json:"removed"`
	Changed    int  `json:"changed"`
	Unchanged  int  `json:"unchanged"`
}

func (d *Diff) Summary() DiffSummary {
	summary := DiffSummary{}
	if d == nil {
		return summary
	}

	if d.Diff != nil {
		switch d.Diff.Op {
		case OpAdd:
			summary.Added++
			summary.HasChanged = true
		case OpRemove:
			summary.Removed++
			summary.HasChanged = true
		case OpChange:
			summary.Changed++
			summary.HasChanged = true
		case OpNoop:
			summary.Unchanged++
		}
	} else {
		for _, child := range d.Children {
			childSummary := child.Summary()
			summary.Added += childSummary.Added
			summary.Removed += childSummary.Removed
			summary.Changed += childSummary.Changed
			summary.Unchanged += childSummary.Unchanged
			if childSummary.HasChanged {
				summary.HasChanged = true
			}
		}
	}
	return summary
}

type DiffOption func(*Diff)

func WithKey(key string) DiffOption {
	return func(dt *Diff) {
		dt.Key = key
	}
}

func WithChildren(children ...*Diff) DiffOption {
	return func(dt *Diff) {
		dt.Children = append(dt.Children, children...)
	}
}

func WithStringDiff(old, new string) DiffOption {
	return withDiff(&StringDiffer{old: old, new: new})
}

func withDiff(diff Diffable) DiffOption {
	df, op := diff.Diff()
	return func(dt *Diff) {
		dt.Diff = &DiffKey{
			Op:   op,
			Diff: df,
		}
	}
}

func NewDiff(opts ...DiffOption) *Diff {
	dt := Diff{}
	for _, opt := range opts {
		opt(&dt)
	}
	return &dt
}

type StringDiffer struct {
	old, new string
}

func (d *StringDiffer) Diff() (string, Op) {
	if d.old != d.new {
		op := OpChange
		if d.old == "" {
			op = OpAdd
		} else if d.new == "" {
			op = OpRemove
		}
		return fmt.Sprintf("'%s' -> '%s'", d.old, d.new), op
	}
	return fmt.Sprintf("'%s' (unchanged)", d.old), OpNoop
}
