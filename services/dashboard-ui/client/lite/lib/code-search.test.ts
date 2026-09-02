import { describe, expect, test } from 'bun:test'
import type { KeyboardEvent } from 'react'
import { matchNavKeyDown } from './code-search'

const press = (key: string, shiftKey = false) => {
  const calls: number[] = []
  let prevented = false
  const event = {
    key,
    shiftKey,
    preventDefault: () => {
      prevented = true
    },
  } as KeyboardEvent<HTMLInputElement>

  matchNavKeyDown(3, (index) => calls.push(index))(event)

  return { calls, prevented }
}

describe('matchNavKeyDown', () => {
  test('steps forward on ArrowDown and Enter', () => {
    expect(press('ArrowDown').calls).toEqual([4])
    expect(press('Enter').calls).toEqual([4])
  })

  test('steps back on ArrowUp and Shift+Enter', () => {
    expect(press('ArrowUp').calls).toEqual([2])
    expect(press('Enter', true).calls).toEqual([2])
  })

  test('takes the caret keys so the input does not move it', () => {
    expect(press('ArrowDown').prevented).toBe(true)
    expect(press('ArrowUp').prevented).toBe(true)
  })

  test('leaves every other key alone', () => {
    for (const key of ['a', 'Escape', 'Tab', 'ArrowLeft', 'ArrowRight']) {
      const { calls, prevented } = press(key)
      expect(calls).toEqual([])
      expect(prevented).toBe(false)
    }
  })
})
