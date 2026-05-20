import { useMemo } from 'react'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useSurfaces } from '@/hooks/use-surfaces'
import { ComponentDependencyGraph, type GraphNode, type GraphEdge } from './ComponentDependencyGraph'
import type { TAppConfig, TComponentType } from '@/types'

type Connection = NonNullable<TAppConfig['component_config_connections']>[number]

function buildTransitiveGraph(
  connections: Connection[],
  componentId: string,
  componentName: string,
  componentType?: TComponentType,
) {
  const connById = new Map(connections.map((c) => [c.component_id!, c]))

  const depIds = new Set<string>()
  const deptIds = new Set<string>()

  // Walk upstream: collect all transitive dependencies
  const walkDeps = (id: string) => {
    const conn = connById.get(id)
    for (const depId of conn?.component_dependency_ids ?? []) {
      if (!depIds.has(depId) && depId !== componentId) {
        depIds.add(depId)
        walkDeps(depId)
      }
    }
  }
  walkDeps(componentId)

  // Walk downstream: collect all transitive dependents
  const dependentsOf = new Map<string, string[]>()
  for (const conn of connections) {
    for (const depId of conn.component_dependency_ids ?? []) {
      const list = dependentsOf.get(depId) ?? []
      list.push(conn.component_id!)
      dependentsOf.set(depId, list)
    }
  }

  const walkDepts = (id: string) => {
    for (const deptId of dependentsOf.get(id) ?? []) {
      if (!deptIds.has(deptId) && deptId !== componentId) {
        deptIds.add(deptId)
        walkDepts(deptId)
      }
    }
  }
  walkDepts(componentId)

  const allIds = new Set([componentId, ...depIds, ...deptIds])

  const nodes: GraphNode[] = []
  for (const id of allIds) {
    const conn = connById.get(id)
    const role = id === componentId ? 'current' as const
      : depIds.has(id) ? 'dependency' as const
      : 'dependent' as const
    nodes.push({
      id,
      name: id === componentId ? componentName : (conn?.component_name ?? id),
      type: id === componentId ? componentType : conn?.type,
      role,
    })
  }

  const edges: GraphEdge[] = []
  for (const conn of connections) {
    if (!allIds.has(conn.component_id!)) continue
    for (const depId of conn.component_dependency_ids ?? []) {
      if (allIds.has(depId)) {
        edges.push({ sourceId: depId, targetId: conn.component_id! })
      }
    }
  }

  return { nodes, edges }
}

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
  const connections = appConfig?.component_config_connections ?? []

  const { nodes, edges } = useMemo(
    () => buildTransitiveGraph(connections, componentId, componentName, componentType),
    [connections, componentId, componentName, componentType],
  )

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
          nodes={nodes}
          edges={edges}
          currentId={componentId}
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
