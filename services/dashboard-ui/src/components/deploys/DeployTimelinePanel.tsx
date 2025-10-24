'use client'

import { Icon } from '@/components/common/Icon'
import { Panel } from '@/components/surfaces/Panel'

import { DeployTimeline } from './DeployTimeline'

interface IDeployTimelinePanel {
  componentName: string
  componentId: string
}

export const DeployTimelinePanel = ({
  componentName,
  componentId,
}: IDeployTimelinePanel) => {
  return (
    <Panel
      heading={`${componentName} deploy history`}
      triggerButton={{
        children: (
          <>
            <Icon variant="ClockCounterClockwise" /> Latest deploys
          </>
        ),
      }}
      size="half"
    >
      <DeployTimeline
        componentName={componentName}
        componentId={componentId}
        pagination={{}}
        shouldPoll
      />
    </Panel>
  )
}
