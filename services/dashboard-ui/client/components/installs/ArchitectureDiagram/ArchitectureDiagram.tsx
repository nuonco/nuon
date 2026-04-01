import { useCallback, useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { ReactFlow, useReactFlow, ReactFlowProvider, type ReactFlowInstance } from '@xyflow/react'
import { toPng } from 'html-to-image'
import '@xyflow/react/dist/style.css'

import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getInstallComponents, getInstallStack, getAppConfig } from '@/lib'
import { computeLayout } from './diagram-layout'
import { nodeTypes } from './diagram-nodes'

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

const DiagramCanvas = () => {
  const { org } = useOrg()
  const { install } = useInstall()

  const {
    data: components,
    isLoading: componentsLoading,
    isError: componentsError,
  } = useQuery({
    queryKey: ['install-components-diagram', org?.id, install?.id],
    queryFn: () =>
      getInstallComponents({
        orgId: org.id!,
        installId: install.id!,
        limit: 100,
        offset: 0,
      }),
    enabled: !!org?.id && !!install?.id,
    refetchInterval: 20000,
  })

  const { data: stack } = useQuery({
    queryKey: ['install-stack-diagram', org?.id, install?.id],
    queryFn: () =>
      getInstallStack({ orgId: org.id!, installId: install.id! }),
    enabled: !!org?.id && !!install?.id,
  })

  const { data: appConfig } = useQuery({
    queryKey: [
      'app-config-diagram',
      org?.id,
      install?.app_id,
      install?.app_config_id,
    ],
    queryFn: () =>
      getAppConfig({
        orgId: org.id!,
        appId: install.app_id!,
        appConfigId: install.app_config_id!,
        recurse: true,
      }),
    enabled: !!org?.id && !!install?.app_id && !!install?.app_config_id,
  })

  const nodes = useMemo(() => {
    if (!install || !components) return []
    return computeLayout({
      install,
      components: Array.isArray(components) ? components : [],
      stack: stack ?? undefined,
      appConfig: appConfig ?? undefined,
      orgId: org.id!,
    })
  }, [install, components, stack, appConfig, org.id])

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
          ctx.font = '500 20px Inter, sans-serif'
          ctx.textBaseline = 'bottom'
          ctx.fillText('Exported from', pad, canvas.height - 10)

          const textW = ctx.measureText('Exported from').width
          const logoSize = 16
          const logoX = pad + textW + 8
          const logoY = canvas.height - 10 - logoSize

          const nuonPath = new Path2D()
          const s = logoSize / 200
          nuonPath.addPath(new Path2D(
            'M121.15 40.3118L97.9645 53.715V75.4151L79.1959 64.5597H79.1852L56.8232 77.492V148.651L79.1852 161.584H79.1959L103.205 147.699V126.951L121.161 137.325L144.346 123.922V53.715L121.161 40.3118H121.15ZM62.0528 80.5216L79.1745 70.6297H79.1852L97.9538 81.4744V117.862L62.0528 97.1151V80.5216ZM97.9538 144.669L79.1745 155.514L62.0528 145.622V103.174L97.9538 123.922V144.669ZM139.095 120.881L121.15 131.255L103.205 120.892V84.504L139.106 105.251V120.881H139.095ZM139.095 99.192L103.194 78.4447V56.7447L121.15 46.3711L139.095 56.7447V99.192Z'
          ), { a: s, b: 0, c: 0, d: s, e: logoX, f: logoY })
          ctx.fill(nuonPath)
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

  if (componentsLoading) {
    return (
      <div className="w-full h-full min-h-[420px] flex items-center justify-center" style={{ background: 'var(--background-neutral)' }}>
        <Skeleton width="90%" height="80%" />
      </div>
    )
  }

  if (componentsError || !install) {
    return (
      <div className="w-full h-full min-h-[420px] flex items-center justify-center" style={{ background: 'var(--background-neutral)' }}>
        <Text theme="neutral">
          {componentsError ? 'Failed to load diagram data.' : 'No install data available.'}
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

export const ArchitectureDiagram = () => (
  <ReactFlowProvider>
    <DiagramCanvas />
  </ReactFlowProvider>
)
