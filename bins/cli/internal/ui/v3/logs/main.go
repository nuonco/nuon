/*

An inline tui for viewing logs from the terminal.

*/

package logs

import (
	"context"
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"golang.design/x/clipboard"

	"github.com/nuonco/nuon-go"
	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/bins/cli/internal/config"
)

type model struct {
	// configs and api client
	ctx context.Context
	cfg *config.Config
	api nuon.Client

	// NOTE(fd): these should likely live elsewhere
	// fixed vars
	install_id   string
	deploy_id    string
	logstream_id string

	// dynamic state
	logStream    *models.AppLogStream
	loading      bool
	logs         map[string]*models.AppOtelLogRecord
	filteredLogs map[string]*models.AppOtelLogRecord
	logsCursor   string // this is the cursor for the next request for logs

	// we want the SelectedLog to be updated when the cursor changes to allow the users to
	// open the sidebar and continue to scroll logs which would change the log on display in the sidebar.
	// cursor      int // cursor for the selected table (perhaps this shoudl move into the table model and can be sent up via a message)
	selectedLog *models.AppOtelLogRecord

	searchEnabled bool
	searchTerm    string

	altscreen bool

	// components
	message     string
	keys        keyMap
	table       table.Model
	spinner     spinner.Model
	details     viewport.Model
	help        help.Model
	searchInput textinput.Model
}

func initialModel(
	ctx context.Context,
	cfg *config.Config,
	api nuon.Client,
	install_id string,
	deploy_id string,
	logstream_id string,
) model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205")) // .Padding(0, 0, 0, 1)
	m := model{
		ctx: ctx,
		cfg: cfg,
		api: api,

		install_id:   install_id,
		deploy_id:    deploy_id,
		logstream_id: logstream_id,

		loading: false,
		spinner: s,

		logs: map[string]*models.AppOtelLogRecord{},

		searchInput: textinput.New(),
		help:        help.New(),
		message:     "-",

		keys:      keys,
		altscreen: true,
	}

	table := m.initTable()
	m.table = table
	return m
}

func (m *model) setMessage(message string) {
	// for use from within update
	m.message = message
}

func (m model) Init() tea.Cmd {
	m.getLatestLogs()
	return tea.Batch(tick, m.spinner.Tick)
}

func (m *model) setLoading(v bool) {
	// used to fire off a loading indicator
	// not really used to set loading to false, that happens downstream usually
	m.loading = v
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	// var cmds []tea.Cmd
	switch msg := msg.(type) {

	// handle tick: fetch data
	case tickMsg:
		m.setLoading(true)
		m.getLatestLogs()
		return m, tick

	// handle re-size
	case tea.WindowSizeMsg:
		// when the window resizes, we need to set the width of our components
		vMarginHeight := lipgloss.Height(m.headerView()) + lipgloss.Height(m.footerView()) + 1
		hMargin := 2

		// header search input
		m.searchInput.Width = msg.Width - hMargin - lipgloss.Width(m.spinner.View()) - 3 // 3 is the width of the caret

		// logs table
		m.table.SetWidth(msg.Width - hMargin)
		m.table.SetHeight(msg.Height - vMarginHeight) // idk why we need the three here
		m.resizeTableColumns()

		// 3 is the height of the modal header (8 and 6 are scaling factors)
		m.details = viewport.New(msg.Width-4, msg.Height-vMarginHeight-2) // idk where this 2 comes from
		m.help.Width = msg.Width
		m.setMessage(fmt.Sprintf("resize: w: %d h: %d - table(%d x %d)", msg.Width, msg.Height, m.table.Width(), m.table.Height()))

	// handle keys
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit): // "ctrl+c", "q"
			return m, tea.Quit
		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll

		case key.Matches(msg, m.keys.Enter):
			if m.searchEnabled && m.searchInput.Focused() {
				m.setLoading(true)
				m.SetSearchTerm(m.searchInput.Value())
				m.searchInput.Blur()
				m.table.Focus()
			} else if len(m.logs) > 0 && len(m.table.Rows()) > 0 {
				// set the selectedLog to the log corresponding to selected row's log id
				row := m.table.SelectedRow()
				// TODO: apply filter and keep an extra list of filtered rows
				if len(m.logs) > 1 { // gt 1 because we raise the
					selectedLog, ok := m.logs[row[0]]
					if ok {
						m.selectedLog = selectedLog
						m.details.SetContent(m.getDetailContent())
					} else {
						m.setMessage(fmt.Sprintf("[selected] log with id:%s not found", row[0]))
					}
				}

			}

		case key.Matches(msg, m.keys.Copy):
			if !m.searchEnabled && m.table.Focused() {
				row := m.table.SelectedRow()
				selectedLog, ok := m.logs[row[0]]
				if ok {
					selectedLogID := selectedLog.ID
					clipboard.Write(clipboard.FmtText, []byte(selectedLogID))
					m.setMessage(fmt.Sprintf("[copy] copied to clipboard \"%s\"", selectedLogID))
				}
			}

		case key.Matches(msg, m.keys.Slash):
			// search is only usable from the table view
			if m.selectedLog == nil {
				if m.searchEnabled && !m.searchInput.Focused() {
					m.searchInput.Focus()
				} else if !m.searchEnabled {
					m.ToggleSearch()
				}
			}
			return m, cmd

		case key.Matches(msg, m.keys.Esc):

			if m.selectedLog != nil {
				// if the model is open, close the modal
				m.selectedLog = nil
			} else if m.searchEnabled {
				// if there is a search term, reset the field and focus the table
				m.ResetSearchInput()
				m.table.Focus()
			} else {
				// otherwise, quit
				return m, tea.Quit
			}
		}
	}

	// pass the message to the relevant component
	if m.selectedLog != nil { // send to log modal
		m.details, cmd = m.details.Update(msg)
	} else if m.searchEnabled && m.searchInput.Focused() { // send to search input
		// we use this term to ensure we can have the
		m.searchInput, cmd = m.searchInput.Update(msg)
	} else { // send to table
		m.table, cmd = m.table.Update(msg)
	}

	// cmds = append(cmds, cmd)
	return m, cmd // tea.Batch(cmds...)
}

func (m model) headerView() string {
	s := ""
	spinner := "" // placeholder text
	if m.loading {
		spinner += m.spinner.View()
	}
	if m.searchEnabled {
		s += m.searchInput.View()
		return headerStyleActive.Render(lipgloss.JoinHorizontal(lipgloss.Top, s, spinner)) + "\n"
	}
	s += fmt.Sprintf("Logs for Install:%s deploy:%s", m.install_id, m.deploy_id)

	return headerStyle.Render(lipgloss.JoinHorizontal(lipgloss.Top, s, spinner)) + "\n"
}

func (m model) footerView() string {
	s := ""
	if m.searchTerm != "" {
		s += fmt.Sprintf("Matches: %d | ", len(m.table.Rows()))
	}
	s += fmt.Sprintf("Total Rows: %d", len(m.logs))
	if !m.help.ShowAll {
		s += "          " + "\n"
	}
	s += messageStyle.Render("> "+m.message) + "\n"

	// Help View
	s += m.help.View(m.keys)
	return s
}

func (m model) View() string {
	s := ""
	// HEADER
	s += m.headerView()

	// Main Content
	if m.selectedLog != nil {
		s += logModal.Render(m.details.View()) + "\n"
	} else {

		s += appStyle.Render(m.table.View()) + "\n"
	}

	// Metadata Footer
	s += m.footerView()
	return s
}

func MakeMeAnApp(
	ctx context.Context,
	cfg *config.Config,
	api nuon.Client,
	install_id string,
	deploy_id string,
	logstream_id string,
) {
	// initialize the model
	m := initialModel(ctx, cfg, api, install_id, deploy_id, logstream_id)

	// initialize the program
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Something has gone terribly wrong: %v", err)
		os.Exit(1)
	}
}
