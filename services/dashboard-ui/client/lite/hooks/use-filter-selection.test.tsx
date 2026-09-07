import { describe, expect, test } from 'bun:test'
import { act, renderHook } from '@testing-library/react'
import { useFilterSelection } from './use-filter-selection'

const OPTIONS = ['create', 'update', 'delete'] as const

describe('useFilterSelection', () => {
  test('starts with every option selected', () => {
    const { result } = renderHook(() => useFilterSelection(OPTIONS))

    expect([...result.current.selected]).toEqual([...OPTIONS])
    expect(result.current.isConstrained).toBe(false)
  })

  test('toggles one option without changing the others', () => {
    const { result } = renderHook(() => useFilterSelection(OPTIONS))

    act(() => result.current.toggle('update'))

    expect([...result.current.selected]).toEqual(['create', 'delete'])
    expect(result.current.isConstrained).toBe(true)
  })

  test('isolates an option and resets when isolated again', () => {
    const { result } = renderHook(() => useFilterSelection(OPTIONS))

    act(() => result.current.isolate('update'))
    expect([...result.current.selected]).toEqual(['update'])

    act(() => result.current.isolate('update'))
    expect([...result.current.selected]).toEqual([...OPTIONS])
  })

  test('reset restores every option', () => {
    const { result } = renderHook(() => useFilterSelection(OPTIONS))

    act(() => result.current.toggle('create'))
    act(() => result.current.reset())

    expect([...result.current.selected]).toEqual([...OPTIONS])
  })

  test('starts from the declared defaults', () => {
    const { result } = renderHook(() =>
      useFilterSelection(OPTIONS, ['update', 'delete'])
    )

    expect([...result.current.selected]).toEqual(['update', 'delete'])
    expect(result.current.isConstrained).toBe(false)
  })

  test('reset and re-isolate return to the defaults', () => {
    const { result } = renderHook(() =>
      useFilterSelection(OPTIONS, ['update', 'delete'])
    )

    act(() => result.current.toggle('create'))
    expect(result.current.isConstrained).toBe(true)

    act(() => result.current.reset())
    expect([...result.current.selected]).toEqual(['update', 'delete'])

    act(() => result.current.isolate('create'))
    act(() => result.current.isolate('create'))
    expect([...result.current.selected]).toEqual(['update', 'delete'])
  })

  test('keeps the selection when defaults are passed inline', () => {
    const { result, rerender } = renderHook(() =>
      useFilterSelection(['create', 'update', 'delete'], ['update', 'delete'])
    )

    act(() => result.current.toggle('create'))
    rerender()

    expect([...result.current.selected]).toEqual(['update', 'delete', 'create'])
  })

  test('ignores defaults that are not available options', () => {
    const { result } = renderHook(() =>
      useFilterSelection(OPTIONS, ['update', 'ghost'])
    )

    expect([...result.current.selected]).toEqual(['update'])
  })
})
