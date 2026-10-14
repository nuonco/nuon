import { expect, test } from 'bun:test'
import type { TWorkflowStep } from '@/types'
import { formatElapsed, waitItems } from './progress'

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
