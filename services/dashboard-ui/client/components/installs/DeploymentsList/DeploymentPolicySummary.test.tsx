import { expect, test } from 'bun:test'
import { render, screen } from '@testing-library/react'
import { DeploymentPolicySummary } from './DeploymentPolicySummary'

test('counts skipped applies alongside warnings and omits zero counts', () => {
  const skipped = {
    policy: { deny_count: 0, warn_count: 0, skipped_by_policy: true },
  }
  const { rerender } = render(
    <DeploymentPolicySummary history steps={[skipped, {}]} />
  )
  expect(screen.getByText('1 apply skipped by policy')).toBeDefined()
  expect(screen.queryByText(/warning/)).toBeNull()

  rerender(
    <DeploymentPolicySummary
      history
      steps={[
        skipped,
        skipped,
        { policy: { deny_count: 0, warn_count: 3 } },
        {},
      ]}
    />
  )
  expect(screen.getByText('2 applies skipped by policy')).toBeDefined()
  expect(screen.getByText('3 policy warnings')).toBeDefined()

  rerender(
    <DeploymentPolicySummary
      history
      steps={[{ policy: { deny_count: 2, warn_count: 1 } }]}
    />
  )
  expect(screen.getByText('Blocked by policy')).toBeDefined()
  expect(screen.getByText('2 violations')).toBeDefined()
  expect(screen.getByText('1 warning')).toBeDefined()
  expect(screen.queryByText(/skipped by policy/)).toBeNull()
})
