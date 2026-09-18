import type { ReactNode } from 'react'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Panel } from '@/components/surfaces/Panel'
import { useSurfaces } from '@/hooks/use-surfaces'

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
    <Button
      variant={variant}
      onClick={() => addPanel(<Panel heading={title}>{history}</Panel>)}
      {...props}
    >
      <Icon variant="ClockCounterClockwiseIcon" size={16} />
      {title}
    </Button>
  )
}
