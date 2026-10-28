import { afterEach, beforeEach, expect, test } from 'bun:test'
import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter, useLocation } from 'react-router'
import { AuthContext } from '@/providers/auth-provider'
import { OrgContext } from '@/providers/org-provider'
import { SurfacesProvider } from '@/providers/surfaces-provider'
import { ToastContext } from '@/providers/toast-provider'
import { HealthCoverage } from './InstallsTable'
import { InstallsTableContainer } from './InstallsTableContainer'

const orgId = 'org-acme'
const appId = 'app-acme'
const requests: URL[] = []
let respond: () => Response
const realFetch = globalThis.fetch

const jsonResponse = (body: unknown, status = 200) =>
  new Response(JSON.stringify(body), {
    status,
    headers: {
      'content-type': 'application/json',
      'X-Nuon-Page-Next': 'false',
    },
  })

beforeEach(() => {
  requests.length = 0
  respond = () => jsonResponse([])
  globalThis.fetch = Object.assign(
    async (input: Parameters<typeof fetch>[0]) => {
      requests.push(new URL(String(input), 'https://api.example.com'))
      return respond()
    },
    { preconnect: () => {} }
  ) as unknown as typeof fetch
})

afterEach(() => {
  cleanup()
  globalThis.fetch = realFetch
})

const LocationProbe = () => {
  const location = useLocation()
  return <output data-testid="location">{location.search}</output>
}

function setup(search: string) {
  const client = new QueryClient({
    defaultOptions: { queries: { staleTime: Infinity } },
  })
  render(
    <QueryClientProvider client={client}>
      <MemoryRouter
        initialEntries={[`/${orgId}/apps/${appId}/installs${search}`]}
      >
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
                <InstallsTableContainer appId={appId} shouldPoll={false} />
                <LocationProbe />
              </SurfacesProvider>
            </ToastContext.Provider>
          </OrgContext.Provider>
        </AuthContext.Provider>
      </MemoryRouter>
    </QueryClientProvider>
  )
}

const filterButton = (id: string) =>
  document.getElementById(`dropdown-button-installs-filter-${id}`)!

const installRequests = () =>
  requests.filter((url) => url.pathname.endsWith(`/apps/${appId}/installs`))

test('deep-linked health and lifecycle filters scope the request and show as selected', async () => {
  setup(
    '?health=healthy,unknown&health=healthy&lifecycle_phase=provisioned&offset=20'
  )

  await waitFor(() => expect(installRequests().length).toBe(1))
  const params = installRequests()[0].searchParams
  expect(params.get('health')).toBe('healthy,unknown')
  expect(params.get('lifecycle_phase')).toBe('provisioned')
  expect(params.get('offset')).toBe('20')

  expect(await screen.findByText('No matching installs')).toBeTruthy()
  expect(screen.queryByText('No installs created')).toBeNull()
  expect(filterButton('health').textContent).toContain('Health (2)')
  expect(filterButton('lifecycle-phase').textContent).toContain('Lifecycle (1)')
})

test('changing a filter rewrites only its param and resets the offset', async () => {
  setup('?health=healthy&offset=20&tab=readme')
  await waitFor(() => expect(installRequests().length).toBe(1))

  fireEvent.click(filterButton('health'))
  await waitFor(() =>
    expect(
      document.querySelector('input[name="installs-filter-health-degraded"]')
    ).toBeTruthy()
  )
  fireEvent.click(
    document.querySelector('input[name="installs-filter-health-degraded"]')!
  )

  await waitFor(() =>
    expect(screen.getByTestId('location').textContent).toBe(
      '?health=degraded%2Chealthy&tab=readme'
    )
  )
  await waitFor(() =>
    expect(installRequests().at(-1)?.searchParams.get('health')).toBe(
      'degraded,healthy'
    )
  )
  expect(installRequests().at(-1)?.searchParams.get('offset')).toBe('0')
})

test('an unsupported filter value shows an error instead of an empty list', async () => {
  setup('?health=healthyish')

  expect(await screen.findByText('Installs failed to load')).toBeTruthy()
  expect(
    screen.getByText(/Unsupported health filter value: healthyish/)
  ).toBeTruthy()
  expect(installRequests().length).toBe(0)
})

test('an API rejection shows the API error instead of an empty list', async () => {
  respond = () =>
    jsonResponse(
      { error: 'invalid lifecycle_phase value', user_error: true },
      400
    )
  setup('?lifecycle_phase=provisioned')

  expect(await screen.findByText('Installs failed to load')).toBeTruthy()
  expect(screen.getByText('invalid lifecycle_phase value')).toBeTruthy()
  expect(installRequests().length).toBe(1)
})

test('no filter params preserve the unfiltered request', async () => {
  setup('')

  await waitFor(() => expect(installRequests().length).toBe(1))
  const params = installRequests()[0].searchParams
  expect(params.has('health')).toBe(false)
  expect(params.has('lifecycle_phase')).toBe(false)
  expect(await screen.findByText('No installs created')).toBeTruthy()
})

test('health coverage describes recorded verdicts and keeps unknown, other, and unset apart', () => {
  render(
    <HealthCoverage
      componentHealthStatuses={{
        api: 'healthy',
        db: 'unknown',
        cache: '',
        dns: 'not-applicable',
        queue: 'custom-verdict',
        worker: null,
        malformed: ' healthy ',
      }}
    />
  )
  expect(screen.getByText('5 of 7 with recorded verdicts')).toBeTruthy()
  expect(
    screen.getByText(
      '2 other · 1 unknown · 1 healthy · 1 not applicable · 2 no recorded verdict'
    )
  ).toBeTruthy()
})
