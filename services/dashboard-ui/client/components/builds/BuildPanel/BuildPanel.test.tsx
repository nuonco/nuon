import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { afterEach, expect, test } from 'bun:test'
import { cleanup, render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { AppContext } from '@/providers/app-provider'
import { OrgContext } from '@/providers/org-provider'
import { SurfacesProvider } from '@/providers/surfaces-provider'
import { ToastProvider } from '@/providers/toast-provider'
import { BuildPanel } from './BuildPanel'

afterEach(cleanup)

class FakeEventSource {
  onmessage: ((event: MessageEvent) => void) | null = null
  onerror: (() => void) | null = null
  onopen: (() => void) | null = null
  addEventListener() {}
  removeEventListener() {}
  close() {}
}

test('shows the build page summary, logs, and trace', async () => {
  const originalFetch = globalThis.fetch
  const originalEventSource = globalThis.EventSource
  globalThis.EventSource = FakeEventSource as unknown as typeof EventSource
  const original = originalFetch
  globalThis.fetch = (async (input: RequestInfo | URL) => {
    const url = typeof input === 'string' ? input : input instanceof URL ? input.href : input.url
    if (url.includes('/builds/build-api')) {
      return new Response(
        JSON.stringify({
          id: 'build-api',
          org_id: 'org-acme',
          component_id: 'cmp-api',
          component_name: 'api',
          created_at: '2026-10-15T14:00:00Z',
          updated_at: '2026-10-15T14:05:00Z',
          status_description: 'Building the component image.',
          status_v2: {
            status: 'in-progress',
            status_human_description: 'Building image',
          },
          created_by: { email: 'jane@example.com' },
          vcs_connection_commit: {
            sha: 'a1b2c3d4e5f67890',
            message: 'Add cache component',
            author_name: 'Jane Doe',
            created_at: '2026-10-15T13:40:00Z',
          },
        }),
        { status: 200, headers: { 'content-type': 'application/json' } }
      )
    }
    if (url.includes('/components/cmp-api')) {
      return new Response(
        JSON.stringify({
          id: 'cmp-api',
          name: 'api',
          type: 'docker_build',
          app_id: 'app-acme',
        }),
        { status: 200, headers: { 'content-type': 'application/json' } }
      )
    }
    return original(input)
  }) as typeof fetch

  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })

  try {
    render(
      <QueryClientProvider client={client}>
        <OrgContext.Provider
          value={{ org: { id: 'org-acme' } as never, refresh: () => {} }}
        >
          <AppContext.Provider
            value={{
              app: { id: 'app-acme', name: 'acme-app', org_id: 'org-acme' } as never,
              labelColors: {},
              refresh: () => {},
            }}
          >
            <ToastProvider>
              <MemoryRouter>
                <SurfacesProvider>
                  <BuildPanel
                    isVisible
                    componentId="cmp-api"
                    buildId="build-api"
                    title="api"
                  />
                </SurfacesProvider>
              </MemoryRouter>
            </ToastProvider>
          </AppContext.Provider>
        </OrgContext.Provider>
      </QueryClientProvider>
    )

    await waitFor(() => {
      expect(screen.getByText('Add cache component')).toBeTruthy()
    })
    expect(screen.getByText('Building image')).toBeTruthy()
    expect(screen.getAllByText('In progress').length).toBeGreaterThan(0)
    expect(screen.getByRole('button', { name: 'Summary' })).toBeTruthy()
    expect(screen.getByText('Build source')).toBeTruthy()
    expect(screen.getByText('Execution')).toBeTruthy()
    expect(screen.getByRole('button', { name: 'Logs' })).toBeTruthy()
    expect(screen.getByRole('button', { name: 'Trace' })).toBeTruthy()
    expect(screen.getByText('jane@example.com')).toBeTruthy()
    expect(screen.getByText('Add cache component')).toBeTruthy()
  } finally {
    globalThis.fetch = originalFetch
    globalThis.EventSource = originalEventSource
  }
})
