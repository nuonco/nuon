import type { ReactNode } from 'react'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Panel } from '@/components/surfaces/Panel'
import { useSurfaces } from '@/hooks/use-surfaces'
import { cn } from '@/utils/classnames'

export interface IHistoryRail {
  children: ReactNode
  className?: string
  history: ReactNode
  title?: string
}

export const HistoryRail = ({
  children,
  className,
  history,
  title = 'History',
}: IHistoryRail) => (
  <div className={cn('grid grid-cols-1 @5xl:grid-cols-12 gap-6', className)}>
    <div className="@5xl:col-span-8 flex flex-col gap-6 min-w-0">
      {children}
    </div>
    <div className="hidden @5xl:flex flex-col @5xl:col-span-4 gap-4 min-w-0">
      <Text variant="base" weight="strong">
        {title}
      </Text>
      {history}
    </div>
  </div>
)

export interface IHistoryPanelButton
  extends Omit<IButtonAsButton, 'children' | 'onClick'> {
  history: ReactNode
  title?: string
}

export const HistoryPanelButton = ({
  history,
  title = 'History',
  variant = 'secondary',
  ...props
}: IHistoryPanelButton) => {
  const { addPanel } = useSurfaces()

  return (
    <div className="@5xl:hidden">
      <Button
        variant={variant}
        onClick={() => addPanel(<Panel heading={title}>{history}</Panel>)}
        {...props}
      >
        <Icon variant="ClockCounterClockwiseIcon" size={16} />
        {title}
      </Button>
    </div>
  )
}
