import {
  useState,
  type HTMLAttributes,
  type ReactNode,
  type UIEvent,
} from 'react'
import { cn } from '@/utils/classnames'
import { ShellBackground } from '../../atoms/ShellBackground'
import { FocusHeader } from '../../organisms/FocusHeader'

export interface IFocusShell
  extends Omit<HTMLAttributes<HTMLDivElement>, 'children'> {
  actions?: ReactNode
  children: ReactNode
  contentClassName?: string
  fullBleed?: boolean
  homeHref?: string
}

export const FocusShell = ({
  actions,
  children,
  contentClassName,
  fullBleed = false,
  homeHref = '/',
  className,
  ...props
}: IFocusShell) => {
  const [contentScrolled, setContentScrolled] = useState(false)

  const handleContentScroll = (event: UIEvent<HTMLElement>) => {
    setContentScrolled(event.currentTarget.scrollTop > 0)
  }

  return (
    <div
      className={cn(
        'relative isolate flex h-screen w-full flex-col overflow-hidden bg-surface-default',
        className
      )}
      {...props}
    >
      <ShellBackground />
      <main
        data-focus-scroll
        className="relative z-10 flex min-h-0 flex-1 flex-col gap-8 overflow-y-auto pb-4"
        onScroll={handleContentScroll}
      >
        <FocusHeader
          actions={actions}
          homeHref={homeHref}
          scrolled={contentScrolled}
          className="mx-4 mt-3"
        />
        <div
          className={cn(
            fullBleed
              ? 'min-h-full w-full'
              : 'mx-auto w-full max-w-6xl px-4 pb-8 sm:px-6 sm:pb-10 lg:px-8'
          )}
        >
          <div className={cn('mx-auto min-h-full w-full', contentClassName)}>
            {children}
          </div>
        </div>
      </main>
    </div>
  )
}
