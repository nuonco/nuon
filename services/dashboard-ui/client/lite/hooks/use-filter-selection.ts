import { useMemo, useState } from 'react'

const setsMatch = <T>(a: Set<T>, b: Set<T>) =>
  a.size === b.size && [...a].every((value) => b.has(value))

export const useFilterSelection = <T extends string>(
  options: readonly T[],
  defaultSelected?: readonly T[]
) => {
  const available = useMemo(() => new Set(options), [options])
  const defaults = useMemo(
    () => new Set((defaultSelected ?? options).filter((v) => available.has(v))),
    [available, defaultSelected, options]
  )

  const key = `${[...available].join('\u0000')}|${[...defaults].join('\u0000')}`
  const [state, setState] = useState(() => ({
    key,
    selected: new Set(defaults),
  }))

  const isCurrent = state.key === key
  if (!isCurrent) setState({ key, selected: new Set(defaults) })
  const selected = isCurrent ? state.selected : defaults

  const setSelected = (next: Set<T> | ((current: Set<T>) => Set<T>)) => {
    setState((current) => ({
      key,
      selected:
        typeof next === 'function'
          ? next(current.key === key ? current.selected : defaults)
          : next,
    }))
  }

  const toggle = (value: T) => {
    setSelected((current) => {
      const next = new Set(current)
      if (next.has(value)) next.delete(value)
      else next.add(value)
      return next
    })
  }

  const isolate = (value: T) => {
    setSelected((current) =>
      current.size === 1 && current.has(value)
        ? new Set(defaults)
        : new Set([value])
    )
  }

  const reset = () => setSelected(new Set(defaults))

  return {
    defaults,
    isConstrained: !setsMatch(selected, defaults),
    isolate,
    reset,
    selected,
    setSelected,
    toggle,
  }
}
