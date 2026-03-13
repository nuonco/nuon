import { useQuery } from '@tanstack/react-query'
import { EmptyState } from '@/components/common/EmptyState'
import { Skeleton } from '@/components/common/Skeleton'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getInstallComponents } from '@/lib'
import type { TInstallComponent } from '@/types'
import { TopologyNode } from './TopologyNode'
import { TopologyConnector } from './TopologyConnector'

type TComponentLayer = TInstallComponent[]

function resolveComponentLayers(
  components: TInstallComponent[]
): TComponentLayer[] {
  if (components.length === 0) return []

  const idToComponent = new Map<string, TInstallComponent>()
  for (const ic of components) {
    if (ic.component?.id) {
      idToComponent.set(ic.component.id, ic)
    }
  }

  const placed = new Set<string>()
  const layers: TComponentLayer[] = []

  // Iteratively place components whose deps are already placed
  let remaining = [...components]
  let maxIterations = components.length + 1

  while (remaining.length > 0 && maxIterations > 0) {
    maxIterations--
    const layer: TInstallComponent[] = []

    const next: TInstallComponent[] = []
    for (const ic of remaining) {
      const deps = ic.component?.dependencies ?? []
      const allDepsPlaced = deps.every(
        (depId) => placed.has(depId) || !idToComponent.has(depId)
      )
      if (allDepsPlaced) {
        layer.push(ic)
      } else {
        next.push(ic)
      }
    }

    if (layer.length === 0) {
      // Circular or unresolvable -- dump everything into a final layer
      layers.push(remaining)
      break
    }

    for (const ic of layer) {
      if (ic.component?.id) placed.add(ic.component.id)
    }
    layers.push(layer)
    remaining = next
  }

  return layers
}

function TopologySkeleton() {
  return (
    <div className="flex flex-col items-center gap-6 py-8">
      <Skeleton width="220px" height="60px" />
      <Skeleton width="2px" height="24px" />
      <Skeleton width="220px" height="60px" />
      <Skeleton width="2px" height="24px" />
      <div className="flex gap-4">
        <Skeleton width="160px" height="72px" />
        <Skeleton width="160px" height="72px" />
        <Skeleton width="160px" height="72px" />
      </div>
    </div>
  )
}

export function InstallTopology() {
  const { org } = useOrg()
  const { install } = useInstall()

  const { data: componentsResult, isLoading } = useQuery({
    queryKey: ['install-components-overview', org?.id, install?.id],
    queryFn: () =>
      getInstallComponents({
        installId: install.id,
        orgId: org.id,
        limit: 50,
        offset: 0,
      }),
    enabled: !!org?.id && !!install?.id,
    refetchInterval: 20_000,
  })

  if (isLoading) return <TopologySkeleton />

  const hasStack =
    !!install?.install_stack ||
    (!!install?.runner_status && install.runner_status !== '')
  const hasSandbox =
    install?.sandbox_status != null && install.sandbox_status !== ''
  const componentList: TInstallComponent[] =
    componentsResult?.data ?? (install?.install_components as TInstallComponent[] | undefined) ?? []
  const layers = resolveComponentLayers(componentList)
  const hasComponents = layers.length > 0

  const visibleTiers: string[] = []
  if (hasStack) visibleTiers.push('stack')
  if (hasSandbox) visibleTiers.push('sandbox')
  if (hasComponents) visibleTiers.push('components')

  if (visibleTiers.length === 0) {
    return (
      <EmptyState
        emptyTitle="No infrastructure"
        emptyMessage="This install has no stack, sandbox, or components configured yet."
        variant="diagram"
      />
    )
  }

  return (
    <div className="w-full flex flex-col items-center py-8 gap-0 bg-cool-grey-50 dark:bg-dark-grey-950">
      {hasStack && (
        <>
          <TopologyNode
            variant="stack"
            name={install?.app?.name ?? 'Stack'}
            status={
              install?.install_stack?.versions?.[0]?.composite_status?.status ??
              install?.runner_status ??
              'pending'
            }
            href="stacks"
          />
          {(hasSandbox || hasComponents) && <TopologyConnector />}
        </>
      )}

      {hasSandbox && (
        <>
          <TopologyNode
            variant="sandbox"
            name="Sandbox"
            status={install?.sandbox_status ?? 'unknown'}
            href="sandbox"
          />
          {hasComponents && <TopologyConnector />}
        </>
      )}

      {hasComponents && (
        <div className="flex flex-col items-center gap-0 w-full">
          {layers.map((layer, layerIndex) => (
            <div key={layerIndex} className="flex flex-col items-center w-full">
              {layerIndex > 0 && <TopologyConnector variant="branch" count={layer.length} />}
              {layerIndex === 0 && visibleTiers.length > 1 && (
                <TopologyConnector variant="branch" count={layer.length} />
              )}
              {layerIndex === 0 && visibleTiers.length === 1 && layer.length > 1 && (
                <div className="h-2" />
              )}
              <div className="flex items-start justify-center gap-4 flex-wrap">
                {layer.map((ic) => {
                  const rawStatus = ic.status ?? 'unknown'
                  const displayStatus = rawStatus === 'unknown' ? 'not-attempted' : rawStatus
                  return (
                    <TopologyNode
                      key={ic.id}
                      variant="component"
                      name={ic.component?.name ?? 'Unknown'}
                      componentType={ic.component?.type}
                      status={displayStatus}
                      hasDrift={!!ic.drifted_object}
                      href={`components/${ic.id}`}
                    />
                  )
                })}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
