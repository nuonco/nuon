package action

// Type definitions for execute form

type executeInputMapping struct {
	name  string // the ref name
	value string // default value
	input string // the input name
}

type executeFormSubmittedMsg struct {
	err error
}
