import { afterEach, expect, test } from 'bun:test'
import { cleanup, render, screen } from '@testing-library/react'
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

afterEach(() => {
  cleanup()
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
    []
  )
  client.setQueryData(
    [
      'app-branch-runs',
      orgId,
      appId,
      branchId,
      'create-install',
    ],
    [
      {
        status: {
          status: 'error',
          status_human_description:
            'workflow stopped: missing required file metadata.toml',
        },
      },
      { status: { status: 'error' } },
    ]
  )
  client.setQueryData(['cloud-connections', orgId, 'all'], [])

  render(
    <QueryClientProvider client={client}>
      <MemoryRouter>
        <AuthContext.Provider value={{ user: null } as never}>
          <OrgContext.Provider
            value={
              { org: { id: orgId, name: 'acme' }, refresh: () => {} } as never
            }
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

test("when the selected branch's latest run failed after earlier runs, the form shows the run error instead of waiting", async () => {
  setup()

  expect(
    await screen.findByText(
      'workflow stopped: missing required file metadata.toml'
    )
  ).toBeTruthy()
  expect(screen.queryByText(/Waiting for this branch/)).toBeNull()
})
