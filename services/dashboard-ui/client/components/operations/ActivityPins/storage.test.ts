import { beforeEach, expect, test } from 'bun:test'
import {
  MAX_ACTIVITY_PINS,
  activityPinStorageKey,
  readActivityPins,
  toggleActivityPin,
} from './storage'

const key = activityPinStorageKey('user-1', 'inst-1')

beforeEach(() => {
  localStorage.clear()
})

test('starts empty', () => {
  expect(readActivityPins(key)).toEqual([])
})

test('pins and unpins a shortcut', () => {
  toggleActivityPin(key, { kind: 'runbook', id: 'rb-1' })
  toggleActivityPin(key, { kind: 'action', id: 'act-1' })
  expect(readActivityPins(key)).toEqual([
    { kind: 'runbook', id: 'rb-1' },
    { kind: 'action', id: 'act-1' },
  ])

  toggleActivityPin(key, { kind: 'runbook', id: 'rb-1' })
  expect(readActivityPins(key)).toEqual([{ kind: 'action', id: 'act-1' }])
})

test('stops at four pins', () => {
  for (let index = 0; index < MAX_ACTIVITY_PINS + 2; index += 1) {
    toggleActivityPin(key, { kind: 'runbook', id: `rb-${index}` })
  }
  expect(readActivityPins(key)).toHaveLength(MAX_ACTIVITY_PINS)
  expect(readActivityPins(key).some((pin) => pin.id === 'rb-4')).toBe(false)
})

test('drops malformed entries', () => {
  localStorage.setItem(
    key,
    JSON.stringify([
      { kind: 'runbook', id: 'rb-1' },
      { kind: 'cron', id: 'nope' },
      { kind: 'action' },
      'rb-2',
    ])
  )
  expect(readActivityPins(key)).toEqual([{ kind: 'runbook', id: 'rb-1' }])
})
