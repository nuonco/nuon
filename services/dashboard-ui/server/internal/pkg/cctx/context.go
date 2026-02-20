package cctx

// ValueContext expresses only the read-only Value method,
// allowing gin, stdlib, or temporal contexts to be used interchangeably.
type ValueContext interface {
	Value(any) any
}
