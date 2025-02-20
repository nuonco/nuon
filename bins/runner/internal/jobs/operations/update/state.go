package update

type handlerState struct {
	// set during the fetch/validate phase
	expectedVersion string
}
