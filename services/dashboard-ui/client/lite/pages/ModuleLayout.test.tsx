import { afterEach, expect, test } from 'bun:test'
import { cleanup, render, screen } from '@testing-library/react'
import { createMemoryRouter, RouterProvider } from 'react-router'
import {
  ModulesContext,
  type IModulesContext,
} from '../providers/modules-provider'
import type { TModuleId } from '../utils/modules'
import { ModuleLayout } from './ModuleLayout'

afterEach(cleanup)

const modulesValue = (enabled: TModuleId[], ready = true): IModulesContext => ({
  enabled: new Set(enabled),
  has: (id) => enabled.includes(id),
  ready,
})

const renderGate = (
  value: IModulesContext,
  module: TModuleId | readonly TModuleId[]
) => {
  const router = createMemoryRouter(
    [
      {
        element: <ModuleLayout module={module} fallback={<p>Hidden</p>} />,
        children: [{ path: '/', element: <p>Module page</p> }],
      },
    ],
    { initialEntries: ['/'] }
  )

  render(
    <ModulesContext.Provider value={value}>
      <RouterProvider router={router} />
    </ModulesContext.Provider>
  )
}

test('renders the page while the module is enabled', () => {
  renderGate(modulesValue(['installs']), 'installs')

  expect(screen.getByText('Module page')).toBeTruthy()
  expect(screen.queryByText('Hidden')).toBeNull()
})

test('renders the fallback once the org resolves with the module hidden', () => {
  renderGate(modulesValue(['installs']), 'apps')

  expect(screen.getByText('Hidden')).toBeTruthy()
  expect(screen.queryByText('Module page')).toBeNull()
})

test('keeps the page mounted until the org has resolved', () => {
  renderGate(modulesValue([], false), 'apps')

  expect(screen.getByText('Module page')).toBeTruthy()
})

test('a group gate opens when any of its modules is enabled', () => {
  renderGate(modulesValue(['webhooks']), ['connections', 'webhooks'])

  expect(screen.getByText('Module page')).toBeTruthy()
})
