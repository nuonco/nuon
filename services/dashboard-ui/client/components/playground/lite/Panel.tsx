import type { ReactNode } from 'react'
import { cn } from '@/utils/classnames'
import { Block } from './Block'
import { labelWidth } from './utils'

export interface IPanel {
  title: string
  action?: string
  children: ReactNode
  className?: string
}

export const Panel = ({ title, action, children, className }: IPanel) => (
  <section
    className={cn(
      'flex flex-col gap-4 rounded-lg bg-cool-grey-100 dark:bg-dark-grey-800 p-4',
      className
    )}
  >
    <header className="flex items-center justify-between gap-4">
      <Block
        className="h-[14px]"
        style={{ width: labelWidth(title) }}
        title={title}
        text={title}
      />
      {action && (
        <Block
          className="h-[12px] opacity-60"
          style={{ width: labelWidth(action) }}
          title={action}
          text={action}
        />
      )}
    </header>
    {children}
  </section>
)
