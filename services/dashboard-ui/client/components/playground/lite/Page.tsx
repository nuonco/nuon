import type { ReactNode } from 'react'
import { cn } from '@/utils/classnames'
import { Block } from './Block'
import { PageHeader } from './PageHeader'
import { TabNav } from './TabNav'
import type { ICrumb, INavItem } from './types'
import { useScrolled } from './use-scrolled'
import { labelWidth } from './utils'

export interface IPage {
  crumbs?: ICrumb[]
  actions?: string[]
  actionsSlot?: ReactNode
  tabs?: INavItem[]
  children: ReactNode
}

export const Page = ({
  crumbs,
  actions = [],
  actionsSlot,
  tabs = [],
  children,
}: IPage) => {
  const hasActions = actions.length > 0 || Boolean(actionsSlot)
  const { ref, isScrolled } = useScrolled()

  return (
    <>
      <div
        ref={ref}
        className={cn(
          'sticky top-0 z-20 flex flex-none flex-col gap-4 rounded-lg px-4 py-3 transition-all',
          isScrolled
            ? 'bg-cool-grey-100/70 shadow-lg backdrop-blur-md dark:bg-dark-grey-800/70'
            : 'bg-transparent shadow-none'
        )}
      >
        <PageHeader crumbs={crumbs} />

        {(tabs.length > 0 || hasActions) && (
          <div className="flex items-center justify-between gap-4">
            {tabs.length > 0 ? <TabNav tabs={tabs} /> : <span />}

            {hasActions && (
              <div className="flex flex-none items-center gap-4">
                {actions.map((action) => (
                  <Block
                    key={action}
                    className="h-[32px]"
                    style={{ width: labelWidth(action) }}
                    title={action}
                    text={action}
                  />
                ))}
                {actionsSlot}
              </div>
            )}
          </div>
        )}
      </div>
      {children}
    </>
  )
}
