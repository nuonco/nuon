import { type ComponentProps } from 'react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { afterEach, expect, mock, test } from 'bun:test'
import { cleanup, fireEvent, render, screen, waitFor, within } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { OrgContext } from '@/providers/org-provider'
import { SurfacesProvider } from '@/providers/surfaces-provider'
import { ToastProvider } from '@/providers/toast-provider'
import { BranchOverviewV2 } from './BranchOverviewV2'
import { branchOverviewVariant } from '@/hooks/use-branch-overview-v2'
import type { TTrackGroup } from './RolloutTrack'

afterEach(cleanup)

const groups: TTrackGroup[] = [
  {
    id: 'canary',
    name: 'Canary',
    status: 'success',
    installs: [
      {
        id: 'ins_alpha',
        name: 'acme-alpha',
        status: 'success',
        overviewHref: '/org-acme/installs/ins_alpha',
        workflowId: 'wf_alpha',
      },
    ],
  },
  {
    id: 'primary',
    name: 'Primary region',
    status: 'in-progress',
    installs: [
      {
        id: 'ins_delta',
        name: 'acme-delta',
        status: 'in-progress',
        overviewHref: '/org-acme/installs/ins_delta',
        workflowId: 'wf_delta',
      },
    ],
  },
  {
    id: 'remaining',
    name: 'Remaining',
    status: 'pending',
    installs: [],
  },
]

const renderOverview = (
  props?: Partial<ComponentProps<typeof BranchOverviewV2>>
) =>
  render(
    <MemoryRouter>
      <SurfacesProvider>
        <BranchOverviewV2
          hasPlan
          groups={groups}
          rolloutHref="#rollout"
          groupHref={(id) => `#groups/${id}`}
          rollout={{
            id: 'wf_184',
            href: '#run',
            source: { kind: 'manual' },
            title: 'Add cache component',
            status: 'in-progress',
          }}
          {...props}
        />
      </SurfacesProvider>
    </MemoryRouter>
  )

test('keeps the current overview unless both flags are on', () => {
  expect(branchOverviewVariant(false, false)).toBe('legacy')
  expect(branchOverviewVariant(false, true)).toBe('legacy')
  expect(branchOverviewVariant(true, false)).toBe('current')
  expect(branchOverviewVariant(true, true)).toBe('v2')
})

test('shows approve all only when an install is awaiting approval', () => {
  renderOverview()
  expect(screen.queryByRole('button', { name: 'Approve all' })).toBeNull()

  cleanup()
  renderOverview({
    groups: groups.map((group) =>
      group.id === 'primary'
        ? {
            ...group,
            installs: group.installs.map((install) => ({
              ...install,
              status: 'approval-awaiting',
            })),
          }
        : group
    ),
  })
  expect(screen.getAllByRole('button', { name: 'Approve all' }).length).toBeGreaterThan(0)
})

test('approve all marks every install workflow in the group', async () => {
  const calls: { workflowId: string; body: { approval_option?: string } }[] = []
  const original = globalThis.fetch
  globalThis.fetch = (async (input: RequestInfo | URL, init?: RequestInit) => {
    const url = typeof input === 'string' ? input : input instanceof URL ? input.href : input.url
    if (init?.method === 'PATCH' && url.includes('/v1/workflows/')) {
      calls.push({
        workflowId: url.split('/').pop() ?? '',
        body: JSON.parse(String(init.body)),
      })
      return new Response(JSON.stringify({ id: 'wf' }), {
        status: 200,
        headers: { 'content-type': 'application/json' },
      })
    }
    return original(input, init)
  }) as typeof fetch

  try {
    const client = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    })
    render(
      <QueryClientProvider client={client}>
        <OrgContext.Provider
          value={{ org: { id: 'org-acme' } as never, refresh: () => {} }}
        >
          <ToastProvider>
            <MemoryRouter>
              <SurfacesProvider>
                <BranchOverviewV2
                  hasPlan
                  groups={[
                    {
                      id: 'primary',
                      name: 'Primary region',
                      status: 'in-progress',
                      installs: [
                        {
                          id: 'ins_done',
                          name: 'acme-done',
                          status: 'success',
                          workflowId: 'wf_done',
                        },
                        {
                          id: 'ins_wait',
                          name: 'acme-wait',
                          status: 'approval-awaiting',
                          workflowId: 'wf_wait',
                        },
                        {
                          id: 'ins_next',
                          name: 'acme-next',
                          status: 'pending',
                          workflowId: 'wf_next',
                        },
                      ],
                    },
                  ]}
                  rolloutHref="#rollout"
                  groupHref={(id) => `#groups/${id}`}
                />
              </SurfacesProvider>
            </MemoryRouter>
          </ToastProvider>
        </OrgContext.Provider>
      </QueryClientProvider>
    )

    fireEvent.click(screen.getAllByRole('button', { name: 'Approve all' })[0])
    const dialog = screen.getByRole('dialog')
    expect(
      within(dialog).getByText(/Every install workflow in Primary region/)
    ).toBeTruthy()
    fireEvent.click(within(dialog).getByRole('button', { name: 'Approve all' }))

    await waitFor(() => {
      expect(calls.map((call) => call.workflowId).sort()).toEqual([
        'wf_done',
        'wf_next',
        'wf_wait',
      ])
    })
    expect(calls.every((call) => call.body.approval_option === 'approve-all')).toBe(true)
  } finally {
    globalThis.fetch = original
  }
})

test('shows the active group above the plan and pending groups as pending', () => {
  renderOverview()

  const deploying = screen.getByText(/Group 2 of 3/)
  const plan = screen.getByText('Deployment plan')
  expect(
    deploying.compareDocumentPosition(plan) & Node.DOCUMENT_POSITION_FOLLOWING
  ).toBeTruthy()

  const current = document.getElementById('deployment-group-primary')
  const later = document.getElementById('deployment-group-remaining')
  expect(current?.getAttribute('data-current')).toBe('true')
  expect(current?.getAttribute('aria-current')).toBe('step')
  expect(current?.textContent).toContain('Current')
  expect(later?.textContent).not.toContain('Current')
  expect(later?.className).not.toContain('opacity-60')
  expect(later?.textContent).toContain('Pending')
  expect(later?.textContent).not.toContain('Waiting')
})

test('scrolls to a group when its plan segment is clicked', () => {
  renderOverview()
  const row = document.getElementById('deployment-group-remaining')
  const scroll = mock(() => {})
  row!.scrollIntoView = scroll

  fireEvent.click(screen.getByRole('button', { name: 'Remaining, pending' }))

  expect(scroll).toHaveBeenCalled()
})

test('does not show a group failure as a banner on the active group', () => {
  renderOverview({
    groups: groups.map((group) =>
      group.id === 'primary'
        ? {
            ...group,
            status: 'error',
            error: {
              message: 'acme-delta failed during deploy',
              code: 'install_group.install_update_failed',
            },
          }
        : group
    ),
  })

  expect(screen.queryByText('acme-delta failed during deploy')).toBeNull()
  expect(screen.queryByText('What happened')).toBeNull()
  expect(screen.getByText('Rollout failed')).toBeTruthy()
})

test('counts installs by status on each group', () => {
  renderOverview({
    groups: [
      {
        id: 'canary',
        name: 'Canary',
        status: 'error',
        installs: [
          { id: 'ins_a', name: 'acme-a', status: 'success' },
          { id: 'ins_b', name: 'acme-b', status: 'success' },
          { id: 'ins_c', name: 'acme-c', status: 'error' },
          { id: 'ins_d', name: 'acme-d', status: 'cancelled' },
        ],
      },
      {
        id: 'primary',
        name: 'Primary region',
        status: 'in-progress',
        installs: [
          { id: 'ins_e', name: 'acme-e', status: 'in-progress' },
          { id: 'ins_f', name: 'acme-f', status: 'pending' },
        ],
      },
      {
        id: 'remaining',
        name: 'Remaining',
        status: 'pending',
        installs: [],
      },
    ],
  })

  const canary = document.getElementById('deployment-group-canary')
  expect(canary?.textContent).toContain('2 completed')
  expect(canary?.textContent).toContain('1 errored')
  expect(canary?.textContent).toContain('1 cancelled')
  expect(canary?.textContent).not.toContain('in progress')

  const primary = document.getElementById('deployment-group-primary')
  expect(primary?.textContent).toContain('1 in progress')
  expect(primary?.textContent).toContain('1 pending')

  const remaining = document.getElementById('deployment-group-remaining')
  expect(remaining?.textContent).toContain('None')
  expect(remaining?.textContent).toContain('Pending')
})

test('shows the last approved version without the commit message', () => {
  renderOverview({
    approvedVersions: {
      canary: {
        version: 12,
        source: {
          kind: 'pull-request',
          number: 128,
          url: 'https://github.com/acme/platform/pull/128',
        },
        sha: 'a1b2c3d4e5f6',
        message: 'Add cache component\n\nMore detail for the tooltip.',
        author: 'jane@example.com',
        createdAt: '2026-10-15T13:40:00Z',
      },
    },
  })

  const row = document.getElementById('deployment-group-canary')
  expect(row?.textContent).toContain('v12')
  expect(row?.textContent).toContain('PR #128')
  expect(row?.textContent).not.toContain('Add cache component')
  expect(screen.getByText(/Add cache component · jane@example.com/)).toBeTruthy()

  const pending = document.getElementById('deployment-group-remaining')
  expect(pending?.textContent).toContain('None')
})

test('keeps the approval banner on the active group and out of group details', () => {
  renderOverview({
    approvals: [
      {
        groupName: 'Primary region',
        banner: <div>Plan for Primary region requires approval</div>,
      },
    ],
  })

  expect(screen.getAllByText('Plan for Primary region requires approval')).toHaveLength(1)

  fireEvent.click(screen.getByRole('button', { name: 'View group details' }))

  expect(screen.getAllByText('Plan for Primary region requires approval')).toHaveLength(1)
  expect(screen.getByRole('link', { name: 'acme-delta' })).toBeTruthy()
})

test('links an install name to its page and opens details from the group', () => {
  const onSelectInstall = mock(() => {})
  renderOverview({ onSelectInstall })

  fireEvent.click(screen.getByRole('button', { name: 'acme-delta: in-progress' }))
  expect(onSelectInstall).toHaveBeenCalledWith(
    expect.objectContaining({ id: 'ins_delta', workflowId: 'wf_delta' })
  )

  fireEvent.click(screen.getByRole('button', { name: 'View group details' }))
  expect(screen.getByRole('link', { name: 'acme-delta' }).getAttribute('href')).toBe(
    '/org-acme/installs/ins_delta'
  )
})
