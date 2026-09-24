import { afterEach, describe, expect, mock, test } from 'bun:test'
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { MODULES, presetModules, type TModuleId } from '../../../utils/modules'
import { ModuleManager, type IModuleState } from './ModuleManager'

afterEach(cleanup)

const states = (
  enabled: ReadonlySet<TModuleId>,
  pinned: readonly TModuleId[] = []
): IModuleState[] =>
  MODULES.map((module) => ({
    module,
    enabled: enabled.has(module.id),
    pinned: pinned.includes(module.id),
    registered: true,
  }))

const noop = () => {}

describe('ModuleManager', () => {
  test('names the preset the current selection matches', () => {
    const view = render(
      <ModuleManager
        modules={states(presetModules('install-operations'))}
        onToggle={noop}
        onPreset={noop}
      />
    )

    expect(screen.getAllByText('Install operations').length).toBeGreaterThan(0)
    expect(screen.getByText('2 of 9')).toBeTruthy()

    view.rerender(
      <ModuleManager
        modules={states(new Set(['apps']))}
        onToggle={noop}
        onPreset={noop}
      />
    )

    expect(screen.getAllByText('Custom').length).toBeGreaterThan(0)
  })

  test('routes a switch to onToggle with the module id', () => {
    const onToggle = mock(() => {})
    render(
      <ModuleManager
        modules={states(presetModules('full'))}
        onToggle={onToggle}
        onPreset={noop}
      />
    )

    fireEvent.click(screen.getByRole('switch', { name: 'Webhooks module' }))

    expect(onToggle).toHaveBeenCalledWith('webhooks', false)
  })

  test('locks every switch while a change is pending', () => {
    render(
      <ModuleManager
        modules={states(presetModules('full'))}
        onToggle={noop}
        onPreset={noop}
        pending="apps"
      />
    )

    for (const control of screen.getAllByRole('switch')) {
      expect(control.hasAttribute('disabled')).toBe(true)
    }
  })

  test('counts pinned modules and keeps them off', () => {
    render(
      <ModuleManager
        modules={states(presetModules('full'), ['apps'])}
        onToggle={noop}
        onPreset={noop}
      />
    )

    expect(screen.getByText('Pinned by deployment')).toBeTruthy()
    expect(
      screen
        .getByRole('switch', { name: 'Apps module' })
        .hasAttribute('disabled')
    ).toBe(true)
  })

  test('surfaces a failed save as a banner with the API message', () => {
    render(
      <ModuleManager
        modules={states(presetModules('full'))}
        onToggle={noop}
        onPreset={noop}
        error={{ error: 'admin email is not authorized' }}
      />
    )

    expect(screen.getByText('Module update failed')).toBeTruthy()
    expect(screen.getByText('admin email is not authorized')).toBeTruthy()
  })

  test('offers the deployment config for the current selection', () => {
    render(
      <ModuleManager
        modules={states(presetModules('install-operations'))}
        onToggle={noop}
        onPreset={noop}
      />
    )

    expect(screen.getByText('Deployment config')).toBeTruthy()
    expect(screen.getByRole('button', { name: 'Copy code' })).toBeTruthy()
  })
})
