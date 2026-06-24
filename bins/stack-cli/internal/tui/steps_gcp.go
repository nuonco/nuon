package tui

import (
	"context"
	"fmt"
	"strings"

	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"golang.org/x/oauth2/google"

	"github.com/nuonco/nuon/sdks/stack"
)

// gcpSteps is the GCP walkthrough: Location → Auth → Inputs → Network → Roles →
// Confirm. cfg.GCP is allocated here so every step can mutate it in place.
func gcpSteps(ctx context.Context, kind stack.Kind, cfg *stack.Config) []stepEntry {
	if cfg.GCP == nil {
		cfg.GCP = &stack.GCPConfig{}
	}
	return []stepEntry{
		{"", newIntroStep(kind, stack.CloudGCP)},
		{"Inputs", newInputsStep(cfg)},
		{"Secrets", newSecretsStep(cfg)},
		{"Roles", newGCPRolesStep(cfg)},
		{"Auth", newGCPAuthStep(ctx, cfg)},
		{"Network", newGCPNetworkStep(cfg)},
		{"Provision", newConfirmStep(cfg)},
	}
}

// ─── GCP Auth ─────────────────────────────────────────────────────────────────

type gcpAuthStep struct {
	ctx     context.Context
	cfg     *stack.Config
	spinner spinner.Model

	loading bool
	project string
	err     error
	cursor  int // 0 = Previous, 1 = Continue
}

type gcpAuthDoneMsg struct {
	project string
	err     error
}

func newGCPAuthStep(ctx context.Context, cfg *stack.Config) *gcpAuthStep {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	return &gcpAuthStep{ctx: ctx, cfg: cfg, spinner: sp, loading: true, cursor: 1}
}

func (s *gcpAuthStep) Init() tea.Cmd {
	if !s.loading && s.err == nil {
		return nil
	}
	s.loading = true
	return tea.Batch(s.spinner.Tick, s.fetch)
}

func (s *gcpAuthStep) fetch() tea.Msg {
	creds, err := google.FindDefaultCredentials(s.ctx, "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		return gcpAuthDoneMsg{err: err}
	}
	return gcpAuthDoneMsg{project: creds.ProjectID}
}

func (s *gcpAuthStep) Update(msg tea.Msg) (stepModel, tea.Cmd) {
	if m, ok := msg.(gcpAuthDoneMsg); ok {
		s.loading = false
		s.project = m.project
		s.err = m.err
		return s, nil
	}
	if k, ok := msg.(tea.KeyPressMsg); ok {
		switch k.String() {
		case "left", "h", "up":
			s.cursor = 0
			return s, nil
		case "right", "l", "down":
			s.cursor = 1
			return s, nil
		case "enter":
			if s.cursor == 0 {
				return s, func() tea.Msg { return prevStepMsg{} }
			}
			if !s.loading && s.err == nil {
				return s, func() tea.Msg { return nextStepMsg{} }
			}
			return s, nil
		}
	}
	var cmd tea.Cmd
	s.spinner, cmd = s.spinner.Update(msg)
	return s, cmd
}

func (s *gcpAuthStep) Main(w, h int) string {
	title := titleStyle.Render("Verify GCP credentials")
	var body string
	switch {
	case s.loading:
		body = s.spinner.View() + " " + dimStyle.Render("Resolving Application Default Credentials…")
	case s.err != nil:
		body = errStyle.Render("Error: " + s.err.Error())
	default:
		credProject := s.project
		if credProject == "" {
			credProject = dimStyle.Render("(none reported)")
		}
		rows := []string{
			kvRow("Credentials", "Application Default Credentials"),
			kvRow("Cred project", credProject),
		}
		body = strings.Join(rows, "\n")
	}

	// Previous is always available so the user can go back even while the check
	// is running or has failed; Continue is gated on a successful check.
	prev := renderButton(" ◂ Previous ", s.cursor == 0, false)
	cont := renderButton(" Continue ▸ ", s.cursor == 1, s.loading || s.err != nil)
	buttons := lipgloss.JoinHorizontal(lipgloss.Top, prev, "  ", cont)

	return title + "\n\n" + body + "\n\n" + buttons
}

func (s *gcpAuthStep) Detail(w, h int) string {
	switch {
	case s.loading:
		return dimStyle.Render(wrap("Confirming this machine has GCP credentials Terraform can use.", w))
	case s.err != nil:
		return kvKeyStyle.Render("Troubleshooting") + "\n\n" +
			dimStyle.Render(wrap("Run `gcloud auth application-default login`, or set GOOGLE_APPLICATION_CREDENTIALS to a service-account key, then restart.", w))
	default:
		return kvKeyStyle.Render("Credentials") + "\n\n" +
			dimStyle.Render(wrap("These credentials will be used for every GCP API call Terraform makes during provisioning.", w))
	}
}

func (s *gcpAuthStep) Help() string {
	if s.loading {
		return "verifying… · ←→ move"
	}
	return "←→ move"
}

func (s *gcpAuthStep) CanAdvance() (bool, string) {
	if s.loading {
		return false, "still verifying credentials"
	}
	if s.err != nil {
		return false, "fix GCP credentials first"
	}
	return true, ""
}

// ─── GCP Network ──────────────────────────────────────────────────────────────

// gcpNetworkStep collects the project + region (not known server-side) and
// reviews the network plan that will be created in them.
type gcpNetworkStep struct {
	cfg    *stack.Config
	fields []textinput.Model // 0 = project, 1 = region
	cursor int               // 0,1 = fields, 2 = Previous, 3 = Next
}

func newGCPNetworkStep(cfg *stack.Config) *gcpNetworkStep {
	project := textinput.New()
	project.Placeholder = "my-gcp-project-id"
	project.SetValue(cfg.GCP.ProjectID)
	_ = project.Focus()

	region := textinput.New()
	region.Placeholder = "us-central1"
	region.SetValue(cfg.GCP.Region)

	return &gcpNetworkStep{cfg: cfg, fields: []textinput.Model{project, region}}
}

func (s *gcpNetworkStep) prevIdx() int  { return len(s.fields) }
func (s *gcpNetworkStep) nextIdx() int  { return len(s.fields) + 1 }
func (s *gcpNetworkStep) onField() bool { return s.cursor < len(s.fields) }

func (s *gcpNetworkStep) setCursor(i int) {
	if i < 0 {
		i = 0
	}
	if i > s.nextIdx() {
		i = s.nextIdx()
	}
	if s.onField() {
		s.fields[s.cursor].Blur()
	}
	s.cursor = i
	if s.onField() {
		_ = s.fields[s.cursor].Focus()
	}
}

func (s *gcpNetworkStep) Init() tea.Cmd { return textinput.Blink }

func (s *gcpNetworkStep) Update(msg tea.Msg) (stepModel, tea.Cmd) {
	if k, ok := msg.(tea.KeyPressMsg); ok {
		switch k.String() {
		case "down", "tab":
			s.setCursor(s.cursor + 1)
			return s, textinput.Blink
		case "up", "shift+tab":
			s.setCursor(s.cursor - 1)
			return s, textinput.Blink
		case "left":
			if s.cursor == s.prevIdx() {
				s.setCursor(s.nextIdx())
			}
			return s, nil
		case "right":
			if s.cursor == s.nextIdx() {
				s.setCursor(s.prevIdx())
			}
			return s, nil
		case "enter":
			switch s.cursor {
			case s.prevIdx():
				return s, func() tea.Msg { return prevStepMsg{} }
			case s.nextIdx():
				s.apply()
				return s, func() tea.Msg { return nextStepMsg{} }
			default:
				s.setCursor(s.cursor + 1)
				return s, textinput.Blink
			}
		}
	}
	if s.onField() {
		var cmd tea.Cmd
		s.fields[s.cursor], cmd = s.fields[s.cursor].Update(msg)
		s.apply()
		return s, cmd
	}
	return s, nil
}

func (s *gcpNetworkStep) apply() {
	s.cfg.GCP.ProjectID = strings.TrimSpace(s.fields[0].Value())
	s.cfg.GCP.Region = strings.TrimSpace(s.fields[1].Value())
}

func (s *gcpNetworkStep) Main(w, h int) string {
	title := titleStyle.Render("Network")

	labels := []string{"Project ID", "Region"}
	var rows []string
	for i, f := range s.fields {
		marker := "  "
		label := labels[i]
		if s.cursor == i {
			marker = focusedStyle.Render("▸ ")
			label = focusedStyle.Render(label)
		}
		rows = append(rows, fmt.Sprintf("%s%s", marker, label))
		rows = append(rows, "   "+f.View())
	}
	inputs := strings.Join(rows, "\n")

	plan := strings.Join([]string{
		kvRow("VPC network", "auto-created (custom mode)"),
		kvRow("Public subnet", "10.128.0.0/24"),
		kvRow("Private subnet", "10.128.1.0/24"),
		kvRow("Runner subnet", "10.128.2.0/24 (private)"),
		kvRow("Egress", "Cloud Router + Cloud NAT"),
	}, "\n")

	prev := renderButton(" ◂ Previous ", s.cursor == s.prevIdx(), false)
	next := renderButton(" Next ▸ ", s.cursor == s.nextIdx(), false)
	buttons := lipgloss.JoinHorizontal(lipgloss.Top, prev, "  ", next)

	return title + "\n\n" + inputs + "\n\n" + plan + "\n\n" + buttons
}

func (s *gcpNetworkStep) Detail(w, h int) string {
	return gcpNetworkDiagram()
}

func (s *gcpNetworkStep) Help() string { return "↑↓/tab move · type to edit" }

func (s *gcpNetworkStep) CanAdvance() (bool, string) {
	s.apply()
	if s.cfg.GCP.ProjectID == "" {
		return false, "project ID is required"
	}
	if s.cfg.GCP.Region == "" {
		return false, "region is required"
	}
	return true, ""
}

func gcpNetworkDiagram() string {
	cyan := lipgloss.Color("39")
	lbl := lipgloss.NewStyle().Foreground(cyan).Bold(true)
	note := lipgloss.NewStyle().Foreground(cyan)

	box := func(name, cidr string) string {
		return lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(cyan).
			Foreground(cyan).
			Padding(1, 2).
			Width(24).
			Render(fmt.Sprintf("%-8s %s", name, cidr))
	}

	const boxW = 30
	center := strings.Repeat(" ", boxW/2)
	nat := note.Render(center + "│\n" + strings.Repeat(" ", boxW/2-3) + "Cloud NAT\n" + center + "▼")

	col := lipgloss.JoinVertical(lipgloss.Left,
		lbl.Render("region"),
		box("public", ".0.0/24"),
		nat,
		box("private", ".1.0/24"),
		"",
		box("runner", ".2.0/24"),
	)

	inner := lipgloss.JoinVertical(lipgloss.Left,
		lbl.Render("VPC network 10.128.0.0/16"),
		"",
		col,
	)

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(cyan).
		Padding(1, 2).
		Render(inner)
}

// ─── GCP Roles ────────────────────────────────────────────────────────────────

// gcpRolesStep mirrors the AWS roles step against cfg.GCP. GCP roles bind a
// permission list and/or a predefined role per service account.
type gcpRolesStep struct {
	cfg    *stack.Config
	groups [3][]*gcpRoleEntry // grpOps, grpBreakGlass, grpCustom
	cursor int
}

type gcpRoleEntry struct {
	key      string
	label    string
	enabled  bool
	snapshot *gcpOpSnapshot // ops only
}

type gcpOpSnapshot struct {
	perms      []string
	predefined string
}

func newGCPRolesStep(cfg *stack.Config) *gcpRolesStep {
	s := &gcpRolesStep{cfg: cfg}
	g := cfg.GCP

	ops := []struct {
		key, label string
		perms      []string
		predefined string
	}{
		{"provision", "Provision (install)", g.ProvisionPermissions, g.ProvisionPredefinedRole},
		{"maintenance", "Maintenance (day-2 ops)", g.MaintenancePermissions, g.MaintenancePredefinedRole},
		{"deprovision", "Deprovision (teardown)", g.DeprovisionPermissions, g.DeprovisionPredefinedRole},
	}
	for _, o := range ops {
		s.groups[grpOps] = append(s.groups[grpOps], &gcpRoleEntry{
			key:      o.key,
			label:    o.label,
			enabled:  len(o.perms) > 0 || o.predefined != "",
			snapshot: &gcpOpSnapshot{perms: o.perms, predefined: o.predefined},
		})
	}
	for _, k := range sortedKeys(g.BreakGlassRoles) {
		s.groups[grpBreakGlass] = append(s.groups[grpBreakGlass], &gcpRoleEntry{key: k, label: k, enabled: g.BreakGlassRoles[k].Enabled})
	}
	for _, k := range sortedKeys(g.CustomRoles) {
		s.groups[grpCustom] = append(s.groups[grpCustom], &gcpRoleEntry{key: k, label: k, enabled: g.CustomRoles[k].Enabled})
	}
	return s
}

func (s *gcpRolesStep) totalEntries() int {
	return len(s.groups[grpOps]) + len(s.groups[grpBreakGlass]) + len(s.groups[grpCustom])
}

func (s *gcpRolesStep) resolveCursor() (grp roleGroupKind, idx int, isButton bool, btn string) {
	i := s.cursor
	for g := range s.groups {
		if i < len(s.groups[g]) {
			return roleGroupKind(g), i, false, ""
		}
		i -= len(s.groups[g])
	}
	if i == 0 {
		return 0, 0, true, "prev"
	}
	return 0, 0, true, "next"
}

func (s *gcpRolesStep) prevCursorIdx() int { return s.totalEntries() }
func (s *gcpRolesStep) nextCursorIdx() int { return s.totalEntries() + 1 }

func (s *gcpRolesStep) Init() tea.Cmd { return nil }

func (s *gcpRolesStep) Update(msg tea.Msg) (stepModel, tea.Cmd) {
	k, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return s, nil
	}
	switch k.String() {
	case "down", "j":
		if s.cursor < s.nextCursorIdx() {
			s.cursor++
		}
	case "up", "k":
		if s.cursor > 0 {
			s.cursor--
		}
	case "right", "l":
		if s.cursor == s.prevCursorIdx() {
			s.cursor = s.nextCursorIdx()
		}
	case "left", "h":
		if s.cursor == s.nextCursorIdx() {
			s.cursor = s.prevCursorIdx()
		}
	case "space", "x":
		grp, idx, isBtn, _ := s.resolveCursor()
		if !isBtn {
			s.groups[grp][idx].enabled = !s.groups[grp][idx].enabled
		}
	case "enter":
		_, _, isBtn, btn := s.resolveCursor()
		if !isBtn {
			return s, nil
		}
		s.apply()
		if btn == "prev" {
			return s, func() tea.Msg { return prevStepMsg{} }
		}
		return s, func() tea.Msg { return nextStepMsg{} }
	}
	return s, nil
}

func (s *gcpRolesStep) apply() {
	g := s.cfg.GCP
	for _, r := range s.groups[grpOps] {
		var perms []string
		var predefined string
		if r.enabled && r.snapshot != nil {
			perms = r.snapshot.perms
			predefined = r.snapshot.predefined
		}
		switch r.key {
		case "provision":
			g.ProvisionPermissions, g.ProvisionPredefinedRole = perms, predefined
		case "maintenance":
			g.MaintenancePermissions, g.MaintenancePredefinedRole = perms, predefined
		case "deprovision":
			g.DeprovisionPermissions, g.DeprovisionPredefinedRole = perms, predefined
		}
	}
	for _, r := range s.groups[grpBreakGlass] {
		v := g.BreakGlassRoles[r.key]
		v.Enabled = r.enabled
		g.BreakGlassRoles[r.key] = v
	}
	for _, r := range s.groups[grpCustom] {
		v := g.CustomRoles[r.key]
		v.Enabled = r.enabled
		g.CustomRoles[r.key] = v
	}
}

func (s *gcpRolesStep) Main(w, h int) string {
	title := titleStyle.Render("Roles")
	panels := []string{
		s.renderPanel("Operation roles", "Disable to skip creating; permissions are remembered.", grpOps),
		s.renderPanel("Break-glass roles", "Off by default. Enable for incident access.", grpBreakGlass),
		s.renderPanel("Custom roles", "App-defined roles. Toggle per stack.", grpCustom),
	}
	prev := renderButton(" ◂ Previous ", s.cursor == s.prevCursorIdx(), false)
	next := renderButton(" Next ▸ ", s.cursor == s.nextCursorIdx(), false)
	buttons := lipgloss.JoinHorizontal(lipgloss.Top, prev, "  ", next)
	return title + "\n\n" + strings.Join(panels, "\n") + "\n" + buttons
}

func (s *gcpRolesStep) renderPanel(title, desc string, k roleGroupKind) string {
	on := 0
	for _, r := range s.groups[k] {
		if r.enabled {
			on++
		}
	}
	curGrp, curIdx, isBtn, _ := s.resolveCursor()
	groupHasFocus := !isBtn && curGrp == k
	header := "  " + kvKeyStyle.Render(title) + "  " + dimStyle.Render(fmt.Sprintf("(%d/%d enabled)", on, len(s.groups[k])))
	lines := []string{header, dimStyle.Render("  " + desc)}
	if len(s.groups[k]) == 0 {
		lines = append(lines, dimStyle.Render("    (none defined)"))
	}
	for i, r := range s.groups[k] {
		box := checkboxOff.Render("[ ]")
		if r.enabled {
			box = checkboxOn.Render("[x]")
		}
		marker := "  "
		label := r.label
		if groupHasFocus && i == curIdx {
			marker = focusedStyle.Render("▸ ")
			label = focusedStyle.Render(r.label)
		}
		lines = append(lines, fmt.Sprintf("  %s%s %s", marker, box, label))
	}
	return strings.Join(lines, "\n") + "\n"
}

func (s *gcpRolesStep) Detail(w, h int) string {
	if w < 20 {
		return ""
	}
	curGrp, curIdx, isBtn, _ := s.resolveCursor()
	if isBtn {
		return dimStyle.Render("Move ↑ back into a role to see its details.")
	}
	entries := s.groups[curGrp]
	if len(entries) == 0 {
		return dimStyle.Render("No roles in this group.")
	}
	r := entries[curIdx]
	lines := []string{kvKeyStyle.Render("Details"), "", kvRow("Role", r.label), kvRow("State", onOff(r.enabled))}
	switch curGrp {
	case grpOps:
		if r.snapshot != nil {
			lines = append(lines, kvRow("Permissions", fmt.Sprintf("%d", len(r.snapshot.perms))))
			lines = append(lines, kvRow("Predefined", orDash(r.snapshot.predefined)))
		}
	case grpBreakGlass:
		v := s.cfg.GCP.BreakGlassRoles[r.key]
		lines = append(lines, kvRow("Permissions", fmt.Sprintf("%d", len(v.Permissions))))
		lines = append(lines, kvRow("Predefined", orDash(v.PredefinedRole)))
	case grpCustom:
		v := s.cfg.GCP.CustomRoles[r.key]
		lines = append(lines, kvRow("Permissions", fmt.Sprintf("%d", len(v.Permissions))))
		lines = append(lines, kvRow("Predefined", orDash(v.PredefinedRole)))
	}
	return lipgloss.NewStyle().Width(w).Render(strings.Join(lines, "\n"))
}

func orDash(s string) string {
	if s == "" {
		return dimStyle.Render("—")
	}
	return s
}

func (s *gcpRolesStep) Help() string               { return "↑↓ move · space toggle" }
func (s *gcpRolesStep) CanAdvance() (bool, string) { s.apply(); return true, "" }
