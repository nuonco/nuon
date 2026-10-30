import { afterEach, expect, mock, spyOn, test } from 'bun:test'
import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
  waitForElementToBeRemoved,
  within,
} from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router'
import { ConfigContext } from '@/providers/config-provider'
import { InstallContext } from '@/providers/install-provider'
import { OrgContext } from '@/providers/org-provider'
import { SurfacesProvider } from '@/providers/surfaces-provider'
import { ToastContext } from '@/providers/toast-provider'
import { InstallSettingsPanelContent } from '@/components/installs/InstallSettingsPanel/InstallSettingsPanelContent'
import type { TInstall, TToken } from '@/types'
import { InstallTelemetryContainer } from './InstallTelemetryContainer'

const orgId = 'org-acme'
const installId = 'install-acme'
const clients: QueryClient[] = []

afterEach(() => {
  cleanup()
  clients.splice(0).forEach((client) => client.clear())
  mock.restore()
})

function setup({
  override = false as boolean | null,
  orgDefault = true,
  relayConfigured = true,
  managed = false,
  adminOrg = orgId,
  collectorStatus = 200,
  collectorExists = true,
  configId = 'config-acme',
  writeStatus = 200,
  renderPanel = false,
  tokens = [] as TToken[],
  loadSettings = false,
} = {}) {
  const client = new QueryClient({
    defaultOptions: {
      queries: { retry: false, staleTime: Infinity },
      mutations: { retry: false },
    },
  })
  clients.push(client)
  const install: TInstall = {
    id: installId,
    name: 'acme-production',
    app_id: 'app-acme',
    runner_status: 'offline',
    install_config: configId ? { id: configId } : undefined,
    metadata: { managed_by: managed ? 'nuon/cli/install-config' : '' },
  }
  const settings = () => ({
    enabled: override ?? orgDefault,
    override,
    org_default: orgDefault,
    relay_configured: relayConfigured,
  })
  if (!loadSettings)
    client.setQueryData(['install-telemetry', orgId, installId], settings())
  client.setQueryData(['account'], {
    roles: [{ org_id: adminOrg, role_type: 'org_admin' }],
  })
  const writes: { url: string; options: RequestInit; body: any }[] = []
  let status = collectorStatus
  const fetch = spyOn(globalThis, 'fetch').mockImplementation(
    Object.assign(
      async (
        input: Parameters<typeof globalThis.fetch>[0],
        options: RequestInit = {}
      ) => {
        const url = String(input)
        if (options.method && options.method !== 'GET') {
          const body = options.body
            ? JSON.parse(String(options.body))
            : undefined
          writes.push({ url, options, body })
          if (writeStatus !== 200)
            return Response.json(
              { error: 'Permission denied' },
              { status: writeStatus }
            )
          if (url.includes('/configs')) {
            override = body.telemetry.enabled
            return Response.json({
              id: configId || 'config-created',
              telemetry_enabled: override,
            })
          }
          if (url.endsWith('/telemetry/collector')) {
            collectorExists = options.method !== 'DELETE'
            return options.method === 'DELETE'
              ? new Response(null, { status: 204 })
              : Response.json({ account_id: 'account-collector' })
          }
          if (url.endsWith('/tokens')) {
            tokens = [
              ...tokens,
              {
                id: 'token-new',
                account_id: 'account-collector',
                expires_at: '2099-01-01T00:00:00Z',
              },
            ]
            return Response.json({ token: 'example-one-time-token' })
          }
          if (url.includes('/tokens/')) {
            const tokenId = url.split('/tokens/').at(-1)
            tokens = tokens.filter((token) => token.id !== tokenId)
            return new Response(null, { status: 204 })
          }
          throw new Error(`Unexpected write: ${url}`)
        }
        if (url.endsWith('/telemetry'))
          return loadSettings
            ? Response.json({ error: 'Settings unavailable' }, { status: 503 })
            : Response.json(settings())
        if (url.endsWith('/telemetry/collector'))
          return status !== 200 || !collectorExists
            ? Response.json(
                { error: 'Account unavailable' },
                { status: status !== 200 ? status : 404 }
              )
            : Response.json({ account_id: 'account-collector' })
        if (url.includes('/components?'))
          return Response.json(
            new URL(url, 'https://example.com').searchParams.get('labels') ===
              'nuon.co/telemetry-collector:true'
              ? [{ component_id: 'component-collector', enabled: false }]
              : []
          )
        if (url.includes('/tokens?')) {
          const accountId = new URL(url, 'https://example.com').searchParams.get(
            'account_id'
          )
          return Response.json(
            accountId ? tokens.filter((t) => t.account_id === accountId) : tokens
          )
        }
        if (url.endsWith('/account'))
          return Response.json({
            roles: [{ org_id: adminOrg, role_type: 'org_admin' }],
          })
        throw new Error(`Unexpected read: ${url}`)
      },
      { preconnect: () => {} }
    )
  )
  const addToast = mock()
  const tree = (currentInstall = install) => (
    <MemoryRouter>
      <QueryClientProvider client={client}>
        <ConfigContext.Provider
          value={{
            apiUrl: '',
            runnerApiUrl: 'https://runner-api.example.com',
            appUrl: '',
            githubAppName: '',
            isByoc: false,
          }}
        >
          <OrgContext.Provider
            value={{ org: { id: orgId, name: 'acme' }, refresh: () => {} }}
          >
            <InstallContext.Provider
              value={{
                install: currentInstall,
                labelColors: {},
                refresh: () => {},
              }}
            >
              <ToastContext.Provider
                value={{ addToast, removeToast: () => {} }}
              >
                <SurfacesProvider>
                  {renderPanel ? (
                    <InstallSettingsPanelContent />
                  ) : (
                    <InstallTelemetryContainer />
                  )}
                </SurfacesProvider>
              </ToastContext.Provider>
            </InstallContext.Provider>
          </OrgContext.Provider>
        </ConfigContext.Provider>
      </QueryClientProvider>
    </MemoryRouter>
  )
  const view = render(tree())
  return {
    client,
    writes,
    fetch,
    addToast,
    fixCollector: () => {
      status = 200
    },
    rerenderInstall: () =>
      view.rerender(tree({ ...install, id: 'install-other' })),
  }
}

async function open() {
  fireEvent.click(screen.getByRole('button', { name: 'Manage telemetry' }))
  fireEvent.click(
    await screen.findByRole('button', {
      name: /In-cluster collector \(optional\)/,
    })
  )
  await screen.findByText('Nuon service account')
}

async function choose(label: string) {
  fireEvent.click(
    screen.getByRole('combobox', { name: 'Telemetry forwarding' })
  )
  fireEvent.click(await screen.findByRole('option', { name: label }))
}

test.each([
  { option: 'On', expected: true },
  { option: 'Use org default (On)', expected: null },
])(
  'saves $option without a runner endpoint and without creating access or deploying',
  async ({ option, expected }) => {
    const { writes } = setup()
    await open()
    expect(writes).toHaveLength(0)
    await choose(option)
    fireEvent.click(screen.getByRole('button', { name: 'Save settings' }))
    await screen.findByText(
      'On · ' + (expected === null ? 'Org default' : 'Install override')
    )
    expect(writes).toMatchObject([
      {
        url: `/v1/installs/${installId}/configs/config-acme`,
        options: {
          method: 'PATCH',
          credentials: 'include',
          headers: { 'X-Nuon-Org-ID': orgId },
        },
        body: { telemetry: { enabled: expected } },
      },
    ])
  }
)

test('saving Off keeps collector identity and issued tokens', async () => {
  const { writes } = setup({
    override: true,
    tokens: [{ id: 'token-old', account_id: 'account-collector' }],
  })
  await open()
  await screen.findByText('token-old')
  await choose('Off')
  fireEvent.click(screen.getByRole('button', { name: 'Save settings' }))
  await screen.findByText('Off · Install override')
  expect(writes.map(({ body }) => body)).toEqual([
    { telemetry: { enabled: false } },
  ])
})

test('setup remains visible without relay configuration and failed reads are not absence', async () => {
  const state = setup({
    relayConfigured: false,
    collectorStatus: 403,
    renderPanel: true,
  })
  expect(screen.getByText('Relay setup required')).toBeInTheDocument()
  await open()
  await screen.findByText('Unknown — load failed')
  expect(screen.queryByText('Not created')).toBeNull()
  expect(
    screen.getByRole('button', { name: 'Create bootstrap credential' })
  ).toHaveAttribute('aria-disabled', 'true')
  state.fixCollector()
  fireEvent.click(screen.getByRole('button', { name: 'Retry setup details' }))
  await screen.findByText('Created', { exact: true })
  expect(state.writes).toHaveLength(0)
})

test('the first bootstrap credential creates a missing account without a separate setup action', async () => {
  const { writes } = setup({ collectorExists: false })
  await open()
  await screen.findByText('Not created')
  expect(writes).toHaveLength(0)
  fireEvent.click(
    screen.getByRole('button', { name: 'Create bootstrap credential' })
  )
  const dialog = within(screen.getAllByRole('dialog').at(-1)!)
  fireEvent.click(dialog.getByRole('button', { name: 'Create credential' }))
  await dialog.findByText('example-one-time-token')
  expect(writes.find(({ url }) => url.endsWith('/tokens'))?.body).toMatchObject(
    {
      duration: '8760h',
      invalidate: false,
    }
  )
  fireEvent.click(dialog.getByRole('button', { name: 'Done' }))
  await screen.findByText('Created', { exact: true })
  await screen.findByText('token-new')
  expect(screen.queryByText('Not created')).toBeNull()
})

test('config-managed policy is read-only without hiding access management', async () => {
  const { writes } = setup({ managed: true })
  await open()
  expect(
    screen.getByRole('combobox', { name: 'Telemetry forwarding' })
  ).toBeDisabled()
  expect(
    await screen.findByRole('button', { name: 'Create bootstrap credential' })
  ).toBeInTheDocument()
  expect(writes).toHaveLength(0)
})

test('org admin in a different org cannot manage collector tokens or query org token inventory', async () => {
  const { fetch } = setup({ adminOrg: 'org-other' })
  await open()
  expect(
    screen.queryByRole('button', { name: 'Create bootstrap credential' })
  ).toBeNull()
  expect(
    fetch.mock.calls.some(([url]) => String(url).includes('/tokens'))
  ).toBe(false)
})

test('a missing install config is created once and subsequent edits use its returned ID', async () => {
  const { writes } = setup({ configId: '' })
  await open()
  await choose('On')
  fireEvent.click(screen.getByRole('button', { name: 'Save settings' }))
  await screen.findByText('On · Install override')
  await waitFor(() => expect(screen.queryByRole('dialog')).toBeNull())
  await open()
  await choose('Off')
  fireEvent.click(screen.getByRole('button', { name: 'Save settings' }))
  await screen.findByText('Off · Install override')
  expect(writes).toMatchObject([
    { url: `/v1/installs/${installId}/configs`, options: { method: 'POST' } },
    {
      url: `/v1/installs/${installId}/configs/config-created`,
      options: { method: 'PATCH' },
    },
  ])
})

test('failed policy writes keep saved settings and report an inline error', async () => {
  setup({ writeStatus: 403 })
  await open()
  await choose('On')
  fireEvent.click(screen.getByRole('button', { name: 'Save settings' }))
  await screen.findByText('Permission denied')
  expect(screen.getByText('Off · Install override')).toBeInTheDocument()
})

test('failed settings reads keep the entry point visible and do not allow policy changes', async () => {
  setup({ loadSettings: true, renderPanel: true })
  await screen.findByText('Settings unavailable')
  expect(
    screen.getByRole('button', { name: 'Retry settings' })
  ).toBeInTheDocument()
  expect(screen.queryByRole('combobox')).toBeNull()
})

test('a captured modal saves only its original install after navigation', async () => {
  const { client, writes, rerenderInstall } = setup()
  await open()
  client.setQueryData(['install-telemetry', orgId, 'install-other'], {
    enabled: false,
    override: false,
  })
  rerenderInstall()
  await choose('On')
  fireEvent.click(screen.getByRole('button', { name: 'Save settings' }))
  await waitFor(() => expect(writes).toHaveLength(1))
  expect(writes[0].url).toBe(`/v1/installs/${installId}/configs/config-acme`)
  expect(
    client.getQueryData(['install-telemetry', orgId, 'install-other'])
  ).toMatchObject({ enabled: false })
})

test('rotation keeps old tokens and offers a one-time input token without caching it', async () => {
  const { writes, client } = setup({
    tokens: [
      {
        id: 'token-old',
        account_id: 'account-collector',
      },
      { id: 'token-unrelated', account_id: 'account-other' },
    ],
  })
  await open()
  await screen.findByText('token-old')
  expect(screen.queryByText('token-unrelated')).toBeNull()
  fireEvent.click(
    screen.getByRole('button', { name: 'Create bootstrap credential' })
  )
  const dialog = within(screen.getAllByRole('dialog').at(-1)!)
  fireEvent.click(dialog.getByRole('combobox', { name: 'Expires after' }))
  fireEvent.click(await screen.findByRole('option', { name: '90 days' }))
  fireEvent.click(dialog.getByRole('button', { name: 'Create credential' }))
  await dialog.findByText('example-one-time-token')
  expect(writes).toMatchObject([
    {
      url: `/v1/installs/${installId}/telemetry/collector`,
      options: { method: 'POST' },
    },
    {
      url: '/v1/tokens',
      body: { duration: '2160h', invalidate: false },
    },
  ])
  expect(
    JSON.stringify(
      client
        .getQueryCache()
        .getAll()
        .map((q) => q.state.data)
    )
  ).not.toContain('example-one-time-token')
  expect(
    JSON.stringify(
      client
        .getMutationCache()
        .getAll()
        .map((m) => m.state)
    )
  ).not.toContain('example-one-time-token')
  fireEvent.click(dialog.getByRole('button', { name: 'Done' }))
  await screen.findByText('token-new')
  expect(screen.queryByText('example-one-time-token')).toBeNull()
  expect(screen.getByText('token-old')).toBeInTheDocument()
  fireEvent.click(
    screen.getAllByRole('button', { name: 'Revoke credential' })[0]
  )
  const revoke = within(screen.getAllByRole('dialog').at(-1)!)
  expect(writes).toHaveLength(2)
  fireEvent.click(revoke.getByRole('button', { name: 'Revoke credential' }))
  await waitForElementToBeRemoved(() =>
    screen.queryByText('token-old', { exact: true })
  )
  expect(writes.at(-1)).toMatchObject({
    url: '/v1/tokens/token-old',
    options: { method: 'DELETE' },
  })
})

test('removing collector access requires confirmation and does not alter forwarding or deployments', async () => {
  const { writes } = setup({ override: true })
  await open()
  fireEvent.click(
    await screen.findByRole('button', { name: 'Remove collector access' })
  )
  const confirmation = within(screen.getAllByRole('dialog').at(-1)!)
  expect(writes).toHaveLength(0)
  fireEvent.click(
    confirmation.getByRole('button', { name: 'Remove collector access' })
  )
  await screen.findByText('Not created', { exact: true })
  expect(screen.getByText('On · Install override')).toBeInTheDocument()
  expect(writes).toMatchObject([
    {
      url: `/v1/installs/${installId}/telemetry/collector`,
      options: { method: 'DELETE' },
    },
  ])
})
