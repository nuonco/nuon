import { afterEach, expect, test } from 'bun:test'
import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router'
import { SurfacesProvider } from '@/providers/surfaces-provider'
import { ToastContext } from '@/providers/toast-provider'
import { TriggerBranchRunModalContainer } from './TriggerBranchRunModalContainer'

const realFetch = globalThis.fetch

afterEach(() => {
  cleanup()
  globalThis.fetch = realFetch
})

test.each([
  { name: 'defaults to reuse', toggles: 0, force: false },
  { name: 'requests all rebuilds when checked', toggles: 1, force: true },
  { name: 'returns to reuse when unchecked', toggles: 2, force: false },
])('$name', async ({ toggles, force }) => {
  let request: { path: string; body: unknown } | undefined
  globalThis.fetch = Object.assign(
    async (input: Parameters<typeof fetch>[0], options?: RequestInit) => {
      request = {
        path: new URL(String(input), 'https://example.com').pathname,
        body: JSON.parse(String(options?.body)),
      }
      return Response.json({ id: 'run-acme' }, { status: 201 })
    },
    { preconnect: () => {} }
  ) as typeof fetch

  const client = new QueryClient({
    defaultOptions: { mutations: { retry: false } },
  })

  render(
    <MemoryRouter>
      <QueryClientProvider client={client}>
        <ToastContext.Provider
          value={{ addToast: () => {}, removeToast: () => {} }}
        >
          <SurfacesProvider>
            <TriggerBranchRunModalContainer
              branch={{ id: 'branch-acme', name: 'production' }}
              currentConfig={{ id: 'config-acme' }}
              appId="app-acme"
              orgId="org-acme"
              planOnly={false}
              isVisible
            />
          </SurfacesProvider>
        </ToastContext.Provider>
      </QueryClientProvider>
    </MemoryRouter>
  )

  const checkbox = screen.getByRole('checkbox', {
    name: /Rebuild all components and sandbox/,
  })
  expect(checkbox).not.toBeChecked()
  for (let index = 0; index < toggles; index++) fireEvent.click(checkbox)

  fireEvent.click(screen.getByRole('button', { name: 'Trigger run' }))

  await waitFor(() =>
    expect(request).toEqual({
      path: '/v1/apps/app-acme/branches/branch-acme/runs',
      body: {
        config_id: 'config-acme',
        plan_only: false,
        force,
      },
    })
  )
})
