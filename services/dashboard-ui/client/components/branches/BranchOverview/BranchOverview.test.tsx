import { type ComponentProps } from 'react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { afterEach, expect, mock, test } from 'bun:test'
import { cleanup, fireEvent, render, screen, waitFor, within } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { OrgContext } from '@/providers/org-provider'
import { SurfacesProvider } from '@/providers/surfaces-provider'
import { ToastProvider } from '@/providers/toast-provider'
import { BranchOverview } from './BranchOverview'
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
  props?: Partial<ComponentProps<typeof BranchOverview>>
) =>
  render(
    <MemoryRouter>
      <SurfacesProvider>
        <BranchOverview
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
                <BranchOverview
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

test('shows a loader instead of the run source while fetching the commit', () => {
  renderOverview({
    loadingStages: [
      { id: 'starting', label: 'Starting workflow', status: 'success' },
      { id: 'fetch-commit', label: 'Fetch commit', status: 'in-progress' },
      {
        id: 'app-config',
        label: 'Compile and diff app config',
        status: 'pending',
      },
    ],
  })

  expect(screen.getByRole('status', { name: 'Waiting on latest commit' })).toBeTruthy()
  expect(screen.queryByText('Add cache component')).toBeNull()
  expect(screen.queryByText(/Manual run/)).toBeNull()

  cleanup()
  renderOverview({
    rollout: {
      id: 'wf_184',
      href: '#run',
      source: { kind: 'pull-request', number: 482 },
      title: 'Add cache component',
      status: 'in-progress',
    },
    loadingStages: [
      { id: 'fetch-commit', label: 'Fetch commit', status: 'in-progress' },
    ],
  })
  expect(
    screen.getByRole('status', { name: 'Waiting on pull request' })
  ).toBeTruthy()
})

test('hides the changes button on the compile step', () => {
  renderOverview({
    changes: <button type="button">Changes</button>,
    rollout: {
      id: 'wf_184',
      href: '#run',
      source: { kind: 'manual' },
      title: 'Add cache component',
      sha: 'a1b2c3d4e5f6',
      status: 'in-progress',
    },
    configChanges: <p>Calculating config diff</p>,
    loadingStages: [
      {
        id: 'app-config',
        label: 'Compile and diff app config',
        status: 'in-progress',
      },
      { id: 'build-components', label: 'Build components', status: 'pending' },
    ],
  })

  expect(screen.getByText('Calculating config diff')).toBeTruthy()
  expect(screen.queryByRole('button', { name: 'Changes' })).toBeNull()
})

test('shows the changes button once the diff is ready on another step', () => {
  renderOverview({
    changes: <button type="button">Changes</button>,
    rollout: {
      id: 'wf_184',
      href: '#run',
      source: { kind: 'manual' },
      title: 'Add cache component',
      sha: 'a1b2c3d4e5f6',
      status: 'in-progress',
    },
    configChanges: <p>Added cache component</p>,
    loadingStages: [
      {
        id: 'app-config',
        label: 'Compile and diff app config',
        status: 'success',
      },
      {
        id: 'build-components',
        label: 'Build components',
        status: 'in-progress',
      },
    ],
  })

  expect(screen.getByRole('button', { name: 'Changes' })).toBeTruthy()
  fireEvent.click(
    screen.getByRole('button', { name: /Compile and diff app config/ })
  )
  expect(screen.queryByRole('button', { name: 'Changes' })).toBeNull()
  expect(screen.getByText('Added cache component')).toBeTruthy()
})

test('shows config changes on the compile step', () => {
  renderOverview({
    loadingStages: [
      { id: 'starting', label: 'Starting workflow', status: 'success' },
      { id: 'fetch-commit', label: 'Fetch commit', status: 'success' },
      {
        id: 'app-config',
        label: 'Compile and diff app config',
        status: 'in-progress',
      },
      { id: 'build-components', label: 'Build components', status: 'pending' },
      { id: 'rollout', label: 'Rollout', status: 'pending' },
    ],
    configChanges: <p>Added cache component</p>,
  })

  expect(screen.getByText('Added cache component')).toBeTruthy()
  expect(screen.queryByText(/Group 2 of 3/)).toBeNull()
  expect(screen.queryByRole('button', { name: /Build components/ })).toBeNull()
  expect(screen.queryByRole('button', { name: /^Rollout/ })).toBeNull()
})

test('reopens config changes after the compile step finishes', () => {
  renderOverview({
    loadingStages: [
      {
        id: 'app-config',
        label: 'Compile and diff app config',
        status: 'success',
      },
      { id: 'build-components', label: 'Build components', status: 'success' },
      { id: 'rollout', label: 'Rollout', status: 'in-progress' },
    ],
    configChanges: <p>Added cache component</p>,
  })

  expect(screen.queryByText('Added cache component')).toBeNull()
  expect(screen.getByText(/Group 2 of 3/)).toBeTruthy()

  fireEvent.click(
    screen.getByRole('button', { name: /Compile and diff app config/ })
  )
  expect(screen.getByText('Added cache component')).toBeTruthy()
  expect(screen.queryByText(/Group 2 of 3/)).toBeNull()

  fireEvent.click(screen.getByRole('button', { name: /Rollout/ }))
  expect(screen.queryByText('Added cache component')).toBeNull()
  expect(screen.getByText(/Group 2 of 3/)).toBeTruthy()
})

test('shows the commit above the selected step', () => {
  renderOverview({
    loadingStages: [
      {
        id: 'app-config',
        label: 'Compile and diff app config',
        status: 'success',
      },
      { id: 'build-components', label: 'Build components', status: 'success' },
      { id: 'rollout', label: 'Rollout', status: 'in-progress' },
    ],
    buildRows: [
      {
        id: 'build-api',
        name: 'api',
        status: 'success',
      },
    ],
  })

  fireEvent.click(screen.getByRole('button', { name: /Build components/ }))
  const commit = screen.getByText('Add cache component')
  const builds = screen.getByText('Component name')
  expect(
    commit.compareDocumentPosition(builds) & Node.DOCUMENT_POSITION_FOLLOWING
  ).toBeTruthy()

  fireEvent.click(screen.getByRole('button', { name: /Rollout/ }))
  const group = screen.getByText(/Group 2 of 3/)
  expect(
    screen.getByText('Add cache component').compareDocumentPosition(group) &
      Node.DOCUMENT_POSITION_FOLLOWING
  ).toBeTruthy()
})

test('shows builds instead of a group while components are building', () => {
  renderOverview({
    groups: groups.map((group) => ({
      ...group,
      status: 'pending',
      installs: group.installs.map((install) => ({
        ...install,
        status: 'pending',
      })),
    })),
    loadingStages: [
      { id: 'starting', label: 'Starting workflow', status: 'success' },
      { id: 'fetch-commit', label: 'Fetch commit', status: 'success' },
      {
        id: 'app-config',
        label: 'Compile and diff app config',
        status: 'success',
      },
      {
        id: 'build-components',
        label: 'Build components',
        status: 'in-progress',
      },
      { id: 'rollout', label: 'Rollout', status: 'pending' },
    ],
    buildRows: [
      {
        id: 'build-api',
        name: 'api',
        status: 'in-progress',
        changeReason: 'source_changed',
      },
      {
        id: 'build-cache',
        name: 'cache',
        status: 'pending',
        changeReason: 'config_changed',
      },
    ],
  })

  expect(screen.getByText('api')).toBeTruthy()
  expect(screen.getByText('cache')).toBeTruthy()
  expect(screen.getByText('Source')).toBeTruthy()
  expect(screen.getByText('Config')).toBeTruthy()
  expect(screen.getByText('Component name')).toBeTruthy()
  expect(screen.queryByText(/Group 1 of 3/)).toBeNull()
  expect(screen.queryByRole('button', { name: /^Rollout/ })).toBeNull()

  fireEvent.click(
    screen.getByRole('button', { name: /Compile and diff app config/ })
  )
  expect(screen.queryByText('Component name')).toBeNull()

  fireEvent.click(screen.getByRole('button', { name: /Build components/ }))
  expect(screen.getByText('Component name')).toBeTruthy()
  expect(screen.getByText('api')).toBeTruthy()
})

test('shows cancel rollout on the active group and in its drawer', () => {
  renderOverview({
    cancelRollout: <button type="button">Cancel rollout</button>,
  })

  expect(screen.getAllByRole('button', { name: 'Cancel rollout' })).toHaveLength(1)
  fireEvent.click(screen.getByRole('button', { name: 'View group details' }))
  expect(screen.getAllByRole('button', { name: 'Cancel rollout' })).toHaveLength(2)
})

test('reopens builds after the rollout finishes', () => {
  renderOverview({
    groups: groups.map((group) => ({
      ...group,
      status: 'success',
      installs: group.installs.map((install) => ({
        ...install,
        status: 'success',
      })),
    })),
    loadingStages: [
      { id: 'build-components', label: 'Build components', status: 'success' },
      { id: 'rollout', label: 'Rollout', status: 'success' },
    ],
    buildRows: [
      {
        id: 'build-api',
        name: 'api',
        status: 'success',
        componentId: 'cmp-api',
        href: '/org-acme/apps/app-acme/components/cmp-api/builds/build-api',
      },
    ],
  })

  expect(screen.queryByText('Component name')).toBeNull()
  fireEvent.click(screen.getByRole('button', { name: /Build components/ }))
  expect(screen.getByText('Component name')).toBeTruthy()
  expect(screen.getByText('api')).toBeTruthy()
})

test('opens a build from the builds table', () => {
  const onSelectBuild = mock(() => {})
  renderOverview({
    loadingStages: [
      { id: 'build-components', label: 'Build components', status: 'in-progress' },
    ],
    onSelectBuild,
    buildRows: [
      {
        id: 'build-api',
        name: 'api',
        status: 'in-progress',
        componentId: 'cmp-api',
        href: '/org-acme/apps/app-acme/components/cmp-api/builds/build-api',
      },
    ],
  })

  fireEvent.click(screen.getByRole('button', { name: 'View build' }))
  expect(onSelectBuild).toHaveBeenCalled()
})

test('defaults to the first group after builds and can reopen build details', () => {
  renderOverview({
    groups: groups.map((group) => ({
      ...group,
      status: 'pending',
      installs: group.installs.map((install) => ({
        ...install,
        status: 'pending',
      })),
    })),
    loadingStages: [
      { id: 'starting', label: 'Starting workflow', status: 'success' },
      { id: 'fetch-commit', label: 'Fetch commit', status: 'success' },
      {
        id: 'app-config',
        label: 'Compile and diff app config',
        status: 'success',
      },
      {
        id: 'build-components',
        label: 'Build components',
        status: 'success',
      },
      { id: 'rollout', label: 'Rollout', status: 'in-progress' },
    ],
    buildRows: [
      {
        id: 'build-api',
        name: 'api',
        status: 'success',
        changeReason: 'source_changed',
      },
    ],
    configChanges: <p>Added cache component</p>,
  })

  expect(screen.getByText(/Group 1 of 3/)).toBeTruthy()
  expect(screen.queryByText('Component name')).toBeNull()

  fireEvent.click(screen.getByRole('button', { name: /Build components/ }))
  expect(screen.getByText('Component name')).toBeTruthy()
  expect(screen.queryByText(/Group 1 of 3/)).toBeNull()

  fireEvent.click(
    screen.getByRole('button', { name: /Compile and diff app config/ })
  )
  expect(screen.getByText('Added cache component')).toBeTruthy()
  expect(screen.queryByText('Component name')).toBeNull()

  fireEvent.click(screen.getByRole('button', { name: /Build components/ }))
  expect(screen.getByText('Component name')).toBeTruthy()
  expect(screen.queryByText('Added cache component')).toBeNull()

  fireEvent.click(screen.getByRole('button', { name: /^Rollout/ }))
  expect(screen.getByText(/Group 1 of 3/)).toBeTruthy()
  expect(screen.queryByText('Component name')).toBeNull()
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

test('shows a cancelled rollout instead of a not-started group', () => {
  renderOverview({
    groups: groups.map((group) =>
      group.id === 'primary' || group.id === 'remaining'
        ? { ...group, status: 'cancelled' }
        : group
    ),
  })

  expect(screen.getByText('Rollout cancelled')).toBeTruthy()
  expect(screen.queryByText('Not started')).toBeNull()
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
          { id: 'ins_g', name: 'acme-g', status: 'pending_customer' },
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
  expect(primary?.textContent).toContain('1 pending customer')

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

test('opens install details from the active group segment instead of a link', () => {
  const onSelectInstall = mock(() => {})
  renderOverview({
    onSelectInstall,
    groups: groups.map((group) =>
      group.id === 'primary'
        ? {
            ...group,
            installs: [
              ...group.installs,
              {
                id: 'ins_echo',
                name: 'acme-echo',
                status: 'pending',
                overviewHref: '/org-acme/installs/ins_echo',
              },
            ],
          }
        : group
    ),
  })

  expect(screen.queryByRole('link', { name: 'acme-delta: in-progress' })).toBeNull()
  expect(screen.queryByRole('link', { name: 'acme-echo: pending' })).toBeNull()
  fireEvent.click(screen.getByRole('button', { name: 'acme-delta: in-progress' }))
  fireEvent.click(screen.getByRole('button', { name: 'acme-echo: pending' }))
  expect(onSelectInstall).toHaveBeenCalledWith(
    expect.objectContaining({ id: 'ins_delta', workflowId: 'wf_delta' })
  )
  expect(onSelectInstall).toHaveBeenCalledWith(
    expect.objectContaining({ id: 'ins_echo', status: 'pending' })
  )
})
