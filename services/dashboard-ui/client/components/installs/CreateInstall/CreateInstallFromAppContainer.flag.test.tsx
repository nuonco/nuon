import { afterEach, expect, test } from 'bun:test'
import {
  act,
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router'
import { AuthContext } from '@/providers/auth-provider'
import { OrgContext } from '@/providers/org-provider'
import { SurfacesProvider } from '@/providers/surfaces-provider'
import { ToastContext } from '@/providers/toast-provider'
import type { TApp, TAppBranch } from '@/types'
import { CreateInstallFromAppContainer } from './CreateInstallFromAppContainer'

const orgId = 'org-acme'
const appId = 'app-acme'
const app = { id: appId, name: 'acme' } as TApp

const realFetch = globalThis.fetch

afterEach(() => {
  cleanup()
  globalThis.fetch = realFetch
})

type TFetchHandler = (
  url: string,
  init?: RequestInit
) => { status: number; body: unknown } | undefined

function setup({
  appSyncDisabled,
  hasBranches,
  branches,
  installInputConfigResponse,
  fetchCalls,
  onStateChange,
  queryData,
  fetchHandler,
}: {
  appSyncDisabled: boolean
  hasBranches?: boolean
  branches?: TAppBranch[]
  installInputConfigResponse?: { status: number; body: unknown }
  fetchCalls?: string[]
  onStateChange?: (state: any) => void
  queryData?: [unknown[], unknown][]
  fetchHandler?: TFetchHandler
}) {
  const client = new QueryClient({
    defaultOptions: {
      queries: { retry: false, staleTime: Infinity },
      mutations: { retry: false },
    },
  })

  client.setQueryData(['app-branches', orgId, appId], {
    data:
      branches ??
      (hasBranches ? [{ id: 'branch-main', name: 'main', configs: [] }] : []),
  })

  for (const [key, value] of queryData ?? []) {
    client.setQueryData(key, value)
  }

  beforeEachFetchStub(installInputConfigResponse, fetchCalls, fetchHandler)

  render(
    <QueryClientProvider client={client}>
      <MemoryRouter>
        <AuthContext.Provider value={{ user: null } as never}>
          <OrgContext.Provider
            value={
              {
                org: {
                  id: orgId,
                  name: 'acme',
                  features: appSyncDisabled ? { 'disable-app-sync': true } : {},
                },
                refresh: () => {},
              } as never
            }
          >
            <ToastContext.Provider
              value={{ addToast: () => {}, removeToast: () => {} } as never}
            >
              <SurfacesProvider>
                <CreateInstallFromAppContainer
                  app={app}
                  onStateChange={onStateChange ?? (() => {})}
                />
              </SurfacesProvider>
            </ToastContext.Provider>
          </OrgContext.Provider>
        </AuthContext.Provider>
      </MemoryRouter>
    </QueryClientProvider>
  )
}

function beforeEachFetchStub(
  installInputConfigResponse?: {
    status: number
    body: unknown
  },
  calls?: string[],
  fetchHandler?: TFetchHandler
) {
  globalThis.fetch = Object.assign(
    async (input: Parameters<typeof fetch>[0], init?: RequestInit) => {
      const url = String(input)
      calls?.push(url)
      const handled = fetchHandler?.(url, init)
      if (handled) {
        return new Response(JSON.stringify(handled.body), {
          status: handled.status,
          headers: {
            'content-type': 'application/json',
            'X-Nuon-Page-Next': 'false',
          },
        })
      }
      if (url.includes('/install-input-config')) {
        const response = installInputConfigResponse ?? {
          status: 200,
          body: { app_config_id: 'config-from-input' },
        }
        return new Response(JSON.stringify(response.body), {
          status: response.status,
          headers: { 'content-type': 'application/json' },
        })
      }
      if (url.includes('/configs/config-from-input')) {
        return new Response(
          JSON.stringify({
            id: 'config-from-input',
            component_ids: [],
            input: { id: 'input-1', input_groups: [], inputs: [] },
          }),
          { status: 200, headers: { 'content-type': 'application/json' } }
        )
      }
      return new Response(JSON.stringify([]), {
        status: 200,
        headers: {
          'content-type': 'application/json',
          'X-Nuon-Page-Next': 'false',
        },
      })
    },
    { preconnect: () => {} }
  ) as unknown as typeof fetch
}

test('when app sync is disabled and the app has no branches the empty state text shows and the form does not', async () => {
  setup({ appSyncDisabled: true, hasBranches: false })

  expect(await screen.findByText('No app branches yet')).toBeTruthy()
  expect(
    screen.getByText(
      'Create an app branch, then come back to create installs once it has a successful run.'
    )
  ).toBeTruthy()
  expect(screen.queryByPlaceholderText('Enter install name')).toBeNull()
})

test('when app sync is disabled and the app has no branches, Create branch opens the create branch modal', async () => {
  setup({ appSyncDisabled: true, hasBranches: false })

  fireEvent.click(await screen.findByRole('button', { name: 'Create branch' }))

  expect(await screen.findByText('Create app branch')).toBeTruthy()
})

test('when app sync is disabled and the app has branches, Skip app branch is hidden', async () => {
  setup({ appSyncDisabled: true, hasBranches: true })

  expect(await screen.findByText('main')).toBeTruthy()
  expect(screen.queryByRole('button', { name: 'Skip app branch' })).toBeNull()
})

test('when app sync is enabled and the app has branches, Skip app branch shows', async () => {
  setup({ appSyncDisabled: false, hasBranches: true })

  expect(
    await screen.findByRole('button', { name: 'Skip app branch' })
  ).toBeTruthy()
})

test('when the selected branch has only a failed run, the form explains a successful branch run is needed and links to the branch', async () => {
  let latestState: any = null

  setup({
    appSyncDisabled: false,
    branches: [{ id: 'branch-main', name: 'main', configs: [] }],
    onStateChange: (state) => {
      latestState = state
    },
    fetchHandler: (url) =>
      url.includes('/branches/branch-main/runs')
        ? {
            status: 200,
            body: [
              {
                id: 'run-1',
                status: {
                  status: 'error',
                  status_human_description:
                    'workflow stopped: missing required file metadata.toml',
                },
              },
            ],
          }
        : { status: 200, body: [] },
  })

  fireEvent.click(await screen.findByText('main'))
  await waitFor(() => expect(latestState?.canSubmit).toBe(true))
  await act(async () => {
    latestState.submit()
  })

  expect(
    await screen.findByText("This app branch's first run failed")
  ).toBeTruthy()
  expect(
    screen.getByText(
      'Installs need an app config from a successful branch run. Fix the error below, then run the branch again.'
    )
  ).toBeTruthy()
  expect(
    screen.getByText('workflow stopped: missing required file metadata.toml')
  ).toBeTruthy()
  expect(
    screen.getByRole('link', { name: 'View branch' }).getAttribute('href')
  ).toBe(`/${orgId}/apps/${appId}/branches/branch-main`)
  expect(screen.queryByPlaceholderText('Enter install name')).toBeNull()
})

test('flag off unbranched path uses app_config_id from install-input-config', async () => {
  setup({
    appSyncDisabled: false,
    hasBranches: false,
    installInputConfigResponse: {
      status: 200,
      body: {
        app_config_id: 'config-from-input',
        input_groups: [],
        inputs: [],
      },
    },
  })

  await waitFor(() =>
    expect(screen.getByPlaceholderText('Enter install name')).toBeTruthy()
  )
})

test('flag off shows the server error description when install-input-config fails', async () => {
  setup({
    appSyncDisabled: false,
    hasBranches: false,
    installInputConfigResponse: {
      status: 400,
      body: { error: 'no sync config found for this app' },
    },
  })

  await waitFor(() =>
    expect(screen.getByText('no sync config found for this app')).toBeTruthy()
  )
  expect(
    screen.queryByText(/No active config from `nuon apps sync` was found/)
  ).toBeNull()
})

test('flag off unbranched path never requests the legacy apps-sync config list', async () => {
  const fetchCalls: string[] = []
  setup({
    appSyncDisabled: false,
    hasBranches: false,
    fetchCalls,
  })

  await waitFor(() =>
    expect(screen.getByPlaceholderText('Enter install name')).toBeTruthy()
  )

  expect(
    fetchCalls.some((url) => /\/apps\/[^/]+\/configs(\?|$)/.test(url))
  ).toBe(false)
})

test('when app sync is disabled, selecting a branch and continuing reaches the form for that branch', async () => {
  const calls: { url: string; method: string; body?: any }[] = []
  let latestState: any = null

  const branches = [
    {
      id: 'branch-main',
      name: 'main',
      configs: [
        {
          id: 'config-1',
          status: 'active',
          install_groups: [{ id: 'group-1', name: 'default', default: true }],
        },
      ],
    },
  ] as unknown as TAppBranch[]

  setup({
    appSyncDisabled: true,
    branches,
    onStateChange: (state) => {
      latestState = state
    },
    queryData: [
      [
        ['app-branch-app-configs', orgId, appId, 'branch-main'],
        [
          {
            id: 'config-1',
            status: 'active',
            install_groups: [{ id: 'group-1', name: 'default', default: true }],
          },
        ],
      ],
      [
        ['app-config', orgId, appId, 'config-1'],
        {
          id: 'config-1',
          component_ids: [],
          input: { id: 'input-1', input_groups: [], inputs: [] },
        },
      ],
    ],
    fetchHandler: (url, init) => {
      const method = init?.method ?? 'GET'
      const body = init?.body ? JSON.parse(String(init.body)) : undefined
      calls.push({ url, method, body })
      if (url.includes('/installs') && method === 'POST') {
        return {
          status: 201,
          body: { id: 'install-1', name: body?.name },
        }
      }
      return undefined
    },
  })

  const branchButton = await screen.findByText('main')
  fireEvent.click(branchButton)

  await waitFor(() => expect(latestState?.canSubmit).toBe(true))
  await act(async () => {
    latestState.submit()
  })

  const nameInput = await screen.findByPlaceholderText('Enter install name')
  fireEvent.change(nameInput, { target: { value: 'fresh-install' } })

  expect(calls.some((c) => c.url.includes('/install-input-config'))).toBe(false)

  await waitFor(() =>
    expect(calls.some((c) => c.url.includes('q=fresh-install'))).toBe(true)
  )
  await waitFor(() => expect(latestState?.canSubmit).toBe(true))
  await act(async () => {
    await latestState.submit()
  })

  await waitFor(() => expect(latestState?.phase).toBe('pick-group'))

  const groupButton = await screen.findByText('default')
  fireEvent.click(groupButton)

  await waitFor(() => expect(latestState?.canSubmit).toBe(true))
  await act(async () => {
    latestState.submit()
  })

  await waitFor(() =>
    expect(
      calls.some((c) => c.url.includes('/installs') && c.method === 'POST')
    ).toBe(true)
  )
  const createCall = calls.find(
    (c) => c.url.includes('/installs') && c.method === 'POST'
  )
  expect(createCall?.body?.app_branch_id).toBe('branch-main')
})
