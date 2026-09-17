import { afterEach, describe, expect, jest, mock, test } from 'bun:test'
import { act, cleanup, fireEvent, render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import type { TWorkflow } from '@/types/ctl-api.types'
import { WorkflowTimeline, type IWorkflowFilter } from './WorkflowTimeline'

afterEach(cleanup)

const filter = (label: string): IWorkflowFilter<string> => ({
  label,
  options: [
    { value: 'failed', label: 'Failed' },
    { value: 'succeeded', label: 'Succeeded' },
  ],
  selected: new Set<string>(),
  onToggle: () => {},
  onIsolate: () => {},
  onReset: () => {},
})

const run = (workflow: Partial<TWorkflow>): TWorkflow =>
  ({
    owner_type: 'app_branches',
    type: 'app_branches_manual_update',
    name: 'Run',
    status: { status: 'success' },
    ...workflow,
  }) as TWorkflow

const branchRuns: TWorkflow[] = [
  run({
    id: 'wflq7fplr1up5atx5zpxotbab1',
    name: 'PR #128',
    created_at: '2026-09-10T12:00:00Z',
  }),
  run({
    id: 'wflq7fplr1up5atx5zpxotbab2',
    name: 'VCS push',
    created_at: '2026-09-09T09:30:00Z',
    status: { status: 'error' },
  }),
]

const installRuns: TWorkflow[] = [
  run({
    id: 'wflq7fplr1up5atx5zpxotbab3',
    owner_type: 'installs',
    type: 'drift_run',
    name: 'Deploying to install (api)',
    plan_only: true,
    created_at: '2026-09-10T12:00:00Z',
  }),
]

const base = {
  search: '',
  onSearchChange: () => {},
  offset: 0,
  pageSize: 20,
  hasNext: false,
  onOffsetChange: () => {},
  statusFilter: filter('Status'),
  typeFilter: filter('Type'),
  previewFilter: filter('Preview'),
  dateFilter: filter('Date'),
}

describe('rows', () => {
  test('renders one row per run for a branch owner', () => {
    render(<WorkflowTimeline {...base} workflows={branchRuns} />)

    expect(screen.getByText('PR #128')).toBeTruthy()
    expect(screen.getByText('VCS push')).toBeTruthy()
    expect(screen.getAllByRole('listitem')).toHaveLength(2)
  })

  test('groups rows by day', () => {
    render(<WorkflowTimeline {...base} workflows={branchRuns} />)

    expect(screen.getAllByRole('list')).toHaveLength(2)
  })

  test('renders install workflows with their own markers', () => {
    render(<WorkflowTimeline {...base} workflows={installRuns} />)

    expect(screen.getByText('drift scan')).toBeTruthy()
    expect(screen.queryByText('preview')).toBeNull()
  })

  test('renders the drift-detected marker only for drifted runs', () => {
    const { rerender } = render(
      <WorkflowTimeline {...base} workflows={installRuns} />
    )
    expect(screen.queryByText('drift detected')).toBeNull()

    rerender(
      <WorkflowTimeline
        {...base}
        workflows={installRuns}
        driftedWorkflowIds={new Set(['wflq7fplr1up5atx5zpxotbab3'])}
      />
    )
    expect(screen.getByText('drift detected')).toBeTruthy()
  })

  test('renders rows unlinked when no href builder is given', () => {
    render(<WorkflowTimeline {...base} workflows={branchRuns} />)

    expect(screen.queryByRole('link', { name: 'PR #128' })).toBeNull()
  })

  test('links rows when an href builder is given', () => {
    render(
      <MemoryRouter>
        <WorkflowTimeline
          {...base}
          workflows={branchRuns}
          getWorkflowHref={(workflow) => `/runs/${workflow.id}`}
        />
      </MemoryRouter>
    )

    expect(
      screen.getByRole('link', { name: 'PR #128' }).getAttribute('href')
    ).toBe('/runs/wflq7fplr1up5atx5zpxotbab1')
  })
})

describe('states', () => {
  test('empty, no-matches and failed render distinct content', () => {
    const { rerender } = render(
      <WorkflowTimeline {...base} workflows={[]} />
    )
    expect(screen.getByText('No activity yet')).toBeTruthy()

    rerender(<WorkflowTimeline {...base} workflows={[]} filtered />)
    expect(screen.getByText('No runs match these filters')).toBeTruthy()

    rerender(
      <WorkflowTimeline {...base} workflows={[]} error={new Error('boom')} />
    )
    expect(screen.getByText('Activity failed to load')).toBeTruthy()
  })

  test('loading announces itself and renders no rows', () => {
    render(<WorkflowTimeline {...base} workflows={[]} loading />)

    expect(screen.getByRole('status', { name: 'Loading activity' })).toBeTruthy()
    expect(screen.queryByText('No activity yet')).toBeNull()
  })
})

describe('controls', () => {
  test('search changes call the handler', () => {
    jest.useFakeTimers()
    const onSearchChange = mock()
    render(
      <WorkflowTimeline
        {...base}
        workflows={branchRuns}
        onSearchChange={onSearchChange}
      />
    )

    fireEvent.change(screen.getByRole('searchbox', { name: 'Search activity' }), {
      target: { value: 'pr' },
    })
    act(() => jest.advanceTimersByTime(300))

    expect(onSearchChange).toHaveBeenCalledWith('pr')
    jest.useRealTimers()
  })

  test('every filter axis renders a control', () => {
    render(<WorkflowTimeline {...base} workflows={branchRuns} />)

    expect(screen.getByRole('button', { name: /Status/ })).toBeTruthy()
    expect(screen.getByRole('button', { name: /Type/ })).toBeTruthy()
    expect(screen.getByRole('button', { name: /Preview/ })).toBeTruthy()
    expect(screen.getByRole('button', { name: /Date/ })).toBeTruthy()
  })

  test('next page is unavailable when there is no next page', () => {
    const nextDisabled = () =>
      screen.getByRole('button', { name: 'Next' }).getAttribute('aria-disabled')

    const { rerender } = render(
      <WorkflowTimeline {...base} workflows={branchRuns} />
    )
    expect(nextDisabled()).toBe('true')

    rerender(<WorkflowTimeline {...base} workflows={branchRuns} hasNext />)
    expect(nextDisabled()).toBeNull()
  })
})
