import { describe, expect, test } from 'bun:test'
import type { TWorkflowStep } from '@/types'
import { formatElapsed, installGate, waitItems } from './progress'

const step = (name: string, status: string) => ({ name, status: { status } }) as TWorkflowStep

const defs = [
  { label: 'a', done: true },
  { label: 'b', step: 'fetch commit' },
  { label: 'c', step: 'fetch app config' },
  { label: 'd' },
]

test('first unfinished item is active', () => {
  expect(waitItems(defs).map((i) => i.state)).toEqual(['done', 'active', 'pending', 'pending'])
})

test('items before a started step count as done', () => {
  const states = waitItems(defs, [step('fetch app config', 'in-progress')]).map((i) => i.state)
  expect(states).toEqual(['done', 'done', 'active', 'pending'])
})

test('a gap after a finished step shows the next item as active', () => {
  const steps = [step('fetch commit', 'success'), step('fetch app config', 'success')]
  expect(waitItems(defs, steps).map((i) => i.state)).toEqual(['done', 'done', 'done', 'active'])
})

test('an error stops the active marker', () => {
  const states = waitItems(defs, [step('fetch commit', 'error')]).map((i) => i.state)
  expect(states).toEqual(['done', 'error', 'pending', 'pending'])
})

test('formatElapsed', () => {
  expect(formatElapsed(0)).toBe('0:00')
  expect(formatElapsed(65_400)).toBe('1:05')
})

describe('installGate', () => {
  const builds = step('building components and sandbox', 'success')
  test('waits for the plan step after builds finish', () => {
    expect(installGate({ hasRuns: true, steps: [builds, step('plan install group: default', 'pending')] })).toEqual({
      buildsReady: true,
      readyToCreate: false,
    })
  })
  test('creates once the plan step has picked its installs', () => {
    for (const status of ['auto-skipped', 'approval-awaiting']) {
      expect(installGate({ hasRuns: true, steps: [builds, step('plan install group: default', status)] }).readyToCreate).toBe(true)
    }
  })
  test('waits for builds', () => {
    expect(installGate({ hasRuns: true, steps: [step('building components and sandbox', 'in-progress')] }).readyToCreate).toBe(false)
  })
  test('a run with no plan step is ready when it finishes', () => {
    expect(installGate({ hasRuns: true, runStatus: 'success', steps: [builds] }).readyToCreate).toBe(true)
    expect(installGate({ hasRuns: true, runStatus: 'in-progress', steps: [builds] }).readyToCreate).toBe(false)
  })
  test('a config synced without a branch run has nothing to wait for', () => {
    expect(installGate({ hasRuns: false, configFromBranch: false }).readyToCreate).toBe(true)
  })
})
