import { afterEach, expect, test } from 'bun:test'
import {
  cleanup,
  fireEvent,
  render,
  screen,
  within,
} from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import {
  alignmentResource,
  inputUpdateResource,
  openDeploymentResource,
  stateDriftResource,
  unappliedResource,
} from '@/components/__fixtures__/install-alignment'
import type { TAPIError, TInstallAlignmentResource } from '@/types'
import { DriftResourceCard } from '../DriftResourceCard'
import { ResourceAlignment } from '../ResourceAlignment/ResourceAlignment'
import { DriftedResources } from './DriftedResources'

afterEach(cleanup)

const resources = [
  inputUpdateResource({ id: 'cancelled', name: 'cancelled-inputs' }),
  openDeploymentResource({
    id: 'running',
    name: 'running-deploy',
    open_deployment: { id: 'wf-new', status: 'in-progress' },
    explanation: {
      summary: 'An older attempt was cancelled.',
      status: 'cancelled',
      workflow_id: 'wf-old',
    },
  }),
  stateDriftResource({ id: 'scan', name: 'scan-drift' }),
  alignmentResource({
    id: 'redeploy',
    name: 'aligned-redeploy',
    open_deployment: { id: 'wf-redeploy', status: 'queued' },
  }),
  alignmentResource({ id: 'off', name: 'disabled', category: 'disabled' }),
]

const View = ({
  items = resources,
  error,
  noun = 'components',
  search,
  pageSize = 4,
}: {
  items?: TInstallAlignmentResource[]
  error?: TAPIError
  noun?: string
  search?: string
  pageSize?: number
}) => (
  <MemoryRouter>
    <DriftedResources
      noun={noun}
      resources={items}
      error={error}
      search={search}
      pageSize={pageSize}
      renderCard={(resource) => (
        <DriftResourceCard
          resource={resource}
          deploymentHref={`/workflows/${resource.open_deployment?.id}`}
        />
      )}
    />
  </MemoryRouter>
)

test('ongoing resources appear once above drift, without hiding their unapplied changes', () => {
  render(<View />)
  const ongoing = screen.getByRole('region', { name: 'Ongoing deployments' })
  const drift = screen.getByRole('region', { name: 'Drifted' })
  expect(
    ongoing.compareDocumentPosition(drift) & Node.DOCUMENT_POSITION_FOLLOWING
  ).toBeTruthy()
  expect(
    within(ongoing).getByRole('heading', { name: 'running-deploy' })
  ).toBeTruthy()
  expect(
    within(ongoing).getByText('This component hasn’t been deployed yet.')
  ).toBeTruthy()
  expect(within(ongoing).getByText('In progress')).toBeTruthy()
  expect(within(ongoing).queryByText('Workflow cancelled')).toBeNull()
  expect(
    within(ongoing).getByRole('heading', { name: 'aligned-redeploy' })
  ).toBeTruthy()
  expect(
    within(ongoing).getByText('A deployment is pending or in progress.')
  ).toBeTruthy()
  expect(
    screen.getAllByRole('heading', { name: 'running-deploy' })
  ).toHaveLength(1)
  expect(
    within(ongoing)
      .getAllByRole('link', { name: 'View deployment' })
      .map((link) => link.getAttribute('href'))
  ).toEqual(['/workflows/wf-new', '/workflows/wf-redeploy'])
  expect(
    within(drift).getByRole('heading', { name: 'cancelled-inputs' })
  ).toBeTruthy()
  const filters = within(drift).getByRole('group', { name: 'Filter drift' })
  const all = within(filters).getByRole('button', { name: 'All 2' })
  const stateDrift = within(filters).getByRole('button', {
    name: 'State drift detected 1',
  })
  expect(all.getAttribute('aria-pressed')).toBe('true')
  fireEvent.click(stateDrift)
  expect(stateDrift.getAttribute('aria-pressed')).toBe('true')
  expect(all.getAttribute('aria-pressed')).toBe('false')
  expect(
    within(drift).queryByRole('heading', { name: 'cancelled-inputs' })
  ).toBeNull()
  expect(
    within(drift).getByRole('heading', { name: 'scan-drift' })
  ).toBeTruthy()
  fireEvent.click(
    within(filters).getByRole('button', { name: 'Changes not applied 1' })
  )
  expect(
    within(drift).queryByRole('heading', { name: 'scan-drift' })
  ).toBeNull()
  expect(
    within(drift).getByRole('heading', { name: 'cancelled-inputs' })
  ).toBeTruthy()
  fireEvent.click(all)
  expect(within(drift).getAllByTestId('drift-resource-card')).toHaveLength(2)
  expect(within(ongoing).getAllByTestId('drift-resource-card')).toHaveLength(2)
})

test('a stopped deployment returns to drift until its changes are actually applied', () => {
  const resource = openDeploymentResource({ id: 'changing', name: 'changing' })
  const { rerender } = render(<View items={[resource]} />)
  expect(
    screen.getByRole('region', { name: 'Ongoing deployments' })
  ).toBeTruthy()
  expect(screen.queryByTestId('alignment-summary')).toBeNull()
  expect(screen.queryByRole('region', { name: 'Drifted' })).toBeNull()
  for (const status of ['cancelled', 'success', 'error']) {
    rerender(
      <View
        items={[{ ...resource, open_deployment: { id: 'finished', status } }]}
      />
    )
    expect(
      screen.queryByRole('region', { name: 'Ongoing deployments' })
    ).toBeNull()
    expect(
      within(screen.getByRole('region', { name: 'Drifted' })).getByRole(
        'heading',
        { name: 'changing' }
      )
    ).toBeTruthy()
  }
  rerender(
    <View
      items={[
        alignmentResource({
          id: 'changing',
          name: 'changing',
          last_attempt: { id: 'failed-redeploy', status: 'error' },
        }),
      ]}
    />
  )
  expect(screen.queryByRole('region', { name: 'Drifted' })).toBeNull()
  expect(
    screen.getByText(
      'All components match the desired configuration and inputs. No changes are needed.'
    )
  ).toBeTruthy()
})

test('ongoing and drift pagination are independent and both reset when searching', () => {
  const items = Array.from({ length: 5 }, (_, i) => [
    openDeploymentResource({ id: `ongoing-${i}`, name: `ongoing-${i}` }),
    unappliedResource({ id: `drift-${i}`, name: `drift-${i}` }),
  ]).flat()
  const { rerender } = render(<View items={items} pageSize={2} />)
  fireEvent.click(screen.getByRole('button', { name: 'Next ongoing page' }))
  expect(screen.getByRole('heading', { name: 'ongoing-2' })).toBeTruthy()
  expect(screen.getByRole('heading', { name: 'drift-0' })).toBeTruthy()
  fireEvent.click(screen.getByRole('button', { name: 'Next drifted page' }))
  expect(screen.getByRole('heading', { name: 'drift-2' })).toBeTruthy()
  expect(screen.getByRole('heading', { name: 'ongoing-2' })).toBeTruthy()
  rerender(<View items={items} pageSize={2} search="-4" />)
  expect(screen.getByRole('heading', { name: 'ongoing-4' })).toBeTruthy()
  expect(screen.getByRole('heading', { name: 'drift-4' })).toBeTruthy()
  rerender(<View items={items} pageSize={2} />)
  expect(screen.getByRole('heading', { name: 'ongoing-0' })).toBeTruthy()
  expect(screen.getByRole('heading', { name: 'drift-0' })).toBeTruthy()
})

test('images describe recorded registry syncs separately from deployment drift', () => {
  render(
    <View
      noun="images"
      items={[
        unappliedResource({
          kind: 'image',
          id: 'failed-sync',
          name: 'api',
          reason: 'This image version hasn’t been synced.',
          plan: undefined,
          explanation: {
            summary: 'The latest image sync failed.',
            status: 'error',
          },
        }),
        alignmentResource({
          kind: 'image',
          id: 'queued-sync',
          name: 'worker',
          open_deployment: { id: 'wf-sync', status: 'queued' },
        }),
        openDeploymentResource({
          kind: 'image',
          id: 'retry-sync',
          name: 'retry-image',
          plan: undefined,
        }),
        unappliedResource({
          kind: 'image',
          id: 'cancelled-sync',
          name: 'cancelled-image',
          plan: undefined,
          explanation: {
            summary: 'The image sync was cancelled.',
            status: 'cancelled',
          },
        }),
      ]}
    />
  )
  const ongoing = screen.getByRole('region', { name: 'Syncing' })
  const unapplied = screen.getByRole('region', { name: 'Pending sync' })
  expect(
    ongoing.compareDocumentPosition(unapplied) &
      Node.DOCUMENT_POSITION_FOLLOWING
  ).toBeTruthy()
  expect(
    within(ongoing)
      .getByRole('link', { name: 'View sync' })
      .getAttribute('href')
  ).toBe('/workflows/wf-sync')
  expect(within(unapplied).getByText('Sync failed')).toBeTruthy()
  expect(
    within(unapplied).getByRole('heading', { name: 'retry-image' })
  ).toBeTruthy()
  expect(
    within(unapplied).getByRole('heading', { name: 'cancelled-image' })
  ).toBeTruthy()
  expect(within(unapplied).getByText('Sync cancelled')).toBeTruthy()
  fireEvent.click(
    within(unapplied).getAllByRole('button', {
      name: 'This image version hasn’t been synced.',
    })[0]
  )
  expect(
    within(unapplied).getByText('The latest image sync failed.')
  ).toBeTruthy()
  expect(screen.queryByText('State drift detected')).toBeNull()
  expect(screen.queryByText('Deployment failed')).toBeNull()
  expect(screen.queryByRole('link', { name: 'View deployment' })).toBeNull()
})

test('cached alignment remains visible with a refresh warning in summary, drift, ongoing and filtered states', () => {
  const error = {
    error: 'Internal Server Error',
    description: 'Overview unavailable',
    user_error: false,
  }
  const renderCached = (
    resource: TInstallAlignmentResource,
    search?: string
  ) => (
    <>
      <ResourceAlignment
        label="stack"
        resource={resource}
        card={<span>Cached stack</span>}
        error={error}
      />
      <View items={[resource]} error={error} search={search} />
    </>
  )
  const { rerender } = render(renderCached(alignmentResource()))
  expect(screen.getAllByTestId('alignment-summary')).toHaveLength(2)
  expect(screen.getAllByRole('alert')).toHaveLength(2)
  expect(
    screen.getAllByText(
      'Refreshing resources failed. Showing the last loaded results. Overview unavailable'
    )
  ).toHaveLength(2)

  for (const resource of [unappliedResource(), openDeploymentResource()]) {
    rerender(renderCached(resource))
    expect(screen.getByText('Cached stack')).toBeTruthy()
    expect(screen.getByRole('heading', { name: resource.name })).toBeTruthy()
    expect(screen.getAllByRole('alert')).toHaveLength(2)
  }

  rerender(renderCached(alignmentResource(), 'no-match'))
  expect(screen.getAllByRole('alert')).toHaveLength(2)
  expect(screen.getAllByTestId('alignment-summary')).toHaveLength(1)
})
