import { afterEach, expect, mock, test } from 'bun:test'
import { cleanup, fireEvent, render, screen, within } from '@testing-library/react'
import type { DiffSectionData } from '@/components/approvals/plan-diffs/app-config/AppConfigDiff'
import {
  DashboardPreferencesContext,
  type IDashboardPreferencesContext,
} from '@/providers/dashboard-preferences-provider'
import { ConfigChangesViewer } from './ConfigChangesViewer'

afterEach(cleanup)

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
  planSections: 'expanded',
  setPlanSections: () => {},
  resetPreferences: () => {},
}

const sections: DiffSectionData[] = [
  {
    name: 'Components',
    sectionKey: 'components',
    additions: 1,
    removals: 0,
    changed: 0,
    grouped: true,
    fields: [],
    entities: [
      {
        name: 'cache',
        op: 'add',
        componentType: 'helm_chart',
        fields: [{ key: 'chart_name', op: 'add', diff: "'cache'" }],
      },
    ],
  },
]

const renderViewer = (props?: Partial<{ sections: DiffSectionData[] }>) =>
  render(
    <DashboardPreferencesContext.Provider value={preferences}>
      <ConfigChangesViewer sections={props?.sections ?? sections} />
    </DashboardPreferencesContext.Provider>
  )

test('keeps the outline closed until it is shown', () => {
  const { container } = renderViewer()

  expect(screen.queryByRole('navigation', { name: 'Changes' })).toBeNull()
  expect(container.querySelector('[class*="grid-cols-"]')).toBeNull()

  const show = screen.getByRole('button', { name: 'Show outline' })
  expect(show.getAttribute('aria-expanded')).toBe('false')

  fireEvent.click(show)

  const hide = screen.getByRole('button', { name: 'Hide outline' })
  const nav = screen.getByRole('navigation', { name: 'Changes' })
  expect(hide.getAttribute('aria-expanded')).toBe('true')
  expect(hide.getAttribute('aria-controls')).toBe(nav.id)
  expect(nav.parentElement?.className).toContain(
    'md:grid-cols-[15rem_minmax(0,1fr)]'
  )

  fireEvent.click(hide)

  expect(screen.queryByRole('navigation', { name: 'Changes' })).toBeNull()
  expect(screen.getByRole('button', { name: 'Show outline' })).toBeTruthy()
})

test('scrolls to a section chosen from the outline', () => {
  const scrollIntoView = mock(() => {})
  const original = HTMLElement.prototype.scrollIntoView
  HTMLElement.prototype.scrollIntoView = scrollIntoView

  try {
    renderViewer()
    fireEvent.click(screen.getByRole('button', { name: 'Show outline' }))

    const nav = screen.getByRole('navigation', { name: 'Changes' })
    fireEvent.click(within(nav).getByRole('button', { name: /cache/ }))

    const section = document.querySelector(
      '[data-config-change="components/cache"]'
    )
    expect(section?.className).toContain('outline-primary-400/60')
    expect(scrollIntoView).toHaveBeenCalled()
  } finally {
    HTMLElement.prototype.scrollIntoView = original
  }
})

test('does not offer an outline when there are no changes', () => {
  renderViewer({ sections: [] })

  expect(screen.getByText('No changes')).toBeTruthy()
  expect(screen.queryByRole('button', { name: 'Show outline' })).toBeNull()
})
