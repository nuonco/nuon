import { useCallback, useEffect, useRef, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { EmptyState } from '@/components/common/EmptyState'
import { Icon } from '@/components/common/Icon'
import { Skeleton } from '@/components/common/Skeleton'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getInstallComponents } from '@/lib'
import type { TInstallComponent } from '@/types'
import { cn } from '@/utils/classnames'
import { TopologyConnector } from './TopologyConnector'
import { TopologyNode } from './TopologyNode'

type TComponentLayer = TInstallComponent[]

const MAX_PER_ROW = 4

function chunkArray<T>(arr: T[], size: number): T[][] {
  const chunks: T[][] = []
  for (let i = 0; i < arr.length; i += size) chunks.push(arr.slice(i, i + size))
  return chunks
}

function resolveComponentLayers(components: TInstallComponent[]): TComponentLayer[] {
  if (components.length === 0) return []

  const idToComponent = new Map<string, TInstallComponent>()
  for (const ic of components) {
    if (ic.component?.id) idToComponent.set(ic.component.id, ic)
  }

  const placed = new Set<string>()
  const layers: TComponentLayer[] = []
  let remaining = [...components]
  let maxIterations = components.length + 1

  while (remaining.length > 0 && maxIterations > 0) {
    maxIterations--
    const layer: TInstallComponent[] = []
    const next: TInstallComponent[] = []
    for (const ic of remaining) {
      const deps = ic.component?.dependencies ?? []
      const allDepsPlaced = deps.every((d) => placed.has(d) || !idToComponent.has(d))
      if (allDepsPlaced) layer.push(ic)
      else next.push(ic)
    }
    if (layer.length === 0) { layers.push(remaining); break }
    for (const ic of layer) { if (ic.component?.id) placed.add(ic.component.id) }
    layers.push(layer)
    remaining = next
  }

  return layers
}

function TopologySkeleton() {
  return (
    <div className="flex flex-col items-center gap-6 py-10">
      <Skeleton width="220px" height="60px" />
      <Skeleton width="2px" height="24px" />
      <Skeleton width="220px" height="60px" />
      <Skeleton width="2px" height="24px" />
      <div className="flex gap-3">
        <Skeleton width="180px" height="60px" />
        <Skeleton width="180px" height="60px" />
        <Skeleton width="180px" height="60px" />
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
      getInstallComponents({ installId: install.id, orgId: org.id, limit: 50, offset: 0 }),
    enabled: !!org?.id && !!install?.id,
    refetchInterval: 5_000,
  })

  // ── all hooks must come before any early returns ──────────────────────────
  const containerRef = useRef<HTMLDivElement>(null)
  const [scale, setScale] = useState(1)
  const [isFullscreen, setIsFullscreen] = useState(false)

  const ZOOM_STEP = 0.15
  const MIN_ZOOM = 0.4
  const MAX_ZOOM = 2

  const zoomIn = useCallback(() => setScale((s) => Math.min(MAX_ZOOM, +(s + ZOOM_STEP).toFixed(2))), [])
  const zoomOut = useCallback(() => setScale((s) => Math.max(MIN_ZOOM, +(s - ZOOM_STEP).toFixed(2))), [])
  const resetZoom = useCallback(() => setScale(1), [])

  const toggleFullscreen = useCallback(async () => {
    if (!containerRef.current) return
    if (!document.fullscreenElement) {
      await containerRef.current.requestFullscreen()
    } else {
      await document.exitFullscreen()
    }
  }, [])

  useEffect(() => {
    const onFsChange = () => setIsFullscreen(!!document.fullscreenElement)
    document.addEventListener('fullscreenchange', onFsChange)
    return () => document.removeEventListener('fullscreenchange', onFsChange)
  }, [])

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement) return
      if (e.key === '=' || e.key === '+') zoomIn()
      if (e.key === '-') zoomOut()
      if (e.key === '0') resetZoom()
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [zoomIn, zoomOut, resetZoom])
  // ─────────────────────────────────────────────────────────────────────────

  const hasStack = !!install?.install_stack || (!!install?.runner_status && install.runner_status !== '')
  const hasSandbox = install?.sandbox_status != null && install.sandbox_status !== ''
  const componentList: TInstallComponent[] =
    componentsResult?.data ?? (install?.install_components as TInstallComponent[] | undefined) ?? []
  const layers = resolveComponentLayers(componentList)
  const hasComponents = layers.length > 0

  const visibleTiers: string[] = []
  if (hasStack) visibleTiers.push('stack')
  if (hasSandbox) visibleTiers.push('sandbox')
  if (hasComponents) visibleTiers.push('components')

  if (!isLoading && visibleTiers.length === 0) {
    return (
      <EmptyState
        emptyTitle="No infrastructure"
        emptyMessage="This install has no stack, sandbox, or components configured yet."
        variant="diagram"
      />
    )
  }

  if (isLoading) return <TopologySkeleton />

  const normalizeStatus = (s: string | null | undefined) => {
    if (!s || s === 'unknown') return 'not-attempted'
    return s
  }

  const stackStatus =
    install?.install_stack?.versions?.at(-1)?.composite_status?.status ??
    install?.runner_status ?? 'not-attempted'

  const ctrlBtnClass = cn(
    'flex items-center justify-center w-7 h-7 rounded-md',
    'bg-white dark:bg-dark-grey-800',
    'border border-cool-grey-200 dark:border-dark-grey-600',
    'text-cool-grey-600 dark:text-cool-grey-400',
    'hover:bg-cool-grey-50 dark:hover:bg-dark-grey-700',
    'hover:text-cool-grey-900 dark:hover:text-cool-grey-100',
    'transition-colors duration-100 cursor-pointer select-none',
    'disabled:opacity-40 disabled:cursor-not-allowed',
  )

  return (
    <div
      ref={containerRef}
      className={cn(
        'relative w-full overflow-hidden',
        isFullscreen && 'bg-cool-grey-100 dark:bg-dark-grey-900 flex flex-col',
      )}
    >
      {/* Zoom + fullscreen controls */}
      <div className="absolute top-3 right-3 z-10 flex items-center gap-1">
        <button
          type="button"
          className={ctrlBtnClass}
          onClick={zoomOut}
          disabled={scale <= MIN_ZOOM}
          title="Zoom out ( - )"
        >
          <Icon variant="Minus" size={14} weight="bold" />
        </button>

        <button
          type="button"
          className={cn(ctrlBtnClass, 'w-auto px-2 tabular-nums text-[11px] font-medium min-w-[44px]')}
          onClick={resetZoom}
          title="Reset zoom ( 0 )"
        >
          {Math.round(scale * 100)}%
        </button>

        <button
          type="button"
          className={ctrlBtnClass}
          onClick={zoomIn}
          disabled={scale >= MAX_ZOOM}
          title="Zoom in ( + )"
        >
          <Icon variant="Plus" size={14} weight="bold" />
        </button>

        <div className="w-px h-4 bg-cool-grey-200 dark:bg-dark-grey-600 mx-0.5" />

        <button
          type="button"
          className={ctrlBtnClass}
          onClick={toggleFullscreen}
          title={isFullscreen ? 'Exit fullscreen' : 'Fullscreen'}
        >
          <Icon variant={isFullscreen ? 'CornersIn' : 'CornersOut'} size={14} weight="bold" />
        </button>
      </div>

    <div
      className="w-full flex flex-col items-center py-10 pb-8 gap-0 bg-cool-grey-100 dark:bg-dark-grey-900 topology-canvas"
      style={{
        transform: `scale(${scale})`,
        transformOrigin: 'top center',
        marginBottom: scale < 1 ? `${(scale - 1) * 100}%` : undefined,
      }}
    >

      {hasStack && (
        <>
          <div className="w-[220px]">
            <TopologyNode
              variant="stack"
              name={install?.app?.name ?? 'Stack'}
              status={stackStatus}
              statusDescription={install?.runner_status_description || undefined}
              href="stacks"
            />
          </div>
          {hasSandbox && <TopologyConnector />}
          {!hasSandbox && hasComponents && <TopologyConnector variant="branch" count={Math.min(MAX_PER_ROW, layers[0]?.length ?? 1)} />}
        </>
      )}

      {hasSandbox && (
        <>
          <div className="w-[220px]">
            <TopologyNode
              variant="sandbox"
              name="Sandbox"
              status={normalizeStatus(install?.sandbox_status)}
              statusDescription={install?.sandbox_status_description || undefined}
              href="sandbox"
            />
          </div>
          {hasComponents && <TopologyConnector variant="branch" count={Math.min(MAX_PER_ROW, layers[0]?.length ?? 1)} />}
        </>
      )}

      {hasComponents && (
        <div className="flex flex-col items-center gap-0">
          {layers.map((layer, layerIndex) => {
            const chunks = chunkArray(layer, MAX_PER_ROW)
            return (
              <div key={layerIndex} className="flex flex-col items-center gap-0">
                {/* connector from previous layer (not first layer — that connector is rendered above) */}
                {layerIndex > 0 && (
                  <TopologyConnector variant="branch" count={Math.min(MAX_PER_ROW, chunks[0].length)} />
                )}
                {layerIndex === 0 && visibleTiers.length === 1 && <div className="h-2" />}

                {chunks.map((chunk, chunkIndex) => {
                  const rowWidth = chunk.length * 212 - 12
                  return (
                    <div key={chunkIndex} className="flex flex-col items-center gap-0">
                      {/* between chunks of the same layer: branch connector for continuation */}
                      {chunkIndex > 0 && (
                        <TopologyConnector variant="branch" count={chunk.length} />
                      )}
                      <div
                        className="flex items-start justify-center gap-3"
                        style={{ width: `${rowWidth}px`, minWidth: `${rowWidth}px` }}
                      >
                        {chunk.map((ic) => {
                          const displayStatus = normalizeStatus(ic.status)
                          return (
                            <div key={ic.id} className="w-[200px] shrink-0">
                              <TopologyNode
                                variant="component"
                                name={ic.component?.name ?? 'Unknown'}
                                componentType={ic.component?.type}
                                status={displayStatus}
                                statusDescription={ic.status_v2?.status_human_description || ic.status_description || undefined}
                                hasDrift={!!ic.drifted_object}
                                href={`components/${ic.component_id ?? ic.component?.id}`}
                              />
                            </div>
                          )
                        })}
                      </div>
                    </div>
                  )
                })}
              </div>
            )
          })}
        </div>
      )}

    </div>
    </div>
  )
}
