'use client'

import React, { useCallback, useEffect, useMemo, useState } from 'react'
import {
  ReactFlow,
  Background,
  Controls,
  useNodesState,
  useEdgesState,
  addEdge,
  type Node,
} from '@xyflow/react'
import dagre from '@dagrejs/dagre'
import '@xyflow/react/dist/style.css'
import { Duration } from '@/components/common/Duration'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { useSystemTheme } from '@/hooks/use-system-theme'
import type { THydratedActionRunSteps } from '@/utils/action-utils'

const NODE_WIDTH = 250
const NODE_HEIGHT = 75
const NODE_BG_COLOR_DARK = '#141217'
const NODE_BG_COLOR_LIGHT = '#FFFFFF'
const BG_COLOR_DARK = '#1D1B20'
const BG_COLOR_LIGHT = '#EAEDF0'

const getTheme = (theme: 'dark' | 'light') => ({
  nodeBGColor: theme === 'dark' ? NODE_BG_COLOR_DARK : NODE_BG_COLOR_LIGHT,
  bgColor: theme === 'dark' ? BG_COLOR_DARK : BG_COLOR_LIGHT,
  textColor: theme === 'dark' ? NODE_BG_COLOR_LIGHT : NODE_BG_COLOR_DARK,
})

export const ActionStepGraph = ({
  steps,
  onNodeClick,
}: {
  steps: THydratedActionRunSteps
  onNodeClick?: (data) => void
}) => {
  const theme = useSystemTheme()
  const [colors, setColors] = useState(getTheme(theme))

  useEffect(() => {
    setColors(getTheme(theme))
  }, [theme])

  const { nodes: builtNodes, edges: builtEdges } = useMemo(() => {
    const nodes = (steps || []).map((s, i) => {
      const id = s.id || String(i)
      return {
        id,
        data: {
          label: (
            <div className="flex flex-col gap-2 my-auto w-full px-2">
              <div className="flex items-center gap-2">
                <Status status={s?.status} isWithoutText variant="timeline" />
                <Text variant="body" weight="stronger">
                  {s.name || ''}
                </Text>
              </div>
              {s?.execution_duration ? (
                <Text
                  className="!inline-flex items-center gap-1"
                  variant="label"
                >
                  Finished in
                  <Duration
                    nanoseconds={s?.execution_duration}
                    variant="label"
                  />
                </Text>
              ) : null}
            </div>
          ),
          raw: s,
        },
        position: { x: 0, y: 0 },
        style: {
          display: 'flex',
          width: NODE_WIDTH,
          height: NODE_HEIGHT,
          color: colors?.textColor,
          backgroundColor: colors?.nodeBGColor,
          border: '1px solid var(--border-color)',
          borderRadius: 8,
          textAlign: 'left',
        },
        sourcePosition: 'right',
        targetPosition: 'left',
      }
    })

    const edges = (steps || []).slice(1).map((s, i) => {
      const prev = steps[i]
      const prevId = prev.id || String(i)
      const curId = s.id || String(i + 1)

      return {
        id: `e-${prevId}-${curId}`,
        source: prevId,
        target: curId,
        type: 'straight',
        animated: false,
        style: { stroke: '#3062D4', strokeWidth: 2 },
      }
    })

    return { nodes, edges }
  }, [steps, colors])

  const layoutedNodes = useMemo(() => {
    const g = new dagre.graphlib.Graph()
    g.setDefaultEdgeLabel(() => ({}))
    g.setGraph({ rankdir: 'LR' })

    builtNodes.forEach((n) => {
      g.setNode(n.id, { width: NODE_WIDTH, height: NODE_HEIGHT })
    })
    builtEdges.forEach((e) => {
      g.setEdge(e.source, e.target)
    })

    dagre.layout(g)

    return builtNodes.map((n) => {
      const nodeWithPosition = g.node(n.id)
      return {
        ...n,
        position: {
          x: (nodeWithPosition.x || 0) - NODE_WIDTH / 2,
          y: (nodeWithPosition.y || 0) - NODE_HEIGHT / 2,
        },
      }
    })
  }, [builtNodes, builtEdges, colors])

  const [nodes, setNodes, onNodesChange] = useNodesState(
    layoutedNodes as Node[]
  )
  const [edges, setEdges, onEdgesChange] = useEdgesState(builtEdges)

  const onConnect = useCallback(
    (params) =>
      setEdges((eds) => addEdge({ ...params, type: 'straight' }, eds)),
    [setEdges]
  )

  useEffect(() => {
    const g = new dagre.graphlib.Graph()
    g.setDefaultEdgeLabel(() => ({}))
    g.setGraph({ rankdir: 'LR' })

    builtNodes.forEach((n) => {
      g.setNode(n.id, { width: NODE_WIDTH, height: NODE_HEIGHT })
    })
    builtEdges.forEach((e) => {
      g.setEdge(e.source, e.target)
    })

    dagre.layout(g)

    const reLayouted = builtNodes.map((n) => {
      const nodeWithPosition = g.node(n.id)
      return {
        ...n,
        position: {
          x: (nodeWithPosition.x || 0) - NODE_WIDTH / 2,
          y: (nodeWithPosition.y || 0) - NODE_HEIGHT / 2,
        },
      }
    })

    setNodes(reLayouted as Node[])
    setEdges(builtEdges)
  }, [builtNodes, builtEdges, setNodes, setEdges, colors])

  const handleNodeClick = useCallback(
    (event, node) => {
      if (onNodeClick && node?.data?.raw) onNodeClick(node.data.raw)
    },
    [onNodeClick]
  )

  return (
    <div className="w-full h-[250px]">
      <ReactFlow
        nodes={nodes}
        edges={edges}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        onConnect={onConnect}
        onNodeClick={handleNodeClick}
        fitView
        fitViewOptions={{ padding: 0.2 }}
        minZoom={0.1}
        maxZoom={1.5}
        defaultViewport={{ x: 0, y: 0, zoom: 0 }}
        proOptions={{ hideAttribution: true }}
        style={{
          border: '1px solid var(--border-color)',
          borderRadius: '8px',
          maxHeight: '250px',
        }}
      >
        <Controls
          position="top-right"
          orientation="horizontal"
          style={{
            color: '#141217',
          }}
        />
        <Background
          bgColor={colors?.bgColor}
          color={colors?.textColor}
          gap={16}
        />
      </ReactFlow>
    </div>
  )
}
