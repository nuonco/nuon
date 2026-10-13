import { Text } from '@/components/common/Text'
import { DeployTimeline } from '@/components/deploys/DeployTimeline'
import { InstallComponentConfig } from '@/components/install-components/InstallComponentConfig'
import { Panel, type IPanel } from '@/components/surfaces/Panel'

export const InstallComponentDetailPanel = ({
  componentId,
  name,
  ...panelProps
}: {
  componentId: string
  name: string
} & Omit<IPanel, 'heading' | 'children'>) => (
  <Panel
    heading={name}
    size="half"
    aria-label={`${name} details`}
    panelKey={`install-component-${componentId}`}
    {...panelProps}
  >
    <div className="flex flex-col gap-4">
      <InstallComponentConfig componentId={componentId} />
      <div className="flex flex-col gap-2 border-t pt-4">
        <Text variant="subtext" weight="strong" theme="neutral">
          Deploy history
        </Text>
        <DeployTimeline
          componentId={componentId}
          componentName={name}
          openWorkflowPanel
          shouldPoll
        />
      </div>
    </div>
  </Panel>
)
