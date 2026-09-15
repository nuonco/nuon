import { describe, expect, test } from 'bun:test'
import type { TApp, TComponent } from '@/types'
import {
  appInstallBadge,
  hasActiveComponentBuild,
  hasRunnerConfig,
  latestComponentIds,
  shouldDefaultStackOnly,
} from './app-install-readiness'

const app = (overrides: Partial<TApp> = {}): TApp =>
  ({
    id: 'app1',
    name: 'payments',
    ...overrides,
  }) as TApp

const component = (status?: string): TComponent =>
  ({
    id: 'cmp1',
    latest_build: status
      ? { status_v2: { status } }
      : undefined,
  }) as TComponent

describe('app-install-readiness', () => {
  test('treats a missing runner type as not provisionable', () => {
    const noRunner = app({ runner_config: {} })
    expect(hasRunnerConfig(noRunner)).toBe(false)
    expect(appInstallBadge(noRunner)).toBe('not-provisionable')
    expect(shouldDefaultStackOnly(noRunner)).toBe(false)
  })

  test('uses latest config component_ids for the no-components badge', () => {
    const withConfig = app({
      runner_config: { app_runner_type: 'aws' },
      app_configs: [{ component_ids: [] }],
    })
    expect(latestComponentIds(withConfig)).toEqual([])
    expect(appInstallBadge(withConfig)).toBe('no-components')
    expect(shouldDefaultStackOnly(withConfig)).toBe(true)
  })

  test('waits for components before showing a no-builds badge', () => {
    const withComponents = app({
      runner_config: { app_runner_type: 'aws' },
      app_configs: [{ component_ids: ['cmp1'] }],
    })
    expect(appInstallBadge(withComponents)).toBeUndefined()
    expect(appInstallBadge(withComponents, [component()])).toBe(
      'no-component-builds'
    )
    expect(shouldDefaultStackOnly(withComponents, [component('error')])).toBe(
      true
    )
  })

  test('does not force stack-only when any component has an active build', () => {
    const withComponents = app({
      runner_config: { app_runner_type: 'aws' },
      app_configs: [{ component_ids: ['cmp1', 'cmp2'] }],
    })
    const components = [component(), component('active')]
    expect(hasActiveComponentBuild(components)).toBe(true)
    expect(appInstallBadge(withComponents, components)).toBeUndefined()
    expect(shouldDefaultStackOnly(withComponents, components)).toBe(false)
  })
})
