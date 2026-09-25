import type { ReactNode } from 'react'
import { Text } from '@/components/common/Text'
import { Panel, type IPanel } from '@/components/surfaces/Panel'
import { MAX_ACTIVITY_PINS } from './storage'
import { ActivityPinListContainer } from './ActivityPinListContainer'
import { useActivityPins } from './use-activity-pins'

export const ActivityPinsPanelCopy = ({
  pinnedCount,
}: {
  pinnedCount: number
}) => (
  <div className="flex flex-col gap-2">
    <Text theme="neutral">
      Pin up to {MAX_ACTIVITY_PINS} runbooks and actions you run often. Pins
      stay at the top of Activity on this install, saved in this browser for
      you.
    </Text>
    {pinnedCount >= MAX_ACTIVITY_PINS ? (
      <Text theme="warn">
        You have {MAX_ACTIVITY_PINS} pinned. Remove one before pinning another.
      </Text>
    ) : (
      <Text variant="subtext" theme="neutral">
        {pinnedCount} of {MAX_ACTIVITY_PINS} pinned.
      </Text>
    )}
  </div>
)

export interface IActivityPinsPanelView extends Omit<IPanel, 'heading'> {
  actions: ReactNode
  pinnedCount: number
  runbooks: ReactNode
}

export const ActivityPinsPanelView = ({
  actions,
  pinnedCount,
  runbooks,
  ...props
}: IActivityPinsPanelView) => (
  <Panel {...props} size="half" heading="Pin shortcuts">
    <ActivityPinsPanelCopy pinnedCount={pinnedCount} />
    {runbooks}
    <div className="border-t pt-4">{actions}</div>
  </Panel>
)

const ActivityPinsPanelBody = () => {
  const { pins } = useActivityPins()

  return (
    <>
      <ActivityPinsPanelCopy pinnedCount={pins.length} />
      <ActivityPinListContainer kind="runbook" />
      <div className="border-t pt-4">
        <ActivityPinListContainer kind="action" />
      </div>
    </>
  )
}

export const ActivityPinsPanel = (props: Omit<IPanel, 'children' | 'heading'>) => (
  <Panel {...props} size="half" heading="Pin shortcuts">
    <ActivityPinsPanelBody />
  </Panel>
)
