import { afterEach, expect, test } from 'bun:test'
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router'
import { SurfacesProvider } from '@/providers/surfaces-provider'
import { CollectorSetupContainer } from './CollectorSetupContainer'

const orgId = 'org-acme'
const install = {
  id: 'install-acme',
  name: 'acme-production',
  app_id: 'app-acme',
}

afterEach(cleanup)

function setup(appSyncDisabled: boolean) {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false, staleTime: Infinity } },
  })
  client.setQueryData(['account'], {
    roles: [{ org_id: orgId, role_type: 'org_admin' }],
  })
  client.setQueryData(['install-telemetry-collector', orgId, install.id], null)
  client.setQueryData(
    ['install-components', orgId, install.id, 'telemetry-collector'],
    { data: [], pagination: { hasNext: false } }
  )

  render(
    <MemoryRouter>
      <QueryClientProvider client={client}>
        <SurfacesProvider>
          <CollectorSetupContainer
            install={install as never}
            orgId={orgId}
            onNavigate={() => {}}
            appSyncDisabled={appSyncDisabled}
          />
        </SurfacesProvider>
      </QueryClientProvider>
    </MemoryRouter>
  )

  fireEvent.click(
    screen.getByRole('button', { name: /In-cluster collector \(optional\)/ })
  )
  fireEvent.click(
    screen.getByRole('button', { name: /Set up the in-cluster collector/ })
  )
}

test('flag off tells the user to sync the app config', () => {
  setup(false)
  expect(screen.getByText('Sync the app config.')).toBeTruthy()
  expect(
    screen.getByText(
      (_, element) =>
        element?.tagName === 'LI' &&
        element?.textContent ===
          'Sync the app config. Run nuon apps sync from the app config directory, or run your app branch, to publish the component, label, and input. This does not deploy existing installs.'
    )
  ).toBeTruthy()
})

test('when app sync is disabled tells the user to update the app config', () => {
  setup(true)
  expect(screen.getByText('Update the app config.')).toBeTruthy()
  expect(
    screen.getByText(
      (_, element) =>
        element?.tagName === 'LI' &&
        element?.textContent ===
          'Update the app config. Run your app branch to publish the component, label, and input. This does not deploy existing installs.'
    )
  ).toBeTruthy()
  expect(screen.queryByText('nuon apps sync')).toBeNull()
})
