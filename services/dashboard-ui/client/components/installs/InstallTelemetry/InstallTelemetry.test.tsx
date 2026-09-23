import { afterEach, expect, mock, spyOn, test } from 'bun:test'
import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router'
import { ConfigContext } from '@/providers/config-provider'
import { InstallContext } from '@/providers/install-provider'
import { OrgContext } from '@/providers/org-provider'
import { SurfacesProvider } from '@/providers/surfaces-provider'
import { ToastContext } from '@/providers/toast-provider'
import { InstallSettingsPanelContent } from '@/components/installs/InstallSettingsPanel/InstallSettingsPanelContent'
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
  telemetryOverride = undefined as boolean | null | undefined,
  orgDefault = false,
  telemetryEndpoint = endpoint as unknown,
  runnerId = 'runner-acme',
  runnerStatus = 'active',
  loadSettings = false,
  loadStack = false,
  renderPanel = false,
  isManagedByConfig = false,
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
      override: telemetryOverride,
      org_default: orgDefault,
    })
  }
  const addToast = mock()
  const tree = (id = installId, status = runnerStatus) => (
    <QueryClientProvider client={client}>
      <ConfigContext.Provider
        value={{ apiUrl: '', appUrl: '', githubAppName: '', isByoc }}
      >
        <OrgContext.Provider
          value={{ org: { id: orgId, name: 'acme' }, refresh: () => {} }}
        >
          <InstallContext.Provider
            value={{
              install: {
                id,
                runner_id: runnerId,
                runner_status: status,
                metadata: {
                  managed_by: isManagedByConfig
                    ? 'nuon/cli/install-config'
                    : '',
                },
              },
              labelColors: {},
              refresh: () => {},
            }}
          >
            <ToastContext.Provider value={{ addToast, removeToast: () => {} }}>
              {renderPanel ? (
                <MemoryRouter>
                  <SurfacesProvider>
                    <InstallSettingsPanelContent />
                  </SurfacesProvider>
                </MemoryRouter>
              ) : (
                <InstallTelemetryContainer />
              )}
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
    rerenderInstall: (id: string, status = runnerStatus) =>
      view.rerender(tree(id, status)),
  }
}

test('hides telemetry outside BYOC even with an endpoint and enabled settings', () => {
  const fetch = spyOn(globalThis, 'fetch')
  setup({
    isByoc: false,
    telemetryEnabled: true,
    runnerId: '',
    renderPanel: true,
  })
  expect(screen.queryByText('Telemetry', { exact: true })).toBeNull()
  expect(screen.queryByRole('switch', { name: 'Enable telemetry' })).toBeNull()
  expect(fetch).not.toHaveBeenCalled()
})

test.each([false, true])(
  'settings panel owns the telemetry card and shows it only in BYOC (%p)',
  (isByoc) => {
    setup({ isByoc, runnerId: '', renderPanel: true })
    expect(
      screen.getByText('Configuration', { exact: true })
    ).toBeInTheDocument()
    if (isByoc) {
      const heading = screen.getByText('Telemetry', { exact: true })
      const card = heading.closest('.shadow-sm')!
      expect(card).toContainElement(
        screen.getByRole('switch', { name: 'Enable telemetry' })
      )
      expect(card.querySelector('.shadow-sm')).toBeNull()
      expect(
        screen.getByRole('switch', { name: 'Enable telemetry' })
      ).not.toBeDisabled()
    } else {
      expect(screen.queryByText('Telemetry', { exact: true })).toBeNull()
    }
  }
)

test.each(['', '  ', null, 123])(
  'shows but blocks enabling telemetry without an endpoint (%p)',
  (value) => {
    const fetch = spyOn(globalThis, 'fetch')
    setup({ telemetryEndpoint: value })
    const toggle = screen.getByRole('switch')
    expect(toggle).toBeDisabled()
    expect(toggle).toHaveAttribute('aria-checked', 'false')
    expect(screen.getByText(/Update the install stack/)).toBeInTheDocument()
    expect(screen.queryByText(/runner is not active/)).toBeNull()
    fireEvent.click(toggle)
    expect(fetch).not.toHaveBeenCalled()
  }
)

test('allows enabling telemetry when the endpoint exists but the runner is missing', async () => {
  const fetch = mockFetch(() =>
    Promise.resolve(Response.json({ enabled: true }))
  )
  setup({ runnerId: '' })
  const toggle = screen.getByRole('switch')
  expect(toggle).not.toBeDisabled()
  expect(screen.getByText(/runner is not active/)).toBeInTheDocument()
  expect(screen.queryByText(/Update the install stack/)).toBeNull()
  fireEvent.click(toggle)
  await waitFor(() => expect(fetch).toHaveBeenCalled())
  await waitFor(() => expect(toggle).toHaveAttribute('aria-checked', 'true'))
})

test.each(['offline', 'awaiting-heartbeat', 'unknown', ''])(
  'warns without blocking an inactive runner (%p)',
  async (runnerStatus) => {
    const fetch = mockFetch(() =>
      Promise.resolve(Response.json({ enabled: true }))
    )
    const { rerenderInstall } = setup({ runnerStatus })
    const toggle = screen.getByRole('switch')
    expect(toggle).not.toBeDisabled()
    expect(screen.getByText(/runner is not active/)).toBeInTheDocument()
    expect(screen.queryByText(/Update the install stack/)).toBeNull()
    fireEvent.click(toggle)
    await waitFor(() => expect(fetch).toHaveBeenCalled())
    await waitFor(() => expect(toggle).toHaveAttribute('aria-checked', 'true'))
    await waitFor(() => expect(toggle).not.toBeDisabled())

    rerenderInstall(installId, 'active')
    expect(toggle).not.toBeDisabled()
    expect(screen.queryByText(/runner is not active/)).toBeNull()
  }
)

test('shows both actions when the endpoint is missing and the runner is offline', () => {
  setup({ telemetryEndpoint: '', runnerStatus: 'offline' })
  expect(screen.getByText(/Update the install stack/)).toBeInTheDocument()
  expect(screen.getByText(/runner is not active/)).toBeInTheDocument()
  expect(screen.getByRole('switch')).toBeDisabled()
})

test('can disable telemetry with an offline runner without claiming the endpoint is missing while it loads', async () => {
  mockFetch((url) =>
    String(url).endsWith('/stack')
      ? new Promise<Response>(() => {})
      : Promise.resolve(Response.json({ enabled: false }))
  )
  const { client } = setup({
    telemetryEnabled: true,
    runnerStatus: 'offline',
    loadStack: true,
  })
  const toggle = screen.getByRole('switch')
  expect(toggle).not.toBeDisabled()
  expect(screen.getByText(/runner is not active/)).toBeInTheDocument()
  expect(screen.queryByText(/Update the install stack/)).toBeNull()
  fireEvent.click(toggle)
  await waitFor(() =>
    expect(
      client.getQueryData<TInstallTelemetrySettings>([
        'install-telemetry',
        orgId,
        installId,
      ])
    ).toEqual({ enabled: false })
  )
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
  expect(
    screen.getByRole('link', { name: 'View telemetry setup' })
  ).toBeInTheDocument()
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

test('shows loading while checking the endpoint and can retry a failed stack read', async () => {
  let finish!: (response: Response) => void
  mockFetch((url) =>
    String(url).endsWith('/stack')
      ? new Promise<Response>((resolve) => {
          finish = resolve
        })
      : Promise.resolve(Response.json({ enabled: false }))
  )
  setup({ loadStack: true })
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

test('resets an explicit disable to the enabled org default', async () => {
  const fetch = mockFetch(() =>
    Promise.resolve(
      Response.json({
        enabled: true,
        override: null,
        org_default: true,
      })
    )
  )
  setup({
    telemetryEnabled: false,
    telemetryOverride: false,
    orgDefault: true,
    runnerStatus: 'offline',
  })
  fireEvent.click(screen.getByRole('button', { name: 'Use org default' }))
  await waitFor(() => expect(fetch).toHaveBeenCalled())
  expect(JSON.parse(fetch.mock.calls[0][1]?.body as string)).toEqual({
    enabled: null,
  })
  await waitFor(() =>
    expect(screen.getByRole('switch')).toHaveAttribute('aria-checked', 'true')
  )
  expect(screen.getByText('Using org default')).toBeInTheDocument()
  expect(screen.queryByRole('button', { name: 'Use org default' })).toBeNull()
})

test.each([false, true])(
  'config-managed installs cannot change telemetry (%p)',
  (enabled) => {
    const fetch = spyOn(globalThis, 'fetch')
    setup({
      telemetryEnabled: enabled,
      telemetryOverride: enabled,
      isManagedByConfig: true,
    })
    const toggle = screen.getByRole('switch')
    const reset = screen.getByRole('button', { name: 'Use org default' })
    expect(toggle).toBeDisabled()
    expect(reset).toHaveAttribute('aria-disabled', 'true')
    fireEvent.click(toggle)
    fireEvent.click(reset)
    expect(fetch).not.toHaveBeenCalled()
  }
)

test('reset waits for the enabled org default endpoint check while disabling remains available', async () => {
  let finish!: (response: Response) => void
  const fetch = mockFetch(
    () =>
      new Promise<Response>((resolve) => {
        finish = resolve
      })
  )
  setup({
    telemetryEnabled: true,
    telemetryOverride: true,
    orgDefault: true,
    loadStack: true,
  })
  const reset = screen.getByRole('button', { name: 'Use org default' })
  expect(reset).toHaveAttribute('aria-disabled', 'true')
  expect(screen.getByRole('switch')).not.toBeDisabled()
  fireEvent.click(reset)
  expect(
    fetch.mock.calls.every(([, options]) => options?.method !== 'PATCH')
  ).toBe(true)
  finish(
    Response.json({
      install_stack_outputs: {
        data_contents: { telemetry_endpoint: endpoint },
      },
    })
  )
  await waitFor(() =>
    expect(
      screen.getByRole('button', { name: 'Use org default' })
    ).not.toHaveAttribute('aria-disabled', 'true')
  )
  fireEvent.click(screen.getByRole('button', { name: 'Use org default' }))
  await waitFor(() =>
    expect(
      fetch.mock.calls.some(
        ([, options]) =>
          options?.method === 'PATCH' &&
          options.body === JSON.stringify({ enabled: null })
      )
    ).toBe(true)
  )
  finish(Response.json({ enabled: true, override: null, org_default: true }))
  await screen.findByText('Using org default')
})

test('an enabled install reports stack failure once in a toast and can retry without blocking disable', async () => {
  let failStack = true
  mockFetch(() =>
    Promise.resolve(
      failStack
        ? Response.json({ error: 'Stack unavailable' }, { status: 503 })
        : Response.json({
            install_stack_outputs: {
              data_contents: { telemetry_endpoint: endpoint },
            },
          })
    )
  )
  const { addToast, rerenderInstall } = setup({
    telemetryEnabled: true,
    telemetryOverride: true,
    orgDefault: true,
    loadStack: true,
  })
  await waitFor(() => expect(addToast).toHaveBeenCalledTimes(1))
  expect(screen.queryByRole('alert')).toBeNull()
  expect(screen.getByRole('switch')).not.toBeDisabled()
  expect(
    screen.getByRole('button', { name: 'Use org default' })
  ).toHaveAttribute('aria-disabled', 'true')
  rerenderInstall(installId)
  expect(addToast).toHaveBeenCalledTimes(1)

  const toast = addToast.mock.calls[0][0]
  expect(toast.props.heading).toBe('Telemetry endpoint check failed')
  render(toast)
  expect(screen.getByText('Stack unavailable')).toBeInTheDocument()
  failStack = false
  fireEvent.click(screen.getByRole('button', { name: 'Retry settings' }))
  await waitFor(() =>
    expect(
      screen.getByRole('button', { name: 'Use org default' })
    ).not.toHaveAttribute('aria-disabled', 'true')
  )
  expect(addToast).toHaveBeenCalledTimes(1)
})

test.each([false, true])(
  'reset without an endpoint respects org default %s',
  (orgDefault) => {
    setup({ telemetryOverride: false, orgDefault, telemetryEndpoint: null })
    const button = screen.getByRole('button', { name: 'Use org default' })
    if (orgDefault) {
      const fetch = mockFetch(() => Promise.resolve(Response.json({})))
      expect(button).toHaveAttribute('aria-disabled', 'true')
      fireEvent.click(button)
      expect(fetch).not.toHaveBeenCalled()
    } else {
      expect(button).not.toBeDisabled()
    }
  }
)

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
