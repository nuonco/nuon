package config

// SourceFileSetter is implemented by config types that can track their source file path.
// This is used during parsing to record which file a config object was loaded from.
type SourceFileSetter interface {
	SetSourceFile(path string)
}

// SourceFileGetter is implemented by config types that can report their source file path.
type SourceFileGetter interface {
	GetSourceFile() string
}
