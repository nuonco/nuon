import { expect, test, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import { DeployPlan } from './DeployPlan'

vi.mock('../utils', async (og) => {
  const mod = await og<typeof import('../utils')>()
  return {
    ...mod,
    getFetchOpts: vi.fn().mockResolvedValue({
      cache: 'no-store',
      headers: {
        Authorization: 'Bearer test-token',
        'Content-Type': 'application/json',
        'X-Nuon-Org-ID': 'org-id',
      },
    }),
  }
})

test('renders a deploy plan', async () => {
  const JSX = await DeployPlan({
    orgId: 'test-org',
    installId: 'test-install',
    deployId: 'test-id',
  })
  render(JSX)
  expect(screen.getByTestId('deploy-plan')).toBeInTheDocument()
})
