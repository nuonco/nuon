import { afterEach, expect, mock, test } from 'bun:test'
import { cleanup, render, screen, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router'

let status = 404

const clearOrgSession = mock(() => {})
const setOrgSession = mock(() => {})

mock.module('@/lib/cookies', () => ({
  clearOrgSession,
  setOrgSession,
}))

mock.module('@/lib/ctl-api/orgs', () => ({
  getOrg: async () => {
    throw { status, error: 'nope' }
  },
}))

const { OrgProvider } = await import('./org-provider')

const clients: QueryClient[] = []

afterEach(() => {
  cleanup()
  clients.splice(0).forEach((client) => client.clear())
  clearOrgSession.mockClear()
  setOrgSession.mockClear()
})

function renderOrg() {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  clients.push(client)
  return render(
    <QueryClientProvider client={client}>
      <MemoryRouter>
        <OrgProvider orgId="org-stale">
          <div>org ready</div>
        </OrgProvider>
      </MemoryRouter>
    </QueryClientProvider>
  )
}

test('an org the account cannot open keeps loading and drops the session', async () => {
  status = 404
  renderOrg()

  await waitFor(() => expect(clearOrgSession).toHaveBeenCalled())
  expect(screen.queryByText('Not found')).toBeNull()
  expect(screen.queryByText('org ready')).toBeNull()
  expect(document.querySelector('.animate-spin')).not.toBeNull()
})

test('other org failures stay on the error page', async () => {
  status = 500
  renderOrg()

  expect(await screen.findByText('Something went wrong')).toBeTruthy()
  expect(clearOrgSession).not.toHaveBeenCalled()
  expect(screen.queryByText('org ready')).toBeNull()
})
