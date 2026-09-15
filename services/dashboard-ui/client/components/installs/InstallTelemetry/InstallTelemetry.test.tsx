import { afterEach, expect, mock, spyOn, test } from 'bun:test'
import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { ConfigContext } from '@/providers/config-provider'
import { InstallContext } from '@/providers/install-provider'
import { OrgContext } from '@/providers/org-provider'
import { ToastContext } from '@/providers/toast-provider'
import type { TInstallTelemetrySettings } from '@/types'
import { InstallTelemetryContainer } from './InstallTelemetryContainer'

const orgId = 'org-acme'
const installId = 'install-acme'
const endpoint = 'http://10.0.0.4:4318'
const clients: QueryClient[] = []

const mockFetch = (
  handler: (...args: Parameters<typeof fetch>) => Promise<Response>
) =>
  spyOn(globalThis, 'fetch').mockImplementation(
    Object.assign(handler, { preconnect: () => {} })
  )

afterEach(() => {
  cleanup()
  clients.splice(0).forEach((client) => client.clear())
  mock.restore()
})

function setup({
  isByoc = true,
  telemetryEnabled = false,
  telemetryEndpoint = endpoint as unknown,
  runnerId = 'runner-acme',
  loadSettings = false,
  loadStack = false,
} = {}) {
  const client = new QueryClient({
    defaultOptions: {
      queries: { retry: false, staleTime: Infinity },
      mutations: { retry: false },
    },
  })
  clients.push(client)
  if (!loadStack) {
    client.setQueryData(['install-stack', orgId, installId], {
      install_stack_outputs: {
        data_contents: { telemetry_endpoint: telemetryEndpoint },
      },
    })
  }
  if (!loadSettings) {
    client.setQueryData(['install-telemetry', orgId, installId], {
      enabled: telemetryEnabled,
    })
  }
  const addToast = mock()
  const tree = (id = installId) => (
    <QueryClientProvider client={client}>
      <ConfigContext.Provider
        value={{ apiUrl: '', appUrl: '', githubAppName: '', isByoc }}
      >
        <OrgContext.Provider
          value={{ org: { id: orgId, name: 'acme' }, refresh: () => {} }}
        >
          <InstallContext.Provider
            value={{
              install: { id, runner_id: runnerId },
              labelColors: {},
              refresh: () => {},
            }}
          >
            <ToastContext.Provider value={{ addToast, removeToast: () => {} }}>
              <InstallTelemetryContainer />
            </ToastContext.Provider>
          </InstallContext.Provider>
        </OrgContext.Provider>
      </ConfigContext.Provider>
    </QueryClientProvider>
  )
  const view = render(tree())
  return {
    client,
    addToast,
    rerenderInstall: (id: string) => view.rerender(tree(id)),
  }
}

test('hides telemetry outside BYOC even with an endpoint and enabled settings', () => {
  const fetch = spyOn(globalThis, 'fetch')
  setup({ isByoc: false, telemetryEnabled: true })
  expect(screen.queryByRole('switch')).toBeNull()
  expect(fetch).not.toHaveBeenCalled()
})

test.each(['', '  ', null, 123])(
  'shows but blocks enabling telemetry without an endpoint (%p)',
  (value) => {
    const fetch = spyOn(globalThis, 'fetch')
    setup({ telemetryEndpoint: value })
    expect(screen.getByText('Telemetry', { exact: true })).toBeInTheDocument()
    const toggle = screen.getByRole('switch')
    expect(toggle).toBeDisabled()
    expect(toggle).toHaveAttribute('aria-checked', 'false')
    expect(screen.getByText(/Update the install stack/)).toBeInTheDocument()
    fireEvent.click(toggle)
    expect(fetch).not.toHaveBeenCalled()
  }
)

test('shows but blocks enabling telemetry when the runner is missing', () => {
  const fetch = spyOn(globalThis, 'fetch')
  setup({ runnerId: '' })
  const toggle = screen.getByRole('switch')
  expect(toggle).toBeDisabled()
  fireEvent.click(toggle)
  expect(fetch).not.toHaveBeenCalled()
})

test('shows the saved setting when BYOC, runner, and endpoint are present', () => {
  setup()
  expect(
    screen.getByRole('switch', { name: 'Enable telemetry' })
  ).toHaveAttribute('aria-checked', 'false')
  expect(
    screen.getByRole('link', { name: 'View telemetry setup' })
  ).toHaveAttribute('href', 'https://docs.nuon.co/guides/byoc/telemetry')
})

test.each([false, true])(
  'saves the opposite of %p with scoped credentials and prevents duplicate pending writes',
  async (initial) => {
    let finish!: (response: Response) => void
    const fetch = mockFetch((_url, options) =>
      options?.method === 'PATCH'
        ? new Promise<Response>((resolve) => {
            finish = resolve
          })
        : Promise.resolve(Response.json({ enabled: !initial }))
    )
    const { client, addToast } = setup({ telemetryEnabled: initial })
    const toggle = screen.getByRole('switch')
    fireEvent.click(toggle)
    await waitFor(() => expect(toggle).toBeDisabled())
    expect(toggle).toHaveAttribute('aria-checked', String(initial))
    fireEvent.click(toggle)
    expect(fetch).toHaveBeenCalledTimes(1)
    const [url, options] = fetch.mock.calls[0]
    expect(url).toBe(`/v1/installs/${installId}/telemetry`)
    expect(options?.headers).toMatchObject({ 'X-Nuon-Org-ID': orgId })
    expect(options?.credentials).toBe('include')
    expect(JSON.parse(options?.body as string)).toEqual({ enabled: !initial })
    finish(Response.json({ enabled: !initial }))
    await waitFor(() =>
      expect(toggle).toHaveAttribute('aria-checked', String(!initial))
    )
    await waitFor(() => expect(toggle).not.toBeDisabled())
    expect(
      client.getQueryData<TInstallTelemetrySettings>([
        'install-telemetry',
        orgId,
        installId,
      ])
    ).toEqual({ enabled: !initial })
    expect(addToast.mock.calls[0][0].props.heading).toBe(
      initial ? 'Telemetry disabled' : 'Telemetry enabled'
    )
  }
)

test('keeps disable available when a previously enabled install loses its endpoint and runner', async () => {
  mockFetch(() => Promise.resolve(Response.json({ enabled: false })))
  setup({ telemetryEnabled: true, telemetryEndpoint: '', runnerId: '' })
  expect(screen.getByRole('switch')).toHaveAttribute('aria-checked', 'true')
  expect(
    screen.getByText(/Forwarding can still be disabled/)
  ).toBeInTheDocument()
  fireEvent.click(screen.getByRole('switch'))
  await waitFor(() =>
    expect(screen.getByRole('switch')).toHaveAttribute('aria-checked', 'false')
  )
  expect(screen.getByRole('switch')).toBeDisabled()
  expect(screen.getByText('Telemetry', { exact: true })).toBeInTheDocument()
})

test('preserves saved state and reports a failed write', async () => {
  spyOn(globalThis, 'fetch').mockResolvedValue(
    Response.json({ error: 'Permission denied' }, { status: 403 })
  )
  const { addToast } = setup({ telemetryEnabled: true })
  fireEvent.click(screen.getByRole('switch'))
  await waitFor(() => expect(addToast).toHaveBeenCalledTimes(1))
  expect(screen.getByRole('switch')).toHaveAttribute('aria-checked', 'true')
  expect(addToast.mock.calls[0][0].props.heading).toBe(
    'Telemetry update failed'
  )
  expect(addToast.mock.calls[0][0].props.children.props.children).toBe(
    'Permission denied'
  )
})

test('does not offer a toggle while loading or after a read error, and can retry', async () => {
  let finish!: (response: Response) => void
  const fetch = mockFetch(
    () =>
      new Promise<Response>((resolve) => {
        finish = resolve
      })
  )
  setup({ loadSettings: true })
  expect(screen.getByText('Loading settings...')).toBeInTheDocument()
  expect(screen.queryByRole('switch')).toBeNull()
  finish(Response.json({ error: 'Temporarily unavailable' }, { status: 503 }))
  await screen.findByRole('alert')
  expect(screen.queryByRole('switch')).toBeNull()
  fetch.mockImplementation(
    Object.assign(
      (url: string | URL | Request) =>
        Promise.resolve(
          Response.json(
            String(url).endsWith('/stack')
              ? {
                  install_stack_outputs: {
                    data_contents: { telemetry_endpoint: endpoint },
                  },
                }
              : { enabled: true }
          )
        ),
      { preconnect: () => {} }
    )
  )
  fireEvent.click(screen.getByRole('button', { name: 'Retry settings' }))
  await waitFor(() =>
    expect(screen.getByRole('switch')).toHaveAttribute('aria-checked', 'true')
  )
})

test('keeps the card visible while checking the endpoint and can retry a failed stack read', async () => {
  let finish!: (response: Response) => void
  mockFetch((url) =>
    String(url).endsWith('/stack')
      ? new Promise<Response>((resolve) => {
          finish = resolve
        })
      : Promise.resolve(Response.json({ enabled: false }))
  )
  setup({ loadStack: true })
  expect(screen.getByText('Telemetry', { exact: true })).toBeInTheDocument()
  expect(screen.getByText('Loading settings...')).toBeInTheDocument()
  expect(screen.queryByRole('switch')).toBeNull()
  expect(screen.queryByText(/Update the install stack/)).toBeNull()
  finish(Response.json({ error: 'Stack unavailable' }, { status: 503 }))
  await screen.findByRole('alert')
  expect(screen.queryByRole('switch')).toBeNull()
  fireEvent.click(screen.getByRole('button', { name: 'Retry settings' }))
  finish(
    Response.json({
      install_stack_outputs: {
        data_contents: { telemetry_endpoint: endpoint },
      },
    })
  )
  await waitFor(() => expect(screen.getByRole('switch')).not.toBeDisabled())
  expect(screen.getByRole('switch')).toHaveAttribute('aria-checked', 'false')
})

test('a pending save updates only its original install after navigation', async () => {
  let finish!: (response: Response) => void
  mockFetch(
    () =>
      new Promise<Response>((resolve) => {
        finish = resolve
      })
  )
  const { client, addToast, rerenderInstall } = setup()
  fireEvent.click(screen.getByRole('switch'))
  await waitFor(() => expect(screen.getByRole('switch')).toBeDisabled())
  client.setQueryData(['install-stack', orgId, 'install-other'], {
    install_stack_outputs: { data_contents: { telemetry_endpoint: endpoint } },
  })
  client.setQueryData(['install-telemetry', orgId, 'install-other'], {
    enabled: false,
  })
  rerenderInstall('install-other')
  finish(Response.json({ enabled: true }))
  await waitFor(() => expect(addToast).toHaveBeenCalledTimes(1))
  expect(
    client.getQueryData<TInstallTelemetrySettings>([
      'install-telemetry',
      orgId,
      installId,
    ])
  ).toEqual({
    enabled: true,
  })
  expect(
    client.getQueryData<TInstallTelemetrySettings>([
      'install-telemetry',
      orgId,
      'install-other',
    ])
  ).toEqual({ enabled: false })
  expect(screen.getByRole('switch')).toHaveAttribute('aria-checked', 'false')
})
