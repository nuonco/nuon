package logs

func (m *model) ToggleSearch() {
	m.searchEnabled = !m.searchEnabled
	if m.searchEnabled {
		m.searchInput.Focus()
	}
}

func (m *model) SetSearchTerm(term string) {
	m.searchTerm = term

	// apply search term
}

func (m *model) ResetSearchInput() {
	m.searchTerm = ""
	m.searchInput.Reset()
	m.searchInput.Blur()
	m.searchEnabled = false
	// this is really bordering on overloading but it's the easies way to return to the state we want
	m.selectedLog = nil
}
