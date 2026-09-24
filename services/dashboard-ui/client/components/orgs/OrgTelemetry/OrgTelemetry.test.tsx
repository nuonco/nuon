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
import { OrgContext } from '@/providers/org-provider'
import { SurfacesProvider } from '@/providers/surfaces-provider'
import { ToastContext } from '@/providers/toast-provider'
import type { TOrg } from '@/types'
import { OrgTelemetryButton } from './OrgTelemetryContainer'

const orgId = 'org-acme'
const clients: QueryClient[] = []

afterEach(() => {
  cleanup()
  clients.splice(0).forEach((client) => client.clear())
  mock.restore()
})

function setup({
  enabled = false,
  role = 'org_admin',
  roleOrgId = orgId,
} = {}) {
  const client = new QueryClient({
    defaultOptions: {
      queries: { retry: false, staleTime: Infinity },
      mutations: { retry: false },
    },
  })
  clients.push(client)
  const org = { id: orgId, name: 'acme', telemetry: { enabled } }
  client.setQueryData(['account'], {
    roles: [{ org_id: roleOrgId, role_type: role }],
  })
  client.setQueryData(['org', orgId], org)
  client.setQueryData(['install-telemetry', orgId, 'install-a'], { enabled })
  client.setQueryData(['install-telemetry', 'org-other', 'install-b'], {
    enabled: false,
  })
  const addToast = mock()
  render(
    <QueryClientProvider client={client}>
      <MemoryRouter>
        <OrgContext.Provider value={{ org, refresh: () => {} }}>
          <ToastContext.Provider value={{ addToast, removeToast: () => {} }}>
            <SurfacesProvider>
              <OrgTelemetryButton />
            </SurfacesProvider>
          </ToastContext.Provider>
        </OrgContext.Provider>
      </MemoryRouter>
    </QueryClientProvider>
  )
  return { client, addToast }
}

test.each([
  { role: 'org_read_only', roleOrgId: orgId },
  { role: 'org_admin', roleOrgId: 'org-other' },
])(
  'hides telemetry management without admin access to the current org (%p)',
  (options) => {
    setup(options)
    expect(
      screen.queryByRole('button', { name: 'Manage telemetry' })
    ).toBeNull()
  }
)

test.each([false, true])(
  'saves the changed org default from %p and invalidates only its install telemetry',
  async (enabled) => {
    const fetch = spyOn(globalThis, 'fetch').mockResolvedValue(
      Response.json({
        id: orgId,
        name: 'acme',
        telemetry: { enabled: !enabled },
      })
    )
    const { client, addToast } = setup({ enabled })
    fireEvent.click(screen.getByRole('button', { name: 'Manage telemetry' }))
    const toggle = screen.getByRole('switch', {
      name: 'Enable telemetry by default',
    })
    expect(toggle).toHaveAttribute('aria-checked', String(enabled))
    const save = screen.getByRole('button', { name: 'Save settings' })
    expect(save).toHaveAttribute('aria-disabled', 'true')
    fireEvent.click(save)
    expect(fetch).not.toHaveBeenCalled()

    fireEvent.click(toggle)
    fireEvent.click(screen.getByRole('button', { name: 'Save settings' }))
    await waitFor(() => expect(fetch).toHaveBeenCalledTimes(1))
    const [url, options] = fetch.mock.calls[0]
    expect(String(url)).toEndWith('/v1/orgs/current')
    expect(options?.method).toBe('PATCH')
    expect(options?.headers).toMatchObject({ 'X-Nuon-Org-ID': orgId })
    expect(JSON.parse(options?.body as string)).toEqual({
      telemetry: { enabled: !enabled },
    })
    await waitFor(() => expect(addToast).toHaveBeenCalledTimes(1))
    expect(client.getQueryData<TOrg>(['org', orgId])?.telemetry?.enabled).toBe(
      !enabled
    )
    expect(
      client.getQueryState(['install-telemetry', orgId, 'install-a'])
        ?.isInvalidated
    ).toBe(true)
    expect(
      client.getQueryState(['install-telemetry', 'org-other', 'install-b'])
        ?.isInvalidated
    ).toBe(false)
    await waitFor(() => expect(screen.queryAllByRole('dialog').length).toBe(0))
  }
)

test('keeps the attempted setting after a rejected save and supports retry', async () => {
  const fetch = spyOn(globalThis, 'fetch')
    .mockResolvedValueOnce(
      Response.json(
        { error: 'Only org admins can change the telemetry default' },
        { status: 403 }
      )
    )
    .mockResolvedValueOnce(
      Response.json({ id: orgId, telemetry: { enabled: true } })
    )
  const { client, addToast } = setup()
  fireEvent.click(screen.getByRole('button', { name: 'Manage telemetry' }))
  fireEvent.click(screen.getByRole('switch'))
  fireEvent.click(screen.getByRole('button', { name: 'Save settings' }))
  await screen.findByText('Only org admins can change the telemetry default')
  expect(screen.getByRole('switch')).toHaveAttribute('aria-checked', 'true')
  expect(client.getQueryData<TOrg>(['org', orgId])?.telemetry?.enabled).toBe(
    false
  )
  expect(addToast).not.toHaveBeenCalled()
  fireEvent.click(screen.getByRole('button', { name: 'Save settings' }))
  await waitFor(() => expect(addToast).toHaveBeenCalledTimes(1))
  expect(fetch).toHaveBeenCalledTimes(2)
})

test('disables editing during a pending save', async () => {
  let resolve: (response: Response) => void
  const fetch = spyOn(globalThis, 'fetch').mockImplementation(
    Object.assign(
      () =>
        new Promise<Response>((done) => {
          resolve = done
        }),
      { preconnect: () => {} }
    )
  )
  setup()
  fireEvent.click(screen.getByRole('button', { name: 'Manage telemetry' }))
  fireEvent.click(screen.getByRole('switch'))
  fireEvent.click(screen.getByRole('button', { name: 'Save settings' }))
  const saving = await screen.findByRole('button', { name: 'Saving...' })
  expect(saving).toBeDisabled()
  expect(screen.getByRole('switch')).toBeDisabled()
  fireEvent.click(saving)
  expect(fetch).toHaveBeenCalledTimes(1)
  resolve!(Response.json({ id: orgId, telemetry: { enabled: true } }))
  await waitFor(() => expect(screen.queryAllByRole('dialog').length).toBe(0))
})

test('cancel does not persist a changed toggle', async () => {
  const fetch = spyOn(globalThis, 'fetch')
  setup()
  fireEvent.click(screen.getByRole('button', { name: 'Manage telemetry' }))
  fireEvent.click(screen.getByRole('switch'))
  fireEvent.click(screen.getByRole('button', { name: 'Cancel' }))
  await waitFor(() => expect(screen.queryAllByRole('dialog').length).toBe(0))
  expect(fetch).not.toHaveBeenCalled()
})
