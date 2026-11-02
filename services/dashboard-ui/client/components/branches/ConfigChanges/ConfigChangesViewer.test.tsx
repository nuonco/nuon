import type { ComponentProps } from 'react'
import { afterEach, expect, mock, test } from 'bun:test'
import { cleanup, fireEvent, render, screen, within } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import type { DiffSectionData } from '@/components/approvals/plan-diffs/app-config/AppConfigDiff'
import {
  DashboardPreferencesContext,
  type IDashboardPreferencesContext,
} from '@/providers/dashboard-preferences-provider'
import { configChangesBaseline } from './ConfigChangesBaseline'
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

const renderViewer = (props?: Partial<ComponentProps<typeof ConfigChangesViewer>>) =>
  render(
    <MemoryRouter>
      <DashboardPreferencesContext.Provider value={preferences}>
        <ConfigChangesViewer {...props} sections={props?.sections ?? sections} />
      </DashboardPreferencesContext.Provider>
    </MemoryRouter>
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

test('names the run the diff is compared to', () => {
  renderViewer({
    baseline: {
      run: {
        href: '/org-1/apps/app-1/branches/br-1/runs/wf-1',
        status: 'success',
        message: 'Widen the acme VPC CIDR range',
        sha: '9f8e7d6c5b4a',
      },
    },
  })

  const link = screen.getByRole('link', {
    name: 'Widen the acme VPC CIDR range',
  })
  expect(link.getAttribute('href')).toBe(
    '/org-1/apps/app-1/branches/br-1/runs/wf-1'
  )
  expect(screen.queryByText("This run hasn't finished.")).toBeNull()
})

test('says when the compared run has not finished', () => {
  renderViewer({
    baseline: {
      run: { status: 'running', message: 'Bump the cache chart' },
    },
  })

  expect(screen.getByText('Bump the cache chart')).toBeTruthy()
  expect(screen.getByText('Running')).toBeTruthy()
  expect(screen.getByText("This run hasn't finished.")).toBeTruthy()
})

test('says when there is no previous run', () => {
  renderViewer({ baseline: { run: null } })

  expect(
    screen.getByText(
      'No previous run to compare against. This shows the full config as created.'
    )
  ).toBeTruthy()
})

test('renders the diff when the baseline is missing', () => {
  renderViewer()
  renderViewer({ baseline: { run: {} } })
  renderViewer({ baseline: null, sections: [] })

  expect(screen.getAllByText('cache').length).toBeGreaterThan(0)
  expect(screen.queryByText(/Compared to/)).toBeNull()
  expect(screen.queryByText(/No previous run/)).toBeNull()
})

test('leaves the baseline unset until a comparison exists', () => {
  expect(configChangesBaseline(undefined, { orgId: 'org-1' })).toBeUndefined()
  expect(configChangesBaseline(null)).toBeUndefined()
  expect(configChangesBaseline({})).toEqual({ run: null })
  expect(
    configChangesBaseline(
      {
        base_run: {
          id: 'run-1',
          status: 'pending',
          workflow_id: 'wf-1',
          vcs_connection_commit: {
            sha: 'abcdef1234567890',
            message: 'Bump the cache chart\n\nMore detail.',
          },
        },
      },
      { orgId: 'org-1', appId: 'app-1', branchId: 'br-1' }
    )
  ).toEqual({
    run: {
      href: '/org-1/apps/app-1/branches/br-1/runs/wf-1',
      status: 'pending',
      message: 'Bump the cache chart',
      sha: 'abcdef1234567890',
    },
  })
  expect(
    configChangesBaseline({
      base_run: { id: 'run-1', status: 'success' },
    })
  ).toEqual({
    run: {
      href: undefined,
      status: 'success',
      message: undefined,
      sha: undefined,
    },
  })
})
