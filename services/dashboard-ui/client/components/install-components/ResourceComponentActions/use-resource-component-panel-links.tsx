import { DeployTimeline } from '@/components/deploys/DeployTimeline'
import { ComponentConfigPanel } from '@/components/installs/InstallResourceConfig'
import { Panel, type IPanel } from '@/components/surfaces/Panel'
import { usePanelDeepLinks } from '@/hooks/use-panel-deep-links'
import type { TInstallComponent } from '@/types'

export const IMAGE_CONFIG_EMPTY_MESSAGE =
  'This image is not in the app config this install is using.'

export const componentConfigKey = (componentId: string) =>
  `component-config-${componentId}`

export const componentHistoryKey = (componentId: string) =>
  `component-history-${componentId}`

const idFromKey = (key: string, prefix: string) =>
  key.startsWith(prefix) ? key.slice(prefix.length) : ''

export const ComponentHistoryPanel = ({
  componentId,
  name,
  ...props
}: {
  componentId: string
  name?: string
} & Omit<IPanel, 'heading' | 'children'>) => (
  <Panel
    heading={name ? `${name} sync history` : 'Sync history'}
    panelKey={componentHistoryKey(componentId)}
    size="half"
    {...props}
  >
    <DeployTimeline
      componentId={componentId}
      componentName={name ?? ''}
      openWorkflowPanel
      shouldPoll
      variant="sync"
    />
  </Panel>
)

export const useResourceComponentPanelLinks = ({
  isLoading,
  rows,
  variant = 'component',
}: {
  isLoading: boolean
  rows?: TInstallComponent[]
  variant?: 'component' | 'image'
}) => {
  usePanelDeepLinks((key) => {
    const configId = idFromKey(key, 'component-config-')
    const historyId = idFromKey(key, 'component-history-')
    const componentId = configId || historyId
    if (!componentId) return null
    if (isLoading && !rows) return null
    const match = rows?.find(
      (item) => (item.component?.id ?? item.component_id) === componentId
    )
    const name = match?.component?.name
    if (configId) {
      return (
        <ComponentConfigPanel
          appliedConfigId={match?.app_config_ref?.applied_config_id}
          componentId={configId}
          emptyMessage={
            variant === 'image' ? IMAGE_CONFIG_EMPTY_MESSAGE : undefined
          }
          name={name}
        />
      )
    }
    return <ComponentHistoryPanel componentId={historyId} name={name} />
  })
}
