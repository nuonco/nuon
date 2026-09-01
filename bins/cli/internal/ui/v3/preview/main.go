package preview

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/pkg/cli/styles"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

type Branch struct {
	ID   string
	Name string
}

type Data struct {
	BranchName    string
	ConfigID      string
	PreviewConfig *models.AppAppBranchPreviewConfig
	Sources       *models.HelpersListPreviewSourcesResult
	Installs      []*models.AppInstall
}

type LoadBranchFunc func(context.Context, string) (*Data, error)

type Options struct {
	BranchID   string
	Mode       models.AppAppBranchRunPreviewMode
	PRNumber   *int
	GitRef     string
	HeadSHA    string
	InstallID  string
	CurrentRef string
}

type Result struct {
	BranchID string
	ConfigID string
	Request  *models.ServicePreviewRunRequest
}

type step int

const (
	stepBranch step = iota
	stepMode
	stepSourceScope
	stepSourceKind
	stepSource
	stepInstall
)

type item struct {
	title       string
	description string
	value       string
}

type branchLoadedMsg struct {
	data *Data
	err  error
}

type model struct {
	ctx        context.Context
	branches   []Branch
	loadBranch LoadBranchFunc
	opts       Options
	result     Result

	step      step
	items     []item
	cursor    int
	width     int
	height    int
	loading   bool
	err       error
	cancelled bool
	spinner   spinner.Model

	data          *Data
	currentSource *models.ServicePreviewRunRequest
	request       models.ServicePreviewRunRequest
}

func initialModel(ctx context.Context, branches []Branch, loadBranch LoadBranchFunc, opts Options) model {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(styles.AccentColor)

	m := model{
		ctx:        ctx,
		branches:   branches,
		loadBranch: loadBranch,
		opts:       opts,
		spinner:    sp,
	}
	m.resetBranchValues()

	if opts.BranchID == "" {
		m.setBranchStep()
	} else {
		m.result.BranchID = opts.BranchID
		m.loading = true
	}
	return m
}

func (m model) Init() tea.Cmd {
	if m.loading {
		return tea.Batch(m.loadSelectedBranch(), m.spinner.Tick)
	}
	return nil
}

func (m model) loadSelectedBranch() tea.Cmd {
	return func() tea.Msg {
		data, err := m.loadBranch(m.ctx, m.result.BranchID)
		return branchLoadedMsg{data: data, err: err}
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case branchLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.data = msg.data
		m.result.ConfigID = msg.data.ConfigID
		m.applyDefaults()
		if done := m.advanceAfterLoad(); done {
			return m, tea.Quit
		}
		return m, nil
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.cancelled = true
			return m, tea.Quit
		case "esc":
			if m.loading {
				return m, nil
			}
			if m.goBack() {
				return m, nil
			}
			m.cancelled = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil
		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
			return m, nil
		case "enter":
			if m.loading || m.err != nil || len(m.items) == 0 {
				return m, nil
			}
			return m.selectCurrent()
		}
	default:
		if m.loading {
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

func (m *model) applyDefaults() {
	if m.request.Mode == "" {
		m.request.Mode = models.AppAppBranchRunPreviewModePlanDashOnly
		if m.data.PreviewConfig != nil && m.data.PreviewConfig.Mode != "" {
			m.request.Mode = m.data.PreviewConfig.Mode
		}
	}
	if m.request.InstallID == "" && m.data.PreviewConfig != nil {
		m.request.InstallID = m.defaultInstallID()
	}
	m.currentSource = currentRequest(m.opts.CurrentRef, m.data.Sources, m.opts.HeadSHA)
}

func (m *model) advanceAfterLoad() bool {
	if m.opts.Mode == "" {
		m.setModeStep()
		return false
	}
	if m.request.Source == "" {
		m.setSourceScopeStep()
		return false
	}
	m.fillExplicitSourceSHA()
	return m.advanceToInstall()
}

func (m *model) advanceToInstall() bool {
	if m.request.Mode == models.AppAppBranchRunPreviewModeBuildDashOnly {
		m.finish()
		return true
	}
	if m.opts.InstallID != "" {
		m.finish()
		return true
	}
	m.setInstallStep()
	return false
}

func (m model) selectCurrent() (tea.Model, tea.Cmd) {
	selected := m.items[m.cursor].value
	switch m.step {
	case stepBranch:
		m.result.BranchID = selected
		m.resetBranchValues()
		m.loading = true
		return m, tea.Batch(m.loadSelectedBranch(), m.spinner.Tick)
	case stepMode:
		m.request.Mode = models.AppAppBranchRunPreviewMode(selected)
		if m.request.Source == "" {
			m.setSourceScopeStep()
			return m, nil
		}
		if m.advanceToInstall() {
			return m, tea.Quit
		}
	case stepSourceScope:
		if selected == "current" {
			mode := m.request.Mode
			installID := m.request.InstallID
			m.request = *m.currentSource
			m.request.Mode = mode
			m.request.InstallID = installID
			if m.advanceToInstall() {
				return m, tea.Quit
			}
			return m, nil
		}
		m.setSourceKindStep()
	case stepSourceKind:
		m.setSourceStep(selected)
	case stepSource:
		m.selectSource(selected)
		if m.advanceToInstall() {
			return m, tea.Quit
		}
	case stepInstall:
		m.request.InstallID = selected
		m.finish()
		return m, tea.Quit
	}
	return m, nil
}

func (m *model) resetBranchValues() {
	m.data = nil
	m.currentSource = nil
	m.result.ConfigID = ""
	m.request = models.ServicePreviewRunRequest{
		Mode:      m.opts.Mode,
		GitRef:    m.opts.GitRef,
		HeadSha:   m.opts.HeadSHA,
		InstallID: m.opts.InstallID,
	}
	if m.opts.PRNumber != nil {
		m.request.Source = models.AppAppBranchRunPreviewSourcePr
		m.request.PrNumber = int64(*m.opts.PRNumber)
	} else if m.opts.GitRef != "" {
		m.request.Source = models.AppAppBranchRunPreviewSourceBranch
	}
}

func (m *model) setBranchStep() {
	items := make([]item, len(m.branches))
	for i, branch := range m.branches {
		items[i] = item{title: branch.Name, description: branch.ID, value: branch.ID}
	}
	m.setStep(stepBranch, items, "")
}

func (m *model) setModeStep() {
	m.setStep(stepMode, []item{
		{title: "Plan only", description: "Build and plan without applying", value: string(models.AppAppBranchRunPreviewModePlanDashOnly)},
		{title: "Apply", description: "Build, plan, and apply changes", value: string(models.AppAppBranchRunPreviewModeApply)},
		{title: "Build only", description: "Build and validate without an install", value: string(models.AppAppBranchRunPreviewModeBuildDashOnly)},
	}, string(m.request.Mode))
}

func (m *model) setSourceScopeStep() {
	items := make([]item, 0, 2)
	if m.currentSource != nil {
		items = append(items, item{
			title:       "Current",
			description: sourceDescription(m.currentSource),
			value:       "current",
		})
	}
	items = append(items, item{
		title:       "Any pull request or branch",
		description: "Choose from sources available for this app branch",
		value:       "any",
	})
	m.setStep(stepSourceScope, items, "current")
}

func (m *model) setSourceKindStep() {
	items := make([]item, 0, 2)
	if m.data.Sources != nil && len(m.data.Sources.PullRequests) > 0 {
		items = append(items, item{title: "Pull request", value: "pr"})
	}
	if m.data.Sources != nil && len(m.data.Sources.Branches) > 0 {
		items = append(items, item{title: "Git branch", value: "branch"})
	}
	m.setStep(stepSourceKind, items, "")
}

func (m *model) setSourceStep(kind string) {
	if kind == "pr" {
		items := make([]item, len(m.data.Sources.PullRequests))
		for i, pr := range m.data.Sources.PullRequests {
			items[i] = item{
				title:       fmt.Sprintf("#%d · %s", pr.PrNumber, pr.Title),
				description: pr.HeadRef,
				value:       "pr:" + strconv.FormatInt(pr.PrNumber, 10),
			}
		}
		m.setStep(stepSource, items, "")
		return
	}

	items := make([]item, len(m.data.Sources.Branches))
	for i, branch := range m.data.Sources.Branches {
		items[i] = item{title: branch.Name, value: "branch:" + branch.Name}
	}
	m.setStep(stepSource, items, "")
}

func (m *model) setInstallStep() {
	items := make([]item, len(m.data.Installs))
	for i, install := range m.data.Installs {
		description := install.ID
		if install.AppBranchID != "" && install.AppBranchID != m.result.BranchID {
			branchName := install.AppBranchID
			if install.AppBranch != nil && install.AppBranch.Name != "" {
				branchName = install.AppBranch.Name
			}
			description = fmt.Sprintf("On branch %s · %s", branchName, install.ID)
		}
		items[i] = item{title: install.Name, description: description, value: install.ID}
	}
	m.setStep(stepInstall, items, m.request.InstallID)
}

func (m *model) setStep(next step, items []item, selected string) {
	m.step = next
	m.items = items
	m.cursor = 0
	for i, option := range items {
		if option.value == selected {
			m.cursor = i
			break
		}
	}
}

func (m *model) selectSource(value string) {
	kind, selected, _ := strings.Cut(value, ":")
	m.request.GitRef = ""
	m.request.PrNumber = 0
	m.request.HeadSha = m.opts.HeadSHA

	if kind == "pr" {
		number, _ := strconv.ParseInt(selected, 10, 64)
		m.request.Source = models.AppAppBranchRunPreviewSourcePr
		m.request.PrNumber = number
		for _, pr := range m.data.Sources.PullRequests {
			if pr.PrNumber == number && m.request.HeadSha == "" {
				m.request.HeadSha = pr.HeadSha
				break
			}
		}
		return
	}

	m.request.Source = models.AppAppBranchRunPreviewSourceBranch
	m.request.GitRef = selected
	for _, branch := range m.data.Sources.Branches {
		if branch.Name == selected && m.request.HeadSha == "" {
			m.request.HeadSha = branch.Sha
			break
		}
	}
}

func (m *model) fillExplicitSourceSHA() {
	if m.request.HeadSha != "" || m.data.Sources == nil {
		return
	}
	if m.request.Source == models.AppAppBranchRunPreviewSourcePr {
		for _, pr := range m.data.Sources.PullRequests {
			if pr.PrNumber == m.request.PrNumber {
				m.request.HeadSha = pr.HeadSha
				return
			}
		}
	}
	if m.request.Source == models.AppAppBranchRunPreviewSourceBranch {
		for _, branch := range m.data.Sources.Branches {
			if branch.Name == m.request.GitRef {
				m.request.HeadSha = branch.Sha
				return
			}
		}
	}
}

func (m *model) defaultInstallID() string {
	cfg := m.data.PreviewConfig
	if cfg.InstallID != "" {
		for _, install := range m.data.Installs {
			if install.ID == cfg.InstallID {
				return install.ID
			}
		}
	}
	if cfg.InstallName != "" {
		for _, install := range m.data.Installs {
			if install.Name == cfg.InstallName {
				return install.ID
			}
		}
	}
	return ""
}

func (m *model) goBack() bool {
	switch m.step {
	case stepBranch:
		if m.err != nil {
			m.err = nil
			m.setBranchStep()
			return true
		}
	case stepMode:
		if m.opts.BranchID == "" {
			m.setBranchStep()
			return true
		}
	case stepSourceScope:
		if m.opts.Mode == "" {
			m.setModeStep()
			return true
		}
	case stepSourceKind:
		m.setSourceScopeStep()
		return true
	case stepSource:
		m.setSourceKindStep()
		return true
	case stepInstall:
		if m.request.Source != "" && m.opts.PRNumber == nil && m.opts.GitRef == "" {
			m.setSourceScopeStep()
			return true
		}
		if m.opts.Mode == "" {
			m.setModeStep()
			return true
		}
	}
	return false
}

func (m *model) finish() {
	m.result.Request = &m.request
}

func (m model) View() tea.View {
	v := tea.NewView(m.viewContent())
	v.AltScreen = true
	return v
}

func (m model) viewContent() string {
	if m.cancelled || m.result.Request != nil {
		return ""
	}

	width := min(max(m.width-8, 48), 88)
	if m.loading {
		return lipgloss.NewStyle().Padding(2, 4).Render(fmt.Sprintf("%s Loading preview configuration", m.spinner.View()))
	}

	var b strings.Builder
	b.WriteString(styles.TextBold.Render("Configure app branch preview"))
	b.WriteString("\n")
	b.WriteString(styles.TextDim.Render(m.progress()))
	b.WriteString("\n\n")
	if m.data != nil {
		b.WriteString(styles.TextDim.Render("Branch: " + m.data.BranchName))
		b.WriteString("\n\n")
	}
	if m.err != nil {
		b.WriteString(styles.TextError.Render(m.err.Error()))
		b.WriteString("\n\n")
		b.WriteString(styles.TextDim.Render("Press Esc to go back or q to cancel"))
		return lipgloss.NewStyle().Width(width).Padding(1, 2).Render(b.String())
	}

	b.WriteString(styles.TextBold.Render(m.heading()))
	b.WriteString("\n\n")
	if len(m.items) == 0 {
		b.WriteString(styles.TextError.Render(m.emptyMessage()))
	} else {
		for i, option := range m.items {
			prefix := "  "
			lineStyle := styles.TextPrimary
			if i == m.cursor {
				prefix = "▶ "
				lineStyle = styles.TextAccent.Bold(true)
			}
			line := prefix + option.title
			if option.description != "" {
				line += "  " + styles.TextDim.Render(option.description)
			}
			b.WriteString(lineStyle.Render(line))
			b.WriteString("\n")
		}
	}
	b.WriteString("\n")
	b.WriteString(styles.TextDim.Render("↑/↓ navigate · Enter select · Esc back · q quit"))
	return lipgloss.NewStyle().Width(width).Padding(1, 2).Render(b.String())
}

func (m model) progress() string {
	parts := []string{"App branch", "Mode", "Source", "Install"}
	active := 0
	switch m.step {
	case stepMode:
		active = 1
	case stepSourceScope, stepSourceKind, stepSource:
		active = 2
	case stepInstall:
		active = 3
	}
	if m.request.Mode == models.AppAppBranchRunPreviewModeBuildDashOnly && active == 3 {
		parts = parts[:3]
	}
	for i := range parts {
		if i == active {
			parts[i] = "[" + parts[i] + "]"
		}
	}
	return strings.Join(parts, "  →  ")
}

func (m model) heading() string {
	switch m.step {
	case stepBranch:
		return "Select an app branch"
	case stepMode:
		return "Select preview mode"
	case stepSourceScope:
		return "Select preview source"
	case stepSourceKind:
		return "Select source type"
	case stepSource:
		return "Select a pull request or branch"
	case stepInstall:
		return "Select an installation"
	default:
		return "Configure preview"
	}
}

func (m model) emptyMessage() string {
	switch m.step {
	case stepSourceKind, stepSource:
		return "No pull requests or git branches are available for preview"
	case stepInstall:
		return "No installs found for this app"
	default:
		return "No options available"
	}
}

func currentRequest(currentRef string, sources *models.HelpersListPreviewSourcesResult, headSHA string) *models.ServicePreviewRunRequest {
	if currentRef == "" || sources == nil {
		return nil
	}
	for _, pr := range sources.PullRequests {
		if pr.HeadRef == currentRef {
			sha := headSHA
			if sha == "" {
				sha = pr.HeadSha
			}
			return &models.ServicePreviewRunRequest{
				Source:   models.AppAppBranchRunPreviewSourcePr,
				PrNumber: pr.PrNumber,
				HeadSha:  sha,
			}
		}
	}
	for _, branch := range sources.Branches {
		if branch.Name == currentRef {
			sha := headSHA
			if sha == "" {
				sha = branch.Sha
			}
			return &models.ServicePreviewRunRequest{
				Source:  models.AppAppBranchRunPreviewSourceBranch,
				GitRef:  currentRef,
				HeadSha: sha,
			}
		}
	}
	return nil
}

func sourceDescription(req *models.ServicePreviewRunRequest) string {
	if req.Source == models.AppAppBranchRunPreviewSourcePr {
		return fmt.Sprintf("PR #%d", req.PrNumber)
	}
	return "Branch " + req.GitRef
}

func App(ctx context.Context, branches []Branch, loadBranch LoadBranchFunc, opts Options) (*Result, error) {
	m := initialModel(ctx, branches, loadBranch, opts)
	finalModel, err := ui.NewProgram(m, true).Run()
	if err != nil {
		return nil, fmt.Errorf("run preview wizard: %w", err)
	}
	fm, ok := finalModel.(model)
	if !ok || fm.cancelled {
		return nil, fmt.Errorf("preview cancelled")
	}
	if fm.err != nil {
		return nil, fm.err
	}
	if fm.result.Request == nil {
		return nil, fmt.Errorf("preview configuration incomplete")
	}
	return &fm.result, nil
}
