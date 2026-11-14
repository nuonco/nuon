import { afterEach, expect, test } from 'bun:test'
import { cleanup, render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { OrgContext } from '@/providers/org-provider'
import { InstallContext } from '@/providers/install-provider'
import { SurfacesProvider } from '@/providers/surfaces-provider'
import { PageTitleProvider } from '@/providers/page-title-provider'
import type { TInstall } from '@/types'
import { InstallBranches } from './InstallBranches'

const orgId = 'org-acme'
const install: TInstall = {
  id: 'install-acme',
  name: 'acme-production',
  app_id: 'app-acme',
}

afterEach(cleanup)

function setup(appSyncDisabled: boolean) {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false, staleTime: Infinity } },
  })
  client.setQueryData(['install-app-config-versions', orgId, install.id], [])

  render(
    <MemoryRouter>
      <QueryClientProvider client={client}>
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
          <InstallContext.Provider
            value={{ install, labelColors: {}, refresh: () => {} } as never}
          >
            <PageTitleProvider>
              <SurfacesProvider>
                <InstallBranches />
              </SurfacesProvider>
            </PageTitleProvider>
          </InstallContext.Provider>
        </OrgContext.Provider>
      </QueryClientProvider>
    </MemoryRouter>
  )
}

test('flag off shows the sync-config note', () => {
  setup(false)
  expect(
    screen.getByText(
      'This install uses the latest sync config. Branch-based deployments and install groups are not available until you move it to an app branch.'
    )
  ).toBeTruthy()
})

test('when app sync is disabled shows the no-app-branch note', () => {
  setup(true)
  expect(
    screen.getByText(
      'This install has no app branch. App sync is disabled for this org, so move it to an app branch to get config changes.'
    )
  ).toBeTruthy()
})
