import { afterEach, expect, test } from 'bun:test'
import { cleanup, render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { DeployTimeline } from './DeployTimeline'

afterEach(cleanup)

test('missing deployment records do not claim the component has never deployed', () => {
  render(
    <MemoryRouter>
      <DeployTimeline
        deploys={[]}
        pagination={{ hasNext: false, offset: 0, limit: 10 }}
        orgId="org-1"
        installId="install-1"
        componentId="cmp-ledger"
        componentName="ledger"
        isLoading={false}
        error={null}
      />
    </MemoryRouter>
  )

  expect(screen.getByText('No deploys found')).toBeTruthy()
  expect(
    screen.getByText(
      'No deployment records are available for this component. Recorded deployments will appear here.'
    )
  ).toBeTruthy()
  expect(
    screen.queryByText('This component has not been deployed yet.')
  ).toBeNull()
})
