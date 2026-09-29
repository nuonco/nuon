package cctx

// why: string context values, because gin cannot use custom types and the context types must stay consistent.
//
// NOTE: we use string context values, as we can not use custom types in the gin context, and want to maintain
// consistency between the different context types.

type ValueContext interface {
	Value(any) any
}
