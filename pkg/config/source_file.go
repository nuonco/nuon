package config

type SourceFileSetter interface {
	SetSourceFile(path string)
}

type SourceFileGetter interface {
	GetSourceFile() string
}
