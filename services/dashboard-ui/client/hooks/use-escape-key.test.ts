import { renderHook } from '@testing-library/react'
import { describe, expect, test } from 'bun:test'
import { useEscapeKey } from './use-escape-key'

describe('useEscapeKey', () => {
  test('calls onEscape when enabled', () => {
    let called = 0
    renderHook(() =>
      useEscapeKey(() => {
        called += 1
      }, true)
    )
    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape' }))
    expect(called).toBe(1)
  })

  test('does not call onEscape when disabled', () => {
    let called = 0
    renderHook(() =>
      useEscapeKey(() => {
        called += 1
      }, false)
    )
    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape' }))
    expect(called).toBe(0)
  })
})
