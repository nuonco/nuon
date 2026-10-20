import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { afterEach, expect, test } from 'bun:test'
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { AppContext } from '@/providers/app-provider'
import { OrgContext } from '@/providers/org-provider'
import { SurfacesProvider } from '@/providers/surfaces-provider'
import type { TInstallWorkflowStep } from '@/types'
import type { TTrackGroup } from '@/components/branches/BranchOverview/RolloutTrack'
import { useGroupPlanApprovals } from './use-group-plan-approvals'

afterEach(cleanup)

const step = {
  id: 'step_plan',
  name: 'plan install group: production',
  execution_type: 'approval',
  install_workflow_id: 'wf_184',
  status: { status: 'approval-awaiting' },
  approval: { id: 'apr_1' },
} as TInstallWorkflowStep

const run = { id: 'wf_184', status: { status: 'in-progress' }, steps: [step] }

const groups: TTrackGroup[] = [
  { id: 'production', name: 'production', status: 'success', installs: [] },
]

const Harness = () => {
  const approvals = useGroupPlanApprovals(run, groups)
  return <>{approvals.map((a) => <div key={a.groupName}>{a.banner}</div>)}</>
}

const renderHarness = (initialEntries: string[]) =>
  render(
    <QueryClientProvider client={new QueryClient()}>
      <OrgContext.Provider
        value={{ org: { id: 'org-acme' } as never, refresh: () => {} }}
      >
        <AppContext.Provider
          value={{
            app: { id: 'app-acme', name: 'acme-app', org_id: 'org-acme' } as never,
            labelColors: {},
            refresh: () => {},
          }}
        >
          <MemoryRouter initialEntries={initialEntries}>
            <SurfacesProvider>
              <Harness />
            </SurfacesProvider>
          </MemoryRouter>
        </AppContext.Provider>
      </OrgContext.Provider>
    </QueryClientProvider>
  )

const withMockedFetch = (fn: () => void) => {
  const original = globalThis.fetch
  globalThis.fetch = (async (_input: RequestInfo | URL) =>
    Response.json({})) as typeof fetch
  try {
    fn()
  } finally {
    globalThis.fetch = original
  }
}

test('clicking review changes opens the plan diff panel', () => {
  withMockedFetch(() => {
    renderHarness(['/branches/br-1'])

    fireEvent.click(screen.getByText('Review changes'))

    expect(screen.getByText('Plan for production')).toBeTruthy()
  })
})

test('deep linking straight to the plan diff panel key opens it without a click', () => {
  withMockedFetch(() => {
    renderHarness(['/branches/br-1?panel=plan-diff-step_plan'])

    expect(screen.getByText('Plan for production')).toBeTruthy()
  })
})
