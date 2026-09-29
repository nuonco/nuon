package logs

func (m *model) ToggleSearch() {
	m.searchEnabled = !m.searchEnabled
	if m.searchEnabled {
		m.searchInput.Focus()
	}
}

func (m *model) SetSearchTerm(term string) {
	m.searchTerm = term
}

func (m *model) ResetSearchInput() {
	m.searchTerm = ""
	m.searchInput.Reset()
	m.searchInput.Blur()
	m.searchEnabled = false
	m.selectedLog = nil
}
