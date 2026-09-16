import { afterEach, beforeEach, expect, test } from 'bun:test'
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router'
import { AuthContext } from '@/providers/auth-provider'
import { OrgContext } from '@/providers/org-provider'
import { SurfacesProvider } from '@/providers/surfaces-provider'
import { ToastContext } from '@/providers/toast-provider'
import type { TApp } from '@/types'
import { CreateInstallFromAppContainer } from './CreateInstallFromAppContainer'

const orgId = 'org-acme'
const appId = 'app-acme'
const branchId = 'branch-main'
const app = { id: appId, name: 'acme' } as TApp

// Every layer below the form is real here: the container calls installNameTaken,
// which calls getAppInstalls, which issues the request stubbed out below.
const existingNames = ['staging']
const lookups: string[] = []
const realFetch = globalThis.fetch

beforeEach(() => {
  lookups.length = 0
  globalThis.fetch = Object.assign(
    async (input: Parameters<typeof fetch>[0]) => {
      const url = new URL(String(input), 'https://api.example.com')
      const q = url.searchParams.get('q')
      if (q) lookups.push(q)
      const matches = q
        ? existingNames.filter((name) => name.includes(q)).map((name) => ({ name }))
        : []
      return new Response(JSON.stringify(matches), {
        status: 200,
        headers: {
          'content-type': 'application/json',
          'X-Nuon-Page-Next': 'false',
        },
      })
    },
    { preconnect: () => {} }
  ) as unknown as typeof fetch
})

afterEach(() => {
  cleanup()
  globalThis.fetch = realFetch
})

function setup() {
  const client = new QueryClient({
    defaultOptions: {
      queries: { retry: false, staleTime: Infinity },
      mutations: { retry: false },
    },
  })

  client.setQueryData(['app-branches', orgId, appId], {
    data: [{ id: branchId, name: 'main', configs: [] }],
  })
  client.setQueryData(
    ['app-branch-app-configs', orgId, appId, branchId],
    [{ id: 'config-1', status: 'active' }]
  )
  client.setQueryData(['app-config', orgId, appId, 'config-1'], {
    id: 'config-1',
    component_ids: [],
    input: { id: 'input-1', input_groups: [], inputs: [] },
  })

  render(
    <QueryClientProvider client={client}>
      <MemoryRouter>
        <AuthContext.Provider value={{ user: null } as never}>
          <OrgContext.Provider
            value={{ org: { id: orgId, name: 'acme' }, refresh: () => {} } as never}
          >
            <ToastContext.Provider
              value={{ addToast: () => {}, removeToast: () => {} } as never}
            >
              <SurfacesProvider>
                <CreateInstallFromAppContainer
                  app={app}
                  initialBranchId={branchId}
                  onStateChange={() => {}}
                />
              </SurfacesProvider>
            </ToastContext.Provider>
          </OrgContext.Provider>
        </AuthContext.Provider>
      </MemoryRouter>
    </QueryClientProvider>
  )
}

test('entering from a branch skips the picker and still flags a duplicate name', async () => {
  setup()

  const input = await screen.findByPlaceholderText('Enter install name')
  fireEvent.change(input, { target: { value: 'staging' } })

  await waitFor(() => expect(lookups).toContain('staging'))
  await waitFor(() =>
    expect(screen.getByText('An install named "staging" already exists')).toBeTruthy()
  )
})

test('entering from a branch accepts an available name', async () => {
  setup()

  const input = await screen.findByPlaceholderText('Enter install name')
  fireEvent.change(input, { target: { value: 'fresh' } })

  await waitFor(() => expect(lookups).toContain('fresh'))
  expect(screen.queryByText(/already exists/)).toBeNull()
})
