import { cn } from '@/utils/classnames'
import { Block } from './Block'
import { labelWidth } from './utils'

export type TBranchNodeKind =
  | 'source'
  | 'stack'
  | 'component'
  | 'sandbox'
  | 'action'
  | 'runbook'
  | 'role'
  | 'policy'
  | 'build'
  | 'group'
  | 'install'

export interface IBranchNode {
  id: string
  label: string
  kind: TBranchNodeKind
  x: number
  y: number
  width: number
  height: number
}

const CONFIG_X = 200
const BUILD_X = 430
const GROUP_X = 640
const INSTALL_X = 850

export interface IConfigItem {
  id: string
  label: string
  kind: TBranchNodeKind
}

export const configItems: IConfigItem[] = [
  { id: 'cmp-01', label: 'api', kind: 'component' },
  { id: 'cmp-02', label: 'worker', kind: 'component' },
  { id: 'cmp-03', label: 'web', kind: 'component' },
  { id: 'sandbox', label: 'Sandbox', kind: 'sandbox' },
  { id: 'act-01', label: 'Rotate creds', kind: 'action' },
  { id: 'act-02', label: 'Health check', kind: 'action' },
  { id: 'rbk-01', label: 'Restore', kind: 'runbook' },
  { id: 'rol-01', label: 'deploy-role', kind: 'role' },
]

const groupItems = [
  { id: 'grp-01', label: 'Canary', installs: ['install-01'] },
  { id: 'grp-02', label: 'Wave 1', installs: ['install-02', 'install-03'] },
  { id: 'grp-03', label: 'Wave 2', installs: ['install-04', 'install-05'] },
]

const CONFIG_PITCH = 62
const INSTALL_PITCH = 62

const configNodes: IBranchNode[] = configItems.map((item, i) => ({
  ...item,
  x: CONFIG_X,
  y: i * CONFIG_PITCH,
  width: 170,
  height: 48,
}))

const installList = groupItems.flatMap((group) =>
  group.installs.map((install) => ({ install, groupId: group.id }))
)

const installNodes: IBranchNode[] = installList.map((entry, i) => ({
  id: entry.install,
  label: entry.install,
  kind: 'install',
  x: INSTALL_X,
  y: i * INSTALL_PITCH,
  width: 160,
  height: 48,
}))

const configBottom = configNodes.at(-1)!.y + 48
const installBottom = installNodes.at(-1)!.y + 48
const CANVAS_HEIGHT = Math.max(configBottom, installBottom) + 8
const CANVAS_WIDTH = INSTALL_X + 160

const centerY = CANVAS_HEIGHT / 2 - 28

const sourceNode: IBranchNode = {
  id: 'source',
  label: 'main',
  kind: 'source',
  x: 0,
  y: centerY,
  width: 160,
  height: 56,
}

const buildNode: IBranchNode = {
  id: 'build',
  label: 'Build',
  kind: 'build',
  x: BUILD_X,
  y: centerY,
  width: 150,
  height: 56,
}

const groupNodes: IBranchNode[] = groupItems.map((group, i) => ({
  id: group.id,
  label: group.label,
  kind: 'group',
  x: GROUP_X,
  y: 40 + i * ((CANVAS_HEIGHT - 100) / groupItems.length),
  width: 150,
  height: 48,
}))

export const branchNodes: IBranchNode[] = [
  sourceNode,
  ...configNodes,
  buildNode,
  ...groupNodes,
  ...installNodes,
]

const nodeById = (id: string) =>
  branchNodes.find((node) => node.id === id) as IBranchNode

const edges: [string, string][] = [
  ...configNodes.map((node) => ['source', node.id] as [string, string]),
  ...configNodes.map((node) => [node.id, 'build'] as [string, string]),
  ...groupNodes.map((node) => ['build', node.id] as [string, string]),
  ...installList.map(
    (entry) => [entry.groupId, entry.install] as [string, string]
  ),
]

const edgePath = (fromId: string, toId: string) => {
  const from = nodeById(fromId)
  const to = nodeById(toId)
  const x1 = from.x + from.width
  const y1 = from.y + from.height / 2
  const x2 = to.x
  const y2 = to.y + to.height / 2
  const midX = (x1 + x2) / 2

  return `M ${x1} ${y1} C ${midX} ${y1}, ${midX} ${y2}, ${x2} ${y2}`
}

export interface IBranchGraph {
  selectedId?: string
  onSelect: (node: IBranchNode) => void
}

export const BranchGraph = ({ selectedId, onSelect }: IBranchGraph) => (
  <div className="overflow-x-auto">
    <div
      className="relative"
      style={{ width: CANVAS_WIDTH, height: CANVAS_HEIGHT }}
    >
      <svg
        className="absolute inset-0 text-cool-grey-400 dark:text-dark-grey-400"
        width={CANVAS_WIDTH}
        height={CANVAS_HEIGHT}
      >
        {edges.map(([from, to]) => (
          <path
            key={`${from}-${to}`}
            d={edgePath(from, to)}
            fill="none"
            stroke="currentColor"
            strokeWidth={2}
            opacity={0.6}
          />
        ))}
      </svg>

      {branchNodes.map((node) => (
        <button
          key={node.id}
          type="button"
          title={node.label}
          onClick={() => onSelect(node)}
          className={cn(
            'absolute flex flex-col justify-center gap-1.5 rounded-lg p-3 text-left transition-colors',
            node.kind === 'source' || node.kind === 'build'
              ? 'bg-cool-grey-200 dark:bg-dark-grey-700'
              : 'bg-cool-grey-100 dark:bg-dark-grey-800',
            'hover:bg-cool-grey-300 dark:hover:bg-dark-grey-600',
            selectedId === node.id && 'bg-cool-grey-300 dark:bg-dark-grey-600'
          )}
          style={{
            left: node.x,
            top: node.y,
            width: node.width,
            height: node.height,
          }}
        >
          <Block
            className="h-[10px]"
            text={node.label}
            style={{ width: labelWidth(node.label) }}
          />
          <div className="flex items-center gap-2">
            <Block className="h-[8px] w-[8px] rounded-full" />
            <Block className="h-[8px] w-[44px] opacity-50" />
          </div>
        </button>
      ))}
    </div>
  </div>
)
