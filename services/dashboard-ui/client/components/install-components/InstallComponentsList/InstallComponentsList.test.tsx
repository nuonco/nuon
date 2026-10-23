import { afterEach, expect, test } from 'bun:test'
import {
  cleanup,
  fireEvent,
  render,
  screen,
  within,
} from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { Button } from '@/components/common/Button'
import { LatestDeployCard } from '@/components/install-components/LatestDeployCard'
import { InstallComponentsList } from './InstallComponentsList'

afterEach(cleanup)

test('not-compared components show their reason, not an active badge or empty deploy illustration, in both views', () => {
  let opened = 0
  for (const view of ['list', 'grid'] as const) {
    const { unmount } = render(
      <MemoryRouter>
        <InstallComponentsList
          aligned
          view={view}
          components={[
            {
              id: 'unavailable',
              name: 'ledger',
              href: '/components/ledger',
              alignment: 'unknown',
              alignmentReason: 'the applied deploy could not be determined',
              status: 'active',
              latestDeploy: <LatestDeployCard />,
              actions: <Button onClick={() => opened++}>View details</Button>,
            },
            {
              id: 'missing-alignment',
              name: 'worker',
              status: 'active',
              latestDeploy: <LatestDeployCard />,
            },
            {
              id: 'known',
              name: 'api',
              alignment: 'in-sync',
              status: 'active',
              latestDeploy: <span>Recorded deployment</span>,
            },
            {
              id: 'disabled',
              name: 'reports',
              alignment: 'disabled',
              enabled: false,
              latestDeploy: <LatestDeployCard />,
            },
          ]}
        />
      </MemoryRouter>
    )
    const unknown = within(screen.getByRole('region', { name: 'Not compared' }))
    expect(
      unknown.getByText('the applied deploy could not be determined')
    ).toBeTruthy()
    expect(
      unknown.getByText(
        'Alignment could not be determined from the available records.'
      )
    ).toBeTruthy()
    expect(
      unknown.getByRole('link', { name: 'ledger' }).getAttribute('href')
    ).toBe('/components/ledger')
    expect(unknown.queryByText('Active')).toBeNull()
    expect(screen.queryByText('No deploys yet')).toBeNull()
    expect(
      within(screen.getByRole('region', { name: 'In sync' })).getByText(
        'Recorded deployment'
      )
    ).toBeTruthy()
    expect(
      within(
        screen.getByRole('region', { name: 'Disabled components' })
      ).getByText('Disabled')
    ).toBeTruthy()
    fireEvent.click(unknown.getByRole('button', { name: 'View details' }))
    unmount()
  }
  expect(opened).toBe(2)
})

test('in-sync components keep deployment history and an input-verification tip until verified', () => {
  for (const view of ['list', 'grid'] as const) {
    const component = {
      id: 'api',
      name: 'api',
      alignment: 'in-sync' as const,
      alignmentReason:
        'Inputs not verified. Applied input values weren’t recorded for this deployment.',
      status: 'active',
      latestDeploy: <span>Recorded deployment</span>,
    }
    const content = (verified: boolean) => (
      <MemoryRouter>
        <InstallComponentsList
          aligned
          view={view}
          components={[
            {
              ...component,
              alignmentReason: verified ? undefined : component.alignmentReason,
            },
          ]}
        />
      </MemoryRouter>
    )
    const { rerender, unmount } = render(content(false))
    const inSync = within(screen.getByRole('region', { name: 'In sync' }))
    expect(inSync.getByText('Recorded deployment')).toBeTruthy()
    expect(screen.queryByRole('region', { name: 'Not compared' })).toBeNull()
    const info = inSync.getByRole('button', {
      name: 'Alignment information for api',
    })
    fireEvent.focus(info)
    expect(info).toHaveAccessibleDescription(component.alignmentReason)
    expect(
      screen.getByText(component.alignmentReason).closest('[role="tooltip"]')
    ).toBeTruthy()
    rerender(content(true))
    expect(
      screen.queryByRole('button', { name: 'Alignment information for api' })
    ).toBeNull()
    expect(inSync.getByText('Recorded deployment')).toBeTruthy()
    unmount()
  }
})
