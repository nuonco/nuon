package creator

import (
	"testing"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func testGroup(id, name string, matchLabels map[string]string) *models.AppAppBranchInstallGroup {
	return &models.AppAppBranchInstallGroup{
		ID:   id,
		Name: name,
		LabelSelector: &models.GithubComNuoncoNuonPkgLabelsSelector{
			MatchLabels: matchLabels,
		},
	}
}

// validFormModel is a model whose name check has already cleared, so Enter is
// allowed to move past the form.
func validFormModel(groups []*models.AppAppBranchInstallGroup) model {
	input := textinput.New()
	input.SetValue("staging")

	return model{
		inputs:        []textinput.Model{input},
		inputMappings: []inputMapping{{name: "name", displayName: "Install name", required: true}},
		nameChecked:   "staging",
		groups:        groups,
		presetLabels:  map[string]string{},
		width:         minRequiredWidth,
		height:        minRequiredHeight,
		keys:          keys,
	}
}

func TestEnterAdvancesToGroupStepWhenGroupsExist(t *testing.T) {
	m := validFormModel([]*models.AppAppBranchInstallGroup{
		testGroup("group-a", "staging", map[string]string{"env": "staging"}),
	})

	next, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	updated := next.(model)

	assert.Equal(t, stepGroup, updated.step)
	assert.False(t, updated.submitting)
	assert.Nil(t, cmd)
}

func TestEnterSubmitsDirectlyWhenNoGroups(t *testing.T) {
	m := validFormModel(nil)

	next, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	updated := next.(model)

	assert.Equal(t, stepForm, updated.step)
	assert.True(t, updated.submitting)
	assert.NotNil(t, cmd)
}

func TestGroupSelectionMergesLabels(t *testing.T) {
	m := validFormModel([]*models.AppAppBranchInstallGroup{
		testGroup("group-a", "staging", map[string]string{"env": "staging"}),
	})
	m.presetLabels = map[string]string{"owner": "team-a"}
	m.step = stepGroup

	next, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	updated := next.(model)

	assert.True(t, updated.submitting)
	assert.Equal(t, map[string]string{"owner": "team-a", "env": "staging"}, updated.presetLabels)
}

func TestGroupSelectionRejectsConflictingLabel(t *testing.T) {
	m := validFormModel([]*models.AppAppBranchInstallGroup{
		testGroup("group-a", "staging", map[string]string{"env": "staging"}),
	})
	m.presetLabels = map[string]string{"env": "prod"}
	m.step = stepGroup

	next, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	updated := next.(model)

	assert.False(t, updated.submitting)
	assert.Equal(t, stepGroup, updated.step)
	assert.Equal(t, "error", updated.status.Level)
	assert.Equal(t, map[string]string{"env": "prod"}, updated.presetLabels)
	assert.Nil(t, cmd)
}

func TestSkipRowLeavesLabelsUnchanged(t *testing.T) {
	groups := []*models.AppAppBranchInstallGroup{
		testGroup("group-a", "staging", map[string]string{"env": "staging"}),
	}
	m := validFormModel(groups)
	m.presetLabels = map[string]string{"owner": "team-a"}
	m.step = stepGroup

	// Move past the last group onto the skip row.
	next, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	moved := next.(model)
	require.Equal(t, len(groups), moved.groupIndex)
	require.Nil(t, moved.selectedGroup())

	next, _ = moved.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	updated := next.(model)

	assert.True(t, updated.submitting)
	assert.Equal(t, map[string]string{"owner": "team-a"}, updated.presetLabels)
}

func TestShiftTabReturnsToForm(t *testing.T) {
	m := validFormModel([]*models.AppAppBranchInstallGroup{
		testGroup("group-a", "staging", map[string]string{"env": "staging"}),
	})
	m.step = stepGroup

	next, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift})
	updated := next.(model)

	assert.Equal(t, stepForm, updated.step)
}
