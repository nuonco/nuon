import { afterEach, expect, test } from 'bun:test'
import { cleanup, render, screen } from '@testing-library/react'
import { MiniDeploymentView } from './MiniDeploymentView'

afterEach(cleanup)

test('renders one segment for every install', () => {
  render(
    <MiniDeploymentView
      groups={[{ name: 'enterprise', installs: 3, hasSelector: false }]}
      installs={[
        {
          id: 'install-1',
          name: 'acme-prod',
          group: 'enterprise',
          runStatus: 'success',
          rolledOut: true,
        },
        {
          id: 'install-2',
          name: 'globex-prod',
          group: 'enterprise',
          runStatus: 'in-progress',
          rolledOut: false,
        },
        {
          id: 'install-3',
          name: 'initech-prod',
          group: 'enterprise',
          runStatus: 'failed',
          rolledOut: false,
        },
      ]}
    />
  )

  expect(screen.getByLabelText('acme-prod: Success')).toBeTruthy()
  expect(screen.getByLabelText('globex-prod: In progress')).toBeTruthy()
  expect(screen.getByLabelText('initech-prod: Failed')).toBeTruthy()
  expect(screen.getByLabelText('enterprise: 1/3 installs updated')).toBeTruthy()
})

test('constrains a single install group', () => {
  const { container } = render(
    <MiniDeploymentView
      groups={[{ name: 'canary', installs: 1, hasSelector: false }]}
      installs={[
        {
          id: 'install-1',
          name: 'staging-example',
          group: 'canary',
          runStatus: 'success',
          rolledOut: true,
        },
      ]}
    />
  )

  expect(container.querySelector('.w-12')).toBeTruthy()
})
