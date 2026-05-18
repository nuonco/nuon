import { useMemo } from 'react'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useSurfaces } from '@/hooks/use-surfaces'
import { ComponentDependencyGraph } from './ComponentDependencyGraph'
import type { TAppConfig, TComponentType } from '@/types'

interface IComponentDependencyGraphModal extends IModal {
  componentId: string
  componentName: string
  componentType?: TComponentType
  appConfig: TAppConfig
  basePath: string
}

export const ComponentDependencyGraphModal = ({
  componentId,
  componentName,
  componentType,
  appConfig,
  basePath,
  ...props
}: IComponentDependencyGraphModal) => {
  const { removeModal } = useSurfaces()
  const connections = appConfig?.component_config_connections

  const { current, dependencies, dependents } = useMemo(() => {
    const config = connections?.find((c) => c.component_id === componentId)
    const depIds = config?.component_dependency_ids ?? []

    const deps = depIds
      .map((id) => {
        const conn = connections?.find((c) => c.component_id === id)
        return conn
          ? { id: conn.component_id!, name: conn.component_name ?? conn.component_id!, type: conn.type }
          : null
      })
      .filter(Boolean) as { id: string; name: string; type?: TComponentType }[]

    const depts = (connections ?? [])
      .filter((c) => c.component_dependency_ids?.includes(componentId))
      .map((c) => ({
        id: c.component_id!,
        name: c.component_name ?? c.component_id!,
        type: c.type,
      }))
      .filter((c) => c.id) as { id: string; name: string; type?: TComponentType }[]

    return {
      current: { id: componentId, name: componentName, type: componentType },
      dependencies: deps,
      dependents: depts,
    }
  }, [componentId, componentName, componentType, connections])

  return (
    <Modal
      heading={
        <Text flex className="gap-2" variant="h3" weight="strong">
          <Icon variant="GraphIcon" size="24" />
          Dependency graph
        </Text>
      }
      size="xl"
      {...props}
    >
      <div style={{ width: '100%', height: '32rem' }}>
        <ComponentDependencyGraph
          current={current}
          dependencies={dependencies}
          dependents={dependents}
          basePath={basePath}
          onNavigate={() => removeModal(props.modalId)}
        />
      </div>
    </Modal>
  )
}

interface IComponentDependencyGraphButton extends Omit<IButtonAsButton, 'onClick'> {
  componentId: string
  componentName: string
  componentType?: TComponentType
  appConfig: TAppConfig
  basePath: string
}

export const ComponentDependencyGraphButton = ({
  componentId,
  componentName,
  componentType,
  appConfig,
  basePath,
  ...props
}: IComponentDependencyGraphButton) => {
  const { addModal } = useSurfaces()

  const connections = appConfig?.component_config_connections
  const config = connections?.find((c) => c.component_id === componentId)
  const hasDeps = (config?.component_dependency_ids?.length ?? 0) > 0
  const hasDependents = connections?.some((c) =>
    c.component_dependency_ids?.includes(componentId),
  )

  if (!hasDeps && !hasDependents) return null

  const modal = (
    <ComponentDependencyGraphModal
      componentId={componentId}
      componentName={componentName}
      componentType={componentType}
      appConfig={appConfig}
      basePath={basePath}
    />
  )

  return (
    <Button variant="secondary" onClick={() => addModal(modal)} {...props}>
      <Icon variant="GraphIcon" size={16} />
      Dependency graph
    </Button>
  )
}
