import { afterEach, expect, test } from 'bun:test'
import { cleanup, render, screen, within } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import type { TAppConfig } from '@/types'
import { ImageConsumers } from './ImageConsumers'
import { InstallImagesList } from './InstallImagesList'

afterEach(cleanup)

test('image consumers link to configured output users, not ordering-only or similarly named dependencies', () => {
  const appConfig: TAppConfig = {
    component_config_connections: [
      {
        component_id: 'image',
        component_name: 'api-image',
        type: 'external_image',
      },
      {
        component_id: 'chart',
        component_name: 'api',
        type: 'helm_chart',
        references: [
          'component|api-image|repository|',
          'component|api-image|tag|',
        ],
      },
      {
        component_id: 'manifest',
        component_name: 'jobs',
        type: 'kubernetes_manifest',
        references: ['component|api-image|tag|'],
      },
      {
        component_id: 'terraform',
        component_name: 'lambda',
        type: 'terraform_module',
        references: ['component|api-image|uri|'],
      },
      {
        component_id: 'ordered',
        component_name: 'ordered-only',
        type: 'helm_chart',
        component_dependency_ids: ['image'],
      },
      {
        component_id: 'different',
        component_name: 'different-image',
        type: 'helm_chart',
        references: [
          'component|api-image-worker|tag|',
          'inputs|api-image|value|',
        ],
      },
    ],
  }
  const { rerender } = render(
    <MemoryRouter>
      <ImageConsumers
        appConfig={appConfig}
        componentId="image"
        componentHref={(id) => `/install/components/${id}`}
      />
    </MemoryRouter>
  )
  expect(screen.getByText('Used by')).toBeTruthy()
  expect(
    screen
      .getAllByRole('link')
      .map((link) => [link.textContent, link.getAttribute('href')])
  ).toEqual([
    ['api', '/install/components/chart'],
    ['jobs', '/install/components/manifest'],
    ['lambda', '/install/components/terraform'],
  ])
  rerender(
    <MemoryRouter>
      <ImageConsumers componentId="image" componentHref={(id) => id} />
    </MemoryRouter>
  )
  expect(screen.queryByText('Used by')).toBeNull()
})

test('missing image sync evidence stays neutral outside the synced section, and torn-down images say so, without a not-compared category', () => {
  render(
    <MemoryRouter>
      <InstallImagesList
        aligned
        view="list"
        images={[
          {
            id: 'known',
            name: 'api-image',
            alignment: 'in-sync',
            status: 'active',
            image: null,
          },
          {
            id: 'unknown',
            name: 'worker-image',
            alignment: 'unknown',
            status: 'active',
            image: null,
          },
          {
            id: 'torn-down',
            name: 'ledger-image',
            alignment: 'torn-down',
            status: 'inactive',
            image: null,
          },
        ]}
      />
    </MemoryRouter>
  )
  const synced = screen.getByRole('region', { name: 'Synced' })
  expect(
    within(synced).getByRole('heading', { name: 'api-image' })
  ).toBeTruthy()
  expect(
    within(synced).queryByRole('heading', { name: 'worker-image' })
  ).toBeNull()
  expect(screen.getAllByText('Sync status unavailable')).toHaveLength(1)
  expect(
    within(
      screen.getByRole('heading', { name: 'ledger-image' }).parentElement!
    ).getByText('Torn down')
  ).toBeTruthy()
  expect(
    screen.getAllByText(
      'Delivery of the configured version can’t be confirmed from the available records.'
    )
  ).toHaveLength(1)
  expect(screen.queryByText('Not compared')).toBeNull()
  expect(screen.queryByText('Needs attention')).toBeNull()
})
