import { useState } from 'react'
import { CheckboxInput } from '@/components/common/form/CheckboxInput'
import { componentTypeName } from '@/components/components/ComponentType'
import type { TComponentType } from '@/types'

export const SANDBOX_FILTER = 'sandbox'

export type TBuildTypeFilterKey = TComponentType | typeof SANDBOX_FILTER

export const buildTypeFilterLabel = (type: TBuildTypeFilterKey) =>
  type === SANDBOX_FILTER ? 'Sandbox' : componentTypeName(type)

export const uniqueBuildFilterTypes = (
  types: Array<TBuildTypeFilterKey | undefined | null>
): TBuildTypeFilterKey[] => {
  const seen = new Set<string>()
  const out: TBuildTypeFilterKey[] = []
  for (const type of types) {
    if (!type || type === 'unknown') continue
    if (seen.has(type)) continue
    seen.add(type)
    out.push(type)
  }
  out.sort((a, b) => {
    if (a === SANDBOX_FILTER) return 1
    if (b === SANDBOX_FILTER) return -1
    return buildTypeFilterLabel(a).localeCompare(buildTypeFilterLabel(b))
  })
  return out
}

export const useBuildTypeFilter = (types: TBuildTypeFilterKey[]) => {
  const [deselected, setDeselected] = useState<Set<string>>(() => new Set())

  const toggle = (type: string) => {
    setDeselected((prev) => {
      const next = new Set(prev)
      if (next.has(type)) next.delete(type)
      else next.add(type)
      return next
    })
  }

  const matches = (type: TBuildTypeFilterKey | undefined) => {
    if (!type || type === 'unknown') return true
    return !deselected.has(type)
  }

  return { types, deselected, toggle, matches }
}

interface IBuildTypeFilter {
  types: TBuildTypeFilterKey[]
  deselected: Set<string>
  onToggle: (type: string) => void
}

export const BuildTypeFilter = ({
  types,
  deselected,
  onToggle,
}: IBuildTypeFilter) => {
  if (types.length <= 1) return null

  return (
    <div className="flex flex-wrap items-center gap-x-3 gap-y-1">
      {types.map((type) => (
        <CheckboxInput
          key={type}
          checked={!deselected.has(type)}
          onChange={() => onToggle(type)}
          labelProps={{
            labelText: buildTypeFilterLabel(type),
            className: '!p-1 !gap-1.5',
            labelTextProps: { variant: 'subtext' },
          }}
        />
      ))}
    </div>
  )
}
