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
	Op   Op
	Diff string
}

type Diff struct {
	Key      string
	Diff     *DiffKey
	Children []*Diff
}

func (d *Diff) String(indent string) string {
	if d == nil {
		return ""
	}

	if d.Diff != nil {
		return fmt.Sprintf(indent+"%s: %s\n", d.Key, d.Diff.Diff)
	}
	diff := indent + d.Key + ":\n"
	for _, child := range d.Children {
		diff = diff + child.String(indent+"\t")
	}
	return diff
}

type DiffSummary struct {
	HasChanged bool
	Added      int
	Removed    int
	Changed    int
	Unchanged  int
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
