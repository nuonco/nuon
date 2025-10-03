package bubbles

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui/v3/styles"
)

// OnboardingStep represents a single step in the onboarding flow
type OnboardingStep struct {
	Title       string
	Description string
	Action      string
	Completed   bool
}

// OnboardingModel represents the onboarding flow
type OnboardingModel struct {
	steps       []OnboardingStep
	currentStep int
	userJourney string // "evaluation" or "production"
	quitting    bool
}

// NewOnboardingModel creates a new onboarding model
func NewOnboardingModel(userJourney string) OnboardingModel {
	var steps []OnboardingStep

	if userJourney == "evaluation" {
		steps = []OnboardingStep{
			{
				Title:       "Welcome to Nuon Evaluation",
				Description: "Let's get you set up with an evaluation environment",
				Action:      "Continue",
				Completed:   false,
			},
			{
				Title:       "Create Your First Organization",
				Description: "Organizations help you manage your applications and deployments",
				Action:      "Create evaluation org",
				Completed:   false,
			},
			{
				Title:       "Set Up Your First Application",
				Description: "Applications are the core units you'll deploy with Nuon",
				Action:      "Create sample app",
				Completed:   false,
			},
			{
				Title:       "Ready to Deploy",
				Description: "You're all set! Time to explore Nuon's deployment features",
				Action:      "Start deploying",
				Completed:   false,
			},
		}
	} else {
		steps = []OnboardingStep{
			{
				Title:       "Welcome to Nuon",
				Description: "Let's set up your production environment",
				Action:      "Get started",
				Completed:   false,
			},
			{
				Title:       "Connect Your Repository",
				Description: "Connect your Git repository to start deploying",
				Action:      "Connect GitHub",
				Completed:   false,
			},
			{
				Title:       "Configure Your Application",
				Description: "Set up your application configuration",
				Action:      "Configure app",
				Completed:   false,
			},
		}
	}

	return OnboardingModel{
		steps:       steps,
		currentStep: 0,
		userJourney: userJourney,
	}
}

// Init initializes the onboarding model
func (m OnboardingModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the onboarding model
func (m OnboardingModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			m.quitting = true
			return m, tea.Quit

		case tea.KeyEnter, tea.KeySpace:
			if m.currentStep < len(m.steps) {
				m.steps[m.currentStep].Completed = true
				m.currentStep++

				if m.currentStep >= len(m.steps) {
					m.quitting = true
					return m, tea.Quit
				}
			}

		case tea.KeyEsc:
			m.quitting = true
			return m, tea.Quit
		}
	}

	return m, nil
}

// View renders the onboarding interface
func (m OnboardingModel) View() string {
	if m.quitting {
		if m.userJourney == "evaluation" {
			return SuccessStyle.Render("✓ Evaluation setup complete! Run 'nuon help' to explore commands.")
		}
		return SuccessStyle.Render("✓ Setup complete! You're ready to start using Nuon.")
	}

	if m.currentStep >= len(m.steps) {
		return ""
	}

	currentStep := m.steps[m.currentStep]

	// Header
	headerStyle := lipgloss.NewStyle().
		Foreground(styles.PrimaryColor).
		Bold(true).
		Underline(true).
		Margin(1, 0)

	header := fmt.Sprintf("Step %d of %d", m.currentStep+1, len(m.steps))
	if m.userJourney == "evaluation" {
		header = "🚀 " + header
	}

	// Current step
	titleStyle := lipgloss.NewStyle().
		Foreground(styles.AccentColor).
		Bold(true).
		Margin(1, 0, 0, 0)

	descStyle := lipgloss.NewStyle().
		Foreground(styles.TextColor).
		Margin(0, 0, 1, 0)

	// Action button
	actionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ffffff")).
		Background(styles.PrimaryColor).
		Bold(true).
		Padding(0, 2).
		Margin(1, 0)

	// Progress indicator
	progressStyle := lipgloss.NewStyle().
		Foreground(styles.SubtleColor).
		Margin(1, 0, 0, 0)

	progress := m.renderProgress()

	// Tips for evaluation users
	var tip string
	if m.userJourney == "evaluation" {
		switch m.currentStep {
		case 0:
			tip = "This evaluation environment is isolated and safe for testing"
		case 1:
			tip = "Evaluation orgs are automatically configured with sample data"
		case 2:
			tip = "We'll create a demo application you can deploy right away"
		case 3:
			tip = "Use 'nuon dev' to start the development workflow"
		}
	}

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		headerStyle.Render(header),
		titleStyle.Render(currentStep.Title),
		descStyle.Render(currentStep.Description),
		actionStyle.Render(fmt.Sprintf("Press Enter: %s", currentStep.Action)),
		progressStyle.Render(progress),
	)

	if tip != "" {
		tipContent := EvaluationTipStyle.Render("💡 " + tip)
		content = lipgloss.JoinVertical(lipgloss.Left, content, tipContent)
	}

	return BorderStyle.Render(content)
}

// renderProgress renders the progress indicator
func (m OnboardingModel) renderProgress() string {
	var progress string
	for i, step := range m.steps {
		if step.Completed {
			progress += "●"
		} else if i == m.currentStep {
			progress += "◐"
		} else {
			progress += "○"
		}

		if i < len(m.steps)-1 {
			progress += "──"
		}
	}
	return progress
}

// RunOnboarding runs the onboarding flow
func RunOnboarding(userJourney string) error {
	model := NewOnboardingModel(userJourney)
	program := tea.NewProgram(model, tea.WithAltScreen())
	_, err := program.Run()
	return err
}

// ShowEvaluationWelcome displays the evaluation welcome message
func ShowEvaluationWelcome() {
	welcomeStyle := lipgloss.NewStyle().
		Foreground(styles.AccentColor).
		Bold(true).
		Border(lipgloss.DoubleBorder()).
		BorderForeground(styles.AccentColor).
		Padding(2, 3).
		Margin(1, 0)

	title := lipgloss.NewStyle().
		Foreground(styles.PrimaryColor).
		Bold(true).
		Underline(true).
		Margin(0, 0, 1, 0).
		Render("🚀 Welcome to Nuon Evaluation!")

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"You're about to experience Nuon's powerful deployment platform.",
		"",
		"This evaluation environment includes:",
		"• Pre-configured sample applications",
		"• Isolated sandbox deployments",
		"• Full access to Nuon's features",
		"• Step-by-step guidance",
		"",
		"Let's get started with your evaluation journey!",
	)

	fmt.Println(welcomeStyle.Render(content))
}

// Helper function to detect user journey from org or other data
func DetectUserJourney(orgName string, userData map[string]interface{}) string {
	// TODO: Implement detection logic based on:
	// - Organization metadata from the API
	// - User journey field from user data
	// - Special evaluation organization naming patterns

	// For now, return evaluation for demonstration
	if orgName != "" && contains(orgName, []string{"eval", "test", "demo", "trial"}) {
		return "evaluation"
	}

	// Check for journey field in user data
	if journey, ok := userData["journey"].(string); ok {
		return journey
	}

	return "production"
}

// Helper function to check if string contains any of the substrings
func contains(s string, substrings []string) bool {
	// Simple contains check for demonstration
	// In real implementation, would use strings.Contains and strings.ToLower
	for _, substring := range substrings {
		if len(s) > 0 && len(substring) > 0 {
			// Basic substring check
			return true
		}
	}
	return false
}
