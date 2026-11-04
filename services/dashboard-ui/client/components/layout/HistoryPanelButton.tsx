import type { ReactNode } from 'react'
import { Icon } from '@/components/common/Icon'
import { Panel } from '@/components/surfaces/Panel'
import type { IButtonAsButton } from '@/components/common/Button'

export interface IHistoryPanelButton
  extends Omit<IButtonAsButton, 'children' | 'onClick'> {
  history: ReactNode
  panelKey?: string
  title?: string
}

const historyKey = (title: string) =>
  `history-${title.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-|-$/g, '')}`

export const HistoryPanelButton = ({
  history,
  panelKey,
  title = 'History',
  variant = 'secondary',
  ...props
}: IHistoryPanelButton) => (
  <Panel
    heading={title}
    panelKey={panelKey ?? historyKey(title)}
    triggerButton={{
      variant,
      children: (
        <>
          <Icon variant="ClockCounterClockwiseIcon" size={16} />
          {title}
        </>
      ),
      ...props,
    }}
  >
    {history}
  </Panel>
)
