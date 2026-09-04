import { describe, expect, test } from 'bun:test'
import type { KeyboardEvent } from 'react'
import { parseDiffFromFile } from '@pierre/diffs'
import type { FileContents } from '@pierre/diffs/react'
import { diffMatches, lineMatches, matchNavKeyDown } from './code-search'

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

describe('lineMatches', () => {
  const value = 'apiVersion: apps/v1\nkind: Deployment\nimage: app:1.2.0'

  test('returns one-based line numbers', () => {
    expect(lineMatches(value, 'app')).toEqual([1, 3])
  })

  test('ignores case and surrounding whitespace', () => {
    expect(lineMatches(value, '  DEPLOYMENT ')).toEqual([2])
  })

  test('returns nothing for an empty query', () => {
    expect(lineMatches(value, '   ')).toEqual([])
  })
})

describe('diffMatches', () => {
  const parse = (before: string, after: string) => {
    const file = (contents: string): FileContents => ({
      name: 'change.yaml',
      contents,
      lang: 'yaml',
    })
    return parseDiffFromFile(file(before), file(after))
  }

  const body = (replicas: number, tail: string) => `apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-server
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: '{"apiVersion":"apps/v1"}'
spec:
  replicas: ${replicas}
${tail}`

  const before = body(1, 'status: apps-pending\n')
  const after = body(3, 'status: apps-ready\n')

  test('counts an unchanged line once instead of once per side', () => {
    expect(diffMatches(parse(before, after), 'kubectl')).toEqual([
      { lineNumber: 6, side: 'additions' },
    ])
  })

  test('counts both sides of a changed line', () => {
    expect(diffMatches(parse(before, after), 'replicas')).toEqual([
      { lineNumber: 8, side: 'deletions' },
      { lineNumber: 8, side: 'additions' },
    ])
  })

  test('walks unchanged lines that sit outside the hunks', () => {
    expect(diffMatches(parse(before, after), 'apps')).toEqual([
      { lineNumber: 1, side: 'additions' },
      { lineNumber: 6, side: 'additions' },
      { lineNumber: 9, side: 'deletions' },
      { lineNumber: 9, side: 'additions' },
    ])
  })

  test('returns nothing for an empty query', () => {
    expect(diffMatches(parse(before, after), '')).toEqual([])
  })
})
