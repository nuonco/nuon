import { Button } from '@/components/common/Button'
import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import { Text } from '@/components/common/Text'
import { BuildComponentButton } from '@/components/components/management/BuildComponent'
import { DeployTimeline } from '@/components/deploys/DeployTimeline'
import { DeployComponentButton } from '@/components/install-components/management/DeployComponent'
import { TeardownComponentButton } from '@/components/install-components/management/TeardownComponent'
import { ComponentConfigPanel } from '@/components/installs/InstallResourceConfig'
import { Panel } from '@/components/surfaces/Panel'
import { useSurfaces } from '@/hooks/use-surfaces'
import type { TComponent } from '@/types'

export interface IResourceComponentActions {
  appliedConfigId?: string
  component: TComponent
  currentBuildId?: string
  currentDeployStatus?: string
  variant?: 'component' | 'image'
}

export const ResourceComponentActions = ({
  appliedConfigId,
  component,
  currentBuildId,
  currentDeployStatus,
  variant = 'component',
}: IResourceComponentActions) => {
  const { addPanel } = useSurfaces()
  const isImage = variant === 'image'

  return (
    <Dropdown
      alignment="right"
      buttonText=""
      buttonClassName="!p-1"
      icon={<Icon variant="DotsThreeVerticalIcon" />}
      id={`resource-component-actions-${component.id}`}
      variant="ghost"
      aria-label={`More actions for ${component.name}`}
      tooltipProps={{ tipContent: 'More actions' }}
    >
      <Menu>
        <Button
          onClick={() =>
            addPanel(
              <ComponentConfigPanel
                appliedConfigId={appliedConfigId}
                componentId={component.id}
                emptyMessage={
                  isImage
                    ? 'This image is not in the app config this install is using.'
                    : undefined
                }
                name={component.name}
              />,
              `component-config-${component.id}`
            )
          }
        >
          View config
          <Icon variant="SlidersHorizontalIcon" />
        </Button>
        <Button
          onClick={() =>
            addPanel(
              <Panel
                heading={`${component.name} ${
                  isImage ? 'sync' : 'deploy'
                } history`}
                size="half"
              >
                <DeployTimeline
                  componentName={component.name}
                  componentId={component.id}
                  openWorkflowPanel
                  shouldPoll
                  variant={isImage ? 'sync' : 'deploy'}
                />
              </Panel>
            )
          }
        >
          View history
          <Icon variant="ClockCounterClockwiseIcon" />
        </Button>
        <Text>Controls</Text>
        <DeployComponentButton
          component={component}
          currentBuildId={currentBuildId}
          currentDeployStatus={currentDeployStatus}
          isMenuButton
        >
          {isImage ? 'Sync' : 'Deploy'}
        </DeployComponentButton>
        <BuildComponentButton
          component={component}
          isMenuButton
          redirectOnSuccess={false}
        >
          Rebuild
        </BuildComponentButton>
        <hr />
        <TeardownComponentButton
          component={component}
          isMenuButton
          variant="danger"
        />
      </Menu>
    </Dropdown>
  )
}
