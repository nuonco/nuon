import { expect, test } from 'bun:test'
import {
  appendSurfaceValue,
  parseSurfaceValue,
  removeSurfaceValue,
  truncateSurfaceValues,
} from './surface-url'

test('parses keys with optional resource IDs', () => {
  expect(parseSurfaceValue('install-settings')).toEqual({
    key: 'install-settings',
    value: 'install-settings',
  })
  expect(parseSurfaceValue('deploy:dpl-123')).toEqual({
    key: 'deploy',
    resourceId: 'dpl-123',
    value: 'deploy:dpl-123',
  })
})

test('appends a surface without changing feature query parameters', () => {
  const params = appendSurfaceValue(
    '?tab=logs&panel=install:ins-123',
    'panel',
    'deploy:dpl-123'
  )

  expect(params.get('tab')).toBe('logs')
  expect(params.getAll('panel')).toEqual(['install:ins-123', 'deploy:dpl-123'])
})

test('truncates a stack from the selected layer upward', () => {
  const params = truncateSurfaceValues(
    '?tab=logs&panel=install&panel=deploy&panel=step',
    'panel',
    1
  )

  expect(params.get('tab')).toBe('logs')
  expect(params.getAll('panel')).toEqual(['install'])
})

test('removes one occurrence without changing later values', () => {
  const params = removeSurfaceValue(
    '?tab=logs&panel=install&panel=deploy&panel=step',
    'panel',
    1
  )

  expect(params.get('tab')).toBe('logs')
  expect(params.getAll('panel')).toEqual(['install', 'step'])
})
