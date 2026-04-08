import { useCallback, useMemo } from 'react'
import { ReactFlow, useReactFlow, ReactFlowProvider, type ReactFlowInstance } from '@xyflow/react'
import { toPng } from 'html-to-image'
import '@xyflow/react/dist/style.css'

import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import type { TDiagramData } from './diagram-layout'
import { computeLayout } from './diagram-layout'
import { nodeTypes } from './diagram-nodes'

export interface IArchitectureDiagram extends TDiagramData {
  isLoading: boolean
  isError: boolean
}

const DiagramControls = ({ onExport }: { onExport: () => void }) => {
  const { zoomIn, zoomOut, fitView } = useReactFlow()

  return (
    <div className="absolute bottom-3 right-3 z-10 flex items-center gap-1" role="toolbar" aria-label="Diagram controls">
      <Button size="xs" variant="ghost" onClick={() => zoomIn()} aria-label="Zoom in">
        <Icon variant="PlusIcon" size={14} />
      </Button>
      <Button size="xs" variant="ghost" onClick={() => zoomOut()} aria-label="Zoom out">
        <Icon variant="MinusIcon" size={14} />
      </Button>
      <Button size="xs" variant="ghost" onClick={() => fitView({ padding: 0.2 })} aria-label="Fit to view">
        <Icon variant="CornersOutIcon" size={14} />
      </Button>
      <div className="w-px h-4 bg-cool-grey-300 dark:bg-dark-grey-600 mx-0.5" aria-hidden="true" />
      <Button size="xs" variant="ghost" onClick={onExport} aria-label="Export as PNG">
        <Icon variant="DownloadSimpleIcon" size={14} />
      </Button>
    </div>
  )
}

const DiagramCanvas = ({
  install,
  components,
  stack,
  appConfig,
  permissionsConfig,
  orgId,
  isLoading,
  isError,
}: IArchitectureDiagram) => {
  const nodes = useMemo(() => {
    if (!install || !components) return []
    return computeLayout({
      install,
      components: Array.isArray(components) ? components : [],
      stack: stack ?? undefined,
      appConfig: appConfig ?? undefined,
      permissionsConfig: permissionsConfig ?? undefined,
      orgId,
    })
  }, [install, components, stack, appConfig, permissionsConfig, orgId])

  const memoizedNodeTypes = useMemo(() => nodeTypes, [])

  const handleInit = useCallback((instance: ReactFlowInstance) => {
    setTimeout(() => instance.fitView({ padding: 0.2 }), 50)
  }, [])

  const handleExportPng = useCallback(() => {
    const el = document.querySelector('.react-flow') as HTMLElement
    if (!el) return

    toPng(el, { cacheBust: true, pixelRatio: 2 })
      .then((dataUrl) => {
        const img = new Image()
        img.onload = () => {
          const pad = 40
          const watermarkH = 32
          const canvas = document.createElement('canvas')
          canvas.width = img.width + pad * 2
          canvas.height = img.height + pad * 2 + watermarkH

          const ctx = canvas.getContext('2d')
          if (!ctx) return

          ctx.fillStyle = getComputedStyle(document.documentElement)
            .getPropertyValue('--background-neutral').trim() || '#F0F3F5'
          ctx.fillRect(0, 0, canvas.width, canvas.height)
          ctx.drawImage(img, pad, pad)

          ctx.globalAlpha = 0.4
          ctx.fillStyle = getComputedStyle(document.documentElement)
            .getPropertyValue('--foreground').trim() || '#19171C'
          ctx.font = '500 24px Inter, sans-serif'
          ctx.textBaseline = 'bottom'
          const timestamp = new Date().toLocaleString(undefined, {
            year: 'numeric', month: 'short', day: 'numeric',
            hour: 'numeric', minute: '2-digit',
          })
          const installName = install?.name || 'install'
          ctx.fillText(
            `Exported from Nuon · ${timestamp} · ${installName}`,
            pad, canvas.height - 10
          )
          ctx.globalAlpha = 1

          const a = document.createElement('a')
          a.download = `${install?.name || 'install'}-architecture.png`
          a.href = canvas.toDataURL('image/png')
          a.click()
        }
        img.src = dataUrl
      })
      .catch((err) => {
        console.error('Failed to export diagram:', err)
      })
  }, [install?.name])

  if (isLoading) {
    return (
      <div className="w-full h-full min-h-[420px] flex items-center justify-center" style={{ background: 'var(--background-neutral)' }}>
        <Skeleton width="90%" height="80%" />
      </div>
    )
  }

  if (isError || !install) {
    return (
      <div className="w-full h-full min-h-[420px] flex items-center justify-center" style={{ background: 'var(--background-neutral)' }}>
        <Text theme="neutral">
          {isError ? 'Failed to load diagram data.' : 'No install data available.'}
        </Text>
      </div>
    )
  }

  return (
    <div
      className="w-full h-full min-h-[420px] relative [&_.react-flow__node]:!cursor-default [&_.react-flow__pane]:!cursor-default"
      style={{ background: 'var(--background-neutral)' }}
    >
      <ReactFlow
        nodes={nodes}
        edges={[]}
        nodeTypes={memoizedNodeTypes}
        onInit={handleInit}
        fitView
        fitViewOptions={{ padding: 0.2 }}
        minZoom={0.3}
        maxZoom={1.5}
        nodesDraggable={false}
        nodesConnectable={false}
        elementsSelectable={false}
        panOnDrag={false}
        panOnScroll
        zoomOnScroll={false}
        zoomOnPinch
        proOptions={{ hideAttribution: true }}
      />
      <DiagramControls onExport={handleExportPng} />
    </div>
  )
}

export const ArchitectureDiagram = (props: IArchitectureDiagram) => (
  <ReactFlowProvider>
    <DiagramCanvas {...props} />
  </ReactFlowProvider>
)
