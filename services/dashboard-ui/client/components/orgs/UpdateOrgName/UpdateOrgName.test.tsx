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
import { ToastContext } from '@/providers/toast-provider'
import type { TOrg } from '@/types'
import { UpdateOrgName } from './UpdateOrgNameContainer'

const orgId = 'org-acme'
const clients: QueryClient[] = []

afterEach(() => {
  cleanup()
  clients.splice(0).forEach((client) => client.clear())
  mock.restore()
})

function setup(name = 'acme') {
  const client = new QueryClient({
    defaultOptions: {
      queries: { retry: false, staleTime: Infinity },
      mutations: { retry: false },
    },
  })
  clients.push(client)
  const org = { id: orgId, name } as TOrg
  client.setQueryData(['org', orgId], org)
  client.setQueryData(['orgs', { offset: 0 }], [{ id: orgId, name }])
  const addToast = mock(() => {})
  render(
    <QueryClientProvider client={client}>
      <MemoryRouter>
        <OrgContext.Provider value={{ org, refresh: () => {} }}>
          <ToastContext.Provider value={{ addToast, removeToast: () => {} }}>
            <UpdateOrgName />
          </ToastContext.Provider>
        </OrgContext.Provider>
      </MemoryRouter>
    </QueryClientProvider>
  )
  return { client, addToast, org }
}

const nameInput = () => screen.getByLabelText('Organization name')
const saveButton = () => screen.getByRole('button', { name: 'Save' })

test('does not save when the name is unchanged', () => {
  const fetch = spyOn(globalThis, 'fetch')
  setup()
  expect(nameInput()).toHaveValue('acme')
  expect(saveButton()).toHaveAttribute('aria-disabled', 'true')
  fireEvent.click(saveButton())
  expect(fetch).not.toHaveBeenCalled()
})

test('requires a name', async () => {
  const fetch = spyOn(globalThis, 'fetch')
  setup()
  fireEvent.change(nameInput(), { target: { value: '   ' } })
  fireEvent.blur(nameInput())
  expect(await screen.findByText('Name is required')).toBeTruthy()
  expect(saveButton()).toBeDisabled()
  fireEvent.click(saveButton())
  expect(fetch).not.toHaveBeenCalled()
})

test('saves a trimmed name and updates the org cache', async () => {
  const fetch = spyOn(globalThis, 'fetch').mockResolvedValue(
    Response.json({ id: orgId, name: 'acme-west' })
  )
  const { client, addToast } = setup()
  fireEvent.change(nameInput(), { target: { value: '  acme-west  ' } })
  fireEvent.click(saveButton())

  await waitFor(() => expect(fetch).toHaveBeenCalledTimes(1))
  const [url, options] = fetch.mock.calls[0]
  expect(String(url)).toEndWith('/v1/orgs/current')
  expect(options?.method).toBe('PATCH')
  expect(options?.headers).toMatchObject({ 'X-Nuon-Org-ID': orgId })
  expect(JSON.parse(options?.body as string)).toEqual({ name: 'acme-west' })
  await waitFor(() => expect(addToast).toHaveBeenCalledTimes(1))
  expect(client.getQueryData<TOrg>(['org', orgId])?.name).toBe('acme-west')
  expect(client.getQueryState(['orgs', { offset: 0 }])?.isInvalidated).toBe(
    true
  )
})

test('shows a name conflict instead of the database error', async () => {
  spyOn(globalThis, 'fetch').mockResolvedValue(
    Response.json(
      {
        error:
          'unable to update org: ERROR: duplicate key value violates unique constraint "idx_org_name" (SQLSTATE 23505)',
        description: 'duplicate key',
        user_error: true,
      },
      { status: 409 }
    )
  )
  const { addToast } = setup()
  fireEvent.change(nameInput(), { target: { value: 'taken-name' } })
  fireEvent.click(saveButton())

  expect(
    await screen.findByText('An organization with this name already exists.')
  ).toBeTruthy()
  expect(screen.getByText('Choose a different name.')).toBeTruthy()
  expect(screen.queryByText(/duplicate key/)).toBeNull()
  expect(screen.queryByText(/idx_org_name/)).toBeNull()
  expect(nameInput()).toHaveValue('taken-name')
  expect(addToast).not.toHaveBeenCalled()
})

test('shows the permission error from the API', async () => {
  spyOn(globalThis, 'fetch').mockResolvedValue(
    Response.json(
      {
        error: 'this action requires write access to organization settings',
        description:
          'Your role (Read-only) does not have write access to organization settings. Ask an organization admin to assign a role that does.',
        user_error: true,
      },
      { status: 403 }
    )
  )
  setup()
  fireEvent.change(nameInput(), { target: { value: 'acme-west' } })
  fireEvent.click(saveButton())

  expect(
    await screen.findByText(
      'this action requires write access to organization settings'
    )
  ).toBeTruthy()
  expect(
    screen.getByText(
      'Your role (Read-only) does not have write access to organization settings. Ask an organization admin to assign a role that does.'
    )
  ).toBeTruthy()
})
