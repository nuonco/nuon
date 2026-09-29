package logs

import (
	"context"
	"fmt"
	"os"
	"time"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/table"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"

	"charm.land/lipgloss/v2"
	"github.com/nuonco/nuon/bins/cli/internal/ui/teaprogram"

	"golang.design/x/clipboard"

	"github.com/nuonco/nuon/bins/cli/internal/config"
	"github.com/nuonco/nuon/bins/cli/internal/ui/v3/common"
	"github.com/nuonco/nuon/pkg/cli/styles"
	"github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

const (
	maxSidebarWidth   int = 90
	minRequiredWidth  int = 100
	minRequiredHeight int = 15
)

type model struct {
	ctx context.Context
	cfg *config.Config
	api nuon.Client

	install_id   string
	deploy_id    string
	logstream_id string

	logStream    *models.AppLogStream
	loading      bool
	logs         map[string]*models.AppOtelLogRecord
	filteredLogs map[string]*models.AppOtelLogRecord
	logsCursor   string

	selectedLog *models.AppOtelLogRecord

	searchEnabled bool
	searchTerm    string

	altscreen    bool
	width        int
	height       int
	mainHeight   int
	sidebarWidth int

	message     common.StatusBarRequest
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
	s.Style = lipgloss.NewStyle().Foreground(styles.PrimaryColor)
	m := model{
		ctx: ctx,
		cfg: cfg,
		api: api,

		install_id:   install_id,
		deploy_id:    deploy_id,
		logstream_id: logstream_id,

		loading: true,
		spinner: s,

		logs: map[string]*models.AppOtelLogRecord{},

		searchInput: textinput.New(),
		help:        help.New(),
		message:     common.StatusBarRequest{Message: ""},

		keys:      keys,
		altscreen: true,
	}
	table := m.initTable()
	m.table = table
	return m
}

func (m *model) setMessage(message string, level string) {
	m.message.Message = message
	m.message.Level = level
}

func (m model) Init() tea.Cmd {
	m.getLatestLogs()
	return tea.Batch(common.TickCmd(common.DefaultRefreshInterval), m.spinner.Tick)
}

func (m *model) setLoading(v bool) {
	m.loading = v
}
func (m *model) resize() {
	vMargin := lipgloss.Height(m.headerView()) + lipgloss.Height(m.footerView()) + 2
	hMargin := 2
	m.mainHeight = m.height - vMargin
	m.sidebarWidth = max(maxSidebarWidth, int(m.width/3)) - 4

	m.searchInput.SetWidth(m.width - hMargin - lipgloss.Width(m.spinner.View()) - 3)

	m.table.SetHeight(m.height - vMargin)
	if m.selectedLog == nil {
		m.table.SetWidth(m.width - hMargin)
	} else {
		m.table.SetWidth(m.width - (hMargin + m.sidebarWidth + 2))
	}
	m.resizeTableColumns()

	m.details.SetWidth(m.sidebarWidth)
	m.details.SetHeight(m.height - vMargin)
	m.help.SetWidth(m.width)
}

func (m *model) handleResize(msg tea.WindowSizeMsg) {
	m.width = msg.Width
	m.height = msg.Height
	m.message.Width = msg.Width
	m.resize()
}

func (m *model) setSelected() {
	row := m.table.SelectedRow()
	// TODO: apply filter and keep an extra list of filtered rows
	if len(m.logs) > 1 {
		selectedLog, ok := m.logs[row[0]]
		if ok {
			m.selectedLog = selectedLog
			m.resize()
			m.details.SetContent(m.getDetailContent())
		} else {
			m.setMessage(fmt.Sprintf("[selected] log with id:%s not found", row[0]), "info")
		}
	}
}

func (m *model) resetSelected() {
	m.selectedLog = nil
	m.table.SetWidth(m.width - 2)
	m.resizeTableColumns()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {

	case common.TickMsg:
		m.setLoading(true)
		m.getLatestLogs()
		return m, common.TickCmd(common.DefaultRefreshInterval)

	case tea.WindowSizeMsg:
		m.handleResize(msg)

	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
			m.resize()

		case key.Matches(msg, m.keys.Enter):
			if m.searchEnabled && m.searchInput.Focused() {
				m.setLoading(true)
				m.SetSearchTerm(m.searchInput.Value())
				m.searchInput.Blur()
				m.table.Focus()
			} else if len(m.logs) > 0 && len(m.table.Rows()) > 0 {
				m.setSelected()
			}

		case key.Matches(msg, m.keys.Copy):
			if !m.searchEnabled && m.table.Focused() {
				row := m.table.SelectedRow()
				selectedLog, ok := m.logs[row[0]]
				if ok {
					selectedLogID := selectedLog.ID
					clipboard.Write(clipboard.FmtText, []byte(selectedLogID))
					m.setMessage(fmt.Sprintf("[copy] copied to clipboard \"%s\"", selectedLogID), "info")
				}
			}

		case key.Matches(msg, m.keys.Slash):
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
				m.resetSelected()
			} else if m.searchEnabled {
				m.ResetSearchInput()
				m.table.Focus()
			} else {
				return m, tea.Quit
			}
		}
	}

	if m.selectedLog != nil {
		m.details, cmd = m.details.Update(msg)
	} else if m.searchEnabled && m.searchInput.Focused() {
		m.searchInput, cmd = m.searchInput.Update(msg)
	} else {
		m.table, cmd = m.table.Update(msg)
	}

	return m, cmd
}

func (m model) headerView() string {
	s := ""
	spinner := ""
	if m.loading {
		spinner += m.spinner.View()
	}
	if m.searchEnabled {
		s += m.searchInput.View()
		return headerStyleActive.Render(lipgloss.JoinHorizontal(lipgloss.Top, s, spinner))
	}
	s += fmt.Sprintf("Logs for Install:%s deploy:%s", m.install_id, m.deploy_id)

	return headerStyle.Width(m.width - 2).Render(lipgloss.JoinHorizontal(lipgloss.Top, s, spinner))
}

func (m model) footerView() string {
	sections := []string{}
	rows := ""
	if m.searchTerm != "" {
		rows += fmt.Sprintf("Matches: %d | ", len(m.table.Rows()))
	}
	rows += fmt.Sprintf("Total Rows: %d", len(m.logs))
	sections = append(sections, styles.TextSubtle.Width(m.width).Render(rows))
	if m.message.Message != "" {
		sections = append(sections, common.StatusBar(m.message))
	}
	sections = append(sections, m.help.View(m.keys))
	return lipgloss.JoinVertical(lipgloss.Top, sections...)
}

func (m model) View() tea.View {
	v := tea.NewView(m.viewContent())
	v.AltScreen = true
	return v
}

func (m model) viewContent() string {
	if m.width == 0 {
		return ""

	} else if m.width < minRequiredWidth || m.height < minRequiredHeight {
		// TODO: make this message full screen
		content := common.FullPageDialog(common.FullPageDialogRequest{
			Width:   m.width,
			Height:  m.height,
			Padding: 2,
			Level:   "warning",
			Content: lipgloss.JoinVertical(
				lipgloss.Center,
				"  This screen is too small, please increase the width.  ",
				fmt.Sprintf("Minimum dimensions %d x %d.  ", minRequiredWidth, minRequiredHeight),
			),
		})
		return content

	}
	header := m.headerView()
	footer := m.footerView()

	main := ""
	tableStyle := appStyle
	if m.table.Focused() && m.selectedLog == nil {
		tableStyle = appStyle.BorderForeground(styles.BorderActiveColor)
	} else {
		tableStyle = appStyle.BorderForeground(styles.BorderInactiveColor)
	}
	tableView := tableStyle.Render(m.table.View())

	if m.selectedLog == nil {
		main = tableView
	} else {
		tableView = tableStyle.
			Width(m.table.Width()).
			Height(m.mainHeight).
			Render(m.table.View())
		logView := logModal.
			Width(m.sidebarWidth).
			Height(m.mainHeight).
			BorderForeground(styles.BorderActiveColor).
			Render(m.details.View())
		main = lipgloss.JoinHorizontal(lipgloss.Left, tableView, logView)
	}

	view := lipgloss.JoinVertical(
		lipgloss.Top,
		header, main, footer,
	)
	return view
}

func LogStreamApp(
	ctx context.Context,
	cfg *config.Config,
	api nuon.Client,
	install_id string,
	deploy_id string,
	logstream_id string,
) {
	if !cfg.Interactive {
		logStreamPlainText(ctx, api, logstream_id)
		return
	}

	m := initialModel(ctx, cfg, api, install_id, deploy_id, logstream_id)
	p := teaprogram.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Something has gone terribly wrong: %v", err)
		os.Exit(1)
	}
}

func logStreamPlainText(ctx context.Context, api nuon.Client, logstreamID string) {
	cursor := "0"
	for {
		logs, err := api.LogStreamReadLogs(ctx, logstreamID, cursor, "", nil)
		if err != nil {
			fmt.Printf("Error reading logs: %v\n", err)
			return
		}

		for _, log := range logs {
			fmt.Printf("[%s] %s %s: %s\n", log.Timestamp, log.SeverityText, log.ServiceName, log.Body)
		}

		for _, log := range logs {
			if log.Timestamp > cursor {
				cursor = log.Timestamp
			}
		}

		logStream, err := api.GetLogStream(ctx, logstreamID)
		if err != nil || !logStream.Open {
			return
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(5 * time.Second):
		}
	}
}
