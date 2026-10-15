import { afterEach, expect, test } from 'bun:test'
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router'
import { AuthContext } from '@/providers/auth-provider'
import { OrgContext } from '@/providers/org-provider'
import { SurfacesProvider } from '@/providers/surfaces-provider'
import { ToastContext } from '@/providers/toast-provider'
import type { TRunner, TRunnerSettings } from '@/types'
import { ManagementDropdown } from './ManagementDropdown'

const runner = { id: 'runner-1', runner_group: { type: 'install' } } as TRunner
const settings = {
  container_image_tag: 'v1.2.3',
  binary_version: 'v0.5.0',
} as TRunnerSettings

afterEach(cleanup)

test('install runner updates open outside the runner provider', () => {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  })

  render(
    <QueryClientProvider client={client}>
      <MemoryRouter>
        <AuthContext.Provider
          value={{
            user: null,
            isAuthenticated: true,
            isAdmin: false,
            isNuonEmployee: false,
            isLoading: false,
            error: null,
            demoMode: false,
            toggleDemoMode: () => {},
          }}
        >
          <OrgContext.Provider
            value={{ org: { id: 'org-1' } as never, refresh: () => {} }}
          >
            <ToastContext.Provider
              value={{ addToast: () => {}, removeToast: () => {} }}
            >
              <SurfacesProvider>
                <ManagementDropdown
                  runner={runner}
                  settings={settings}
                  isInstallRunner
                />
              </SurfacesProvider>
            </ToastContext.Provider>
          </OrgContext.Provider>
        </AuthContext.Provider>
      </MemoryRouter>
    </QueryClientProvider>
  )

  for (const [label, placeholder] of [
    ['Update management version', 'management version'],
    ['Update instance version', 'instance version'],
  ] as const) {
    fireEvent.click(screen.getByRole('button', { name: 'Manage runner' }))
    fireEvent.click(screen.getByRole('button', { name: label }))
    expect(screen.getByPlaceholderText(placeholder)).toBeTruthy()
  }
})
