import { afterEach, expect, test } from 'bun:test'
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { noOpAndReadResourcesPlan } from '@/lib/fixtures/plan-diffs/terraform'
import {
  DashboardPreferencesContext,
  type IDashboardPreferencesContext,
} from '@/providers/dashboard-preferences-provider'
import { TerraformDiff } from './TerraformDiff'

const preferences: IDashboardPreferencesContext = {
  showIds: false,
  setShowIds: () => {},
  isStatusBarEnabled: false,
  setIsStatusBarEnabled: () => {},
  isInstallsTabEnabled: false,
  setIsInstallsTabEnabled: () => {},
  diffViewer: 'v2',
  setDiffViewer: () => {},
  diffWrap: 'scroll',
  setDiffWrap: () => {},
  diffView: 'unified',
  setDiffView: () => {},
  planSections: 'collapsed',
  setPlanSections: () => {},
  resetPreferences: () => {},
}

afterEach(cleanup)

test('hides no-op terraform changes until show no-ops is on', () => {
  render(
    <DashboardPreferencesContext.Provider value={preferences}>
      <TerraformDiff plan={noOpAndReadResourcesPlan} />
    </DashboardPreferencesContext.Provider>
  )

  const toggle = screen.getByRole('switch', { name: 'Show no-ops' })
  expect(toggle.getAttribute('aria-checked')).toBe('false')
  expect(screen.queryByText('aws_s3_bucket.logs')).toBeNull()

  fireEvent.click(toggle)

  expect(screen.getByText('aws_s3_bucket.logs')).toBeTruthy()
})
