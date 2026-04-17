import React, { useMemo } from 'react'
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
      className="hidden md:flex items-center gap-3 h-9 pl-3 pr-2 py-2 w-[240px] border border-[rgba(158,168,179,0.24)] dark:border-[rgba(158,168,179,0.24)] rounded-lg bg-white dark:bg-dark-grey-900 text-[#9ea8b3] dark:text-[#9ea8b3] text-sm font-normal leading-[21px] tracking-[-0.2px] cursor-pointer transition-colors overflow-hidden"
    >
      <Icon variant="MagnifyingGlassIcon" size={16} className="shrink-0" />
      <span className="flex-1 text-left min-w-0">Search...</span>
      <span className="inline-flex gap-1 items-center shrink-0">
        <span className={`flex items-center justify-center h-5 ${mac ? 'w-5' : 'px-1.5'} border border-[rgba(158,168,179,0.24)] dark:border-[rgba(158,168,179,0.24)] rounded bg-white dark:bg-dark-grey-900 text-[11px] font-medium shrink-0`}>
          {mac ? '⌘' : 'CTRL'}
        </span>
        <span className="flex items-center justify-center size-5 border border-[rgba(158,168,179,0.24)] dark:border-[rgba(158,168,179,0.24)] rounded bg-white dark:bg-dark-grey-900 text-[11px] font-medium shrink-0">
          K
        </span>
      </span>
    </button>
  )
}
