import { Button } from '@/components/common/Button'
import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import { DeployTimeline } from '@/components/deploys/DeployTimeline'
import { TeardownComponentButton } from '@/components/install-components/management/TeardownComponent'
import { ToggleComponentButton } from '@/components/install-components/management/ToggleComponent'
import { ComponentConfigPanel } from '@/components/installs/InstallResourceConfig'
import { Panel } from '@/components/surfaces/Panel'
import { useInstall } from '@/hooks/use-install'
import { useSurfaces } from '@/hooks/use-surfaces'
import type { TComponent } from '@/types'

export interface IResourceComponentActions {
  appliedConfigId?: string
  component: TComponent
  currentDeployStatus?: string
  enabled?: boolean | null
  onViewDetails?: () => void
  variant?: 'component' | 'image'
}

export const ResourceComponentActions = ({
  appliedConfigId,
  component,
  currentDeployStatus,
  enabled,
  onViewDetails,
  variant = 'component',
}: IResourceComponentActions) => {
  const { install } = useInstall()
  const { addPanel } = useSurfaces()
  const isImage = variant === 'image'
  const disabled = enabled === false
  const managedByConfig =
    install?.metadata?.managed_by === 'nuon/cli/install-config'

  return (
    <div className="flex flex-wrap items-center gap-2">
      {disabled ? (
        <ToggleComponentButton
          component={component}
          enabling
          variant="secondary"
          disabled={managedByConfig}
          tooltipProps={
            managedByConfig
              ? {
                  tipContent:
                    'Managed by install config. View details for the input to change.',
                }
              : undefined
          }
        />
      ) : null}
      {onViewDetails ? (
        <Button variant="secondary" onClick={onViewDetails}>
          View details
        </Button>
      ) : null}
      {!disabled ? (
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
            {isImage ? (
              <Button
                onClick={() =>
                  addPanel(
                    <Panel
                      heading={`${component.name} sync history`}
                      size="half"
                    >
                      <DeployTimeline
                        componentName={component.name}
                        componentId={component.id}
                        openWorkflowPanel
                        shouldPoll
                        variant="sync"
                      />
                    </Panel>
                  )
                }
              >
                View history
                <Icon variant="ClockCounterClockwiseIcon" />
              </Button>
            ) : null}
            {!isImage && currentDeployStatus !== 'inactive' ? (
              <>
                <hr />
                <TeardownComponentButton
                  component={component}
                  isMenuButton
                  variant="danger"
                />
              </>
            ) : null}
          </Menu>
        </Dropdown>
      ) : null}
    </div>
  )
}
