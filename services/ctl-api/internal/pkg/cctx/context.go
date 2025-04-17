package cctx

// This package contains all of the objects that are added into the context.
//
// NOTE: we use string context values, as we can not use custom types in the gin context, and want to maintain
// consistency between the different context types.

// ValueContext expresses only the read-only Value method of stdlib contexts, allowing any of
// stdlib, gin, or temporal contexts to be provided for functions that only need to retrieve
// values from a context.
type ValueContext interface {
	Value(any) any
}
