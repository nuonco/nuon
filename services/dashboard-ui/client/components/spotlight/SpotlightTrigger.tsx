import React, { useMemo } from 'react'
import { Badge } from '@/components/common/Badge'
import { Icon } from '@/components/common/Icon'
import { useSurfaces } from '@/hooks/use-surfaces'
import { SpotlightModalContainer } from '@/components/spotlight/Spotlight'

const isMac = () =>
  typeof navigator !== 'undefined' &&
  /Mac|iPhone|iPad|iPod/.test(navigator.userAgent)

export const SpotlightTrigger = () => {
  const { addModal } = useSurfaces()
  const mac = useMemo(isMac, [])

  return (
    <button
      type="button"
      onClick={() => addModal(React.createElement(SpotlightModalContainer))}
      className="hidden md:flex items-center gap-2 h-8 px-3 max-w-[280px] w-full border border-border rounded-lg bg-muted/50 text-muted-foreground text-sm cursor-pointer hover:bg-muted transition-colors"
    >
      <Icon variant="MagnifyingGlassIcon" size={14} className="shrink-0" />
      <span className="flex-1 text-left">Search...</span>
      <span className="inline-flex gap-0.5 shrink-0">
        <Badge variant="code" size="sm">
          {mac ? '⌘' : 'CTRL'}
        </Badge>
        <Badge variant="code" size="sm">
          K
        </Badge>
      </span>
    </button>
  )
}
