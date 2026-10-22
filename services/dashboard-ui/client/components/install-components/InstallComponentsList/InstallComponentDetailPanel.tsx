import { CodeBlock } from '@/components/common/CodeBlock'
import { Text } from '@/components/common/Text'
import { DeployTimeline } from '@/components/deploys/DeployTimeline'
import { InstallComponentConfig } from '@/components/install-components/InstallComponentConfig'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { Panel, type IPanel } from '@/components/surfaces/Panel'
import { useInstall } from '@/hooks/use-install'

export const InstallComponentDetailPanel = ({
  componentId,
  disabled = false,
  name,
  ...panelProps
}: {
  componentId: string
  disabled?: boolean
  name: string
} & Omit<IPanel, 'heading' | 'children'>) => {
  const { install } = useInstall()
  const managedByConfig =
    install?.metadata?.managed_by === 'nuon/cli/install-config'

  return (
    <Panel
      heading={name}
      size="half"
      aria-label={`${name} details`}
      panelKey={`install-component-${componentId}`}
      {...panelProps}
    >
      <div className="flex flex-col gap-4">
        {disabled && managedByConfig ? (
          <div
            className="flex flex-col gap-3"
            data-testid="component-enable-config"
          >
            <SectionHeader
              title="Enable through install config"
              description="This install is managed by a config file. Change this toggle to true, then sync the install config to plan and deploy the component."
            />
            <CodeBlock language="toml" showCopy>
              {`[component_toggles]\n${JSON.stringify(name)} = true`}
            </CodeBlock>
          </div>
        ) : null}
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
}
