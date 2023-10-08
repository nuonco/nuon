package generics

import petname "github.com/dustinkirkland/golang-petname"

const (
	defaultWordsPerName int    = 3
	defaultSeperator    string = "-"
)

func DefaultName() string {
	return Name(defaultSeperator, defaultWordsPerName)
}

func Name(separator string, wordsPerName int) string {
	return petname.Generate(wordsPerName, defaultSeperator)
}
