import type { ReactNode } from 'react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { afterEach, expect, test } from 'bun:test'
import { cleanup, render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { AppContext } from '@/providers/app-provider'
import { OrgContext } from '@/providers/org-provider'
import type { TInstallWorkflowStep } from '@/types'
import { PlanInstallChangesContainer } from './PlanInstallChangesContainer'
import { WarmPlanInstallChanges } from './WarmPlanInstallChanges'

afterEach(cleanup)

const step = {
  id: 'step_plan',
  name: 'plan install group: default',
  execution_type: 'approval',
  install_workflow_id: 'wf_184',
  status: { status: 'approval-awaiting' },
  approval: { id: 'apr_1' },
} as TInstallWorkflowStep

const contentsUrl =
  '/api/orgs/org-acme/workflows/wf_184/steps/step_plan/approvals/apr_1/contents'

const renderWithClient = (client: QueryClient, node: ReactNode) =>
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
          <MemoryRouter>{node}</MemoryRouter>
        </AppContext.Provider>
      </OrgContext.Provider>
    </QueryClientProvider>
  )

test('loads plan changes before the review panel opens', async () => {
  const calls: string[] = []
  const originalFetch = globalThis.fetch
  globalThis.fetch = (async (input: RequestInfo | URL) => {
    const url =
      typeof input === 'string'
        ? input
        : input instanceof URL
          ? input.href
          : input.url
    calls.push(url)
    if (url.includes('/contents')) {
      return Response.json({
        install_group: 'default',
        installs: [
          {
            install_id: 'ins_serval',
            install_name: 'Serval',
            new_app_config_id: 'cfg_new',
          },
        ],
      })
    }
    return Response.json({})
  }) as typeof fetch

  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })

  try {
    const warmup = renderWithClient(
      client,
      <WarmPlanInstallChanges steps={[step]} />
    )
    await waitFor(() => {
      expect(calls.filter((url) => url.includes(contentsUrl))).toHaveLength(1)
    })
    warmup.unmount()

    renderWithClient(client, <PlanInstallChangesContainer step={step} />)
    expect(await screen.findByText('Serval')).toBeTruthy()
    expect(calls.filter((url) => url.includes(contentsUrl))).toHaveLength(1)
  } finally {
    globalThis.fetch = originalFetch
    client.clear()
  }
})
