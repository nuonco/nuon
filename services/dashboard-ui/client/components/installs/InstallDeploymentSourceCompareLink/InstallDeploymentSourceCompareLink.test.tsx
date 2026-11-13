import type { ReactNode } from 'react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { afterEach, expect, test } from 'bun:test'
import { cleanup, render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { BuildChangeDetail } from '@/components/branches/ConfigChanges/BuildChangeDetail'
import type { TTemplateBuildChange } from '@/components/branches/ConfigChanges/config-changes'

const originalFetch = globalThis.fetch
afterEach(() => {
  cleanup()
  globalThis.fetch = originalFetch
})

const target = {
  orgId: 'org-acme',
  installId: 'ins-acme',
  workflowId: 'wf-deploy',
  kind: 'component' as const,
  name: 'acme-api',
}
const build: TTemplateBuildChange = {
  id: 'bld-selected',
  name: 'acme-api',
  kind: 'component',
  status: 'active',
  changeReason: 'source_changed',
  href: '/org-acme/apps/app-acme/components/cmp-api/builds/bld-selected',
}
const compareURL =
  'https://github.com/acme/example/compare/1111111111111111111111111111111111111111..2222222222222222222222222222222222222222'

const renderUI = (node: ReactNode) => {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  render(
    <QueryClientProvider client={client}>
      <MemoryRouter>{node}</MemoryRouter>
    </QueryClientProvider>
  )
  return client
}

test('branch build rows keep View build without fetching a source comparison', () => {
  let calls = 0
  globalThis.fetch = (async () => {
    calls++
    return Response.json({})
  }) as unknown as typeof fetch
  renderUI(<BuildChangeDetail build={build} />)
  expect(screen.getByRole('link', { name: 'View build' })).toHaveAttribute(
    'href',
    build.href
  )
  expect(screen.queryByRole('link', { name: /GitHub/ })).toBeNull()
  expect(calls).toBe(0)
})

test('install build rows link to the GitHub comparison of their builds', async () => {
  const calls: URL[] = []
  globalThis.fetch = (async (input: string, init: RequestInit) => {
    calls.push(new URL(input, 'http://example.test'))
    expect((init.headers as Record<string, string>)['X-Nuon-Org-ID']).toBe(
      target.orgId
    )
    return Response.json({
      compare_url: compareURL,
    })
  }) as unknown as typeof fetch
  renderUI(<BuildChangeDetail build={{ ...build, sourceDiffTarget: target }} />)
  const link = await screen.findByRole('link', {
    name: 'View changes on GitHub for acme-api',
  })
  expect(link).toHaveAttribute('href', compareURL)
  expect(link).toHaveAttribute('target', '_blank')
  expect(calls).toHaveLength(1)
  expect(calls[0].pathname).toBe(
    '/v1/installs/ins-acme/deployments/wf-deploy/source-diff'
  )
  expect(calls[0].searchParams.get('kind')).toBe('component')
  expect(calls[0].searchParams.get('name')).toBe('acme-api')
})

test('install build rows show no GitHub link when there is nothing to compare', async () => {
  globalThis.fetch = (async () => Response.json({})) as unknown as typeof fetch
  const client = renderUI(
    <BuildChangeDetail build={{ ...build, sourceDiffTarget: target }} />
  )
  await waitFor(() => expect(client.isFetching()).toBe(0))
  expect(
    client.getQueryState([
      'install-deployment-source-diff',
      ...Object.values(target),
    ])?.status
  ).toBe('success')
  expect(screen.queryByRole('link', { name: /GitHub/ })).toBeNull()
  expect(screen.getByRole('link', { name: 'View build' })).toBeTruthy()
})
