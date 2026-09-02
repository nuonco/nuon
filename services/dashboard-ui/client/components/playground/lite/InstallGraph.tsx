import { cn } from '@/utils/classnames'
import { Block } from './Block'
import { labelWidth } from './utils'

export interface IGraphNode {
  id: string
  label: string
  kind: 'install' | 'stack' | 'access' | 'sandbox' | 'runner' | 'component'
  x: number
  y: number
  width: number
}

const NODE_HEIGHT = 72
const CANVAS_WIDTH = 1080
const CANVAS_HEIGHT = 420

export const graphNodes: IGraphNode[] = [
  {
    id: 'install',
    label: 'Install',
    kind: 'install',
    x: 0,
    y: 174,
    width: 150,
  },
  { id: 'stack', label: 'Stack', kind: 'stack', x: 175, y: 174, width: 140 },
  { id: 'access', label: 'Access', kind: 'access', x: 350, y: 174, width: 140 },
  {
    id: 'sandbox',
    label: 'Sandbox',
    kind: 'sandbox',
    x: 525,
    y: 174,
    width: 140,
  },
  { id: 'runner', label: 'Runner', kind: 'runner', x: 700, y: 174, width: 140 },
  {
    id: 'cmp-01',
    label: 'Component',
    kind: 'component',
    x: 890,
    y: 4,
    width: 180,
  },
  {
    id: 'cmp-02',
    label: 'Component',
    kind: 'component',
    x: 890,
    y: 114,
    width: 180,
  },
  {
    id: 'cmp-03',
    label: 'Component',
    kind: 'component',
    x: 890,
    y: 224,
    width: 180,
  },
  {
    id: 'cmp-04',
    label: 'Component',
    kind: 'component',
    x: 890,
    y: 334,
    width: 180,
  },
]

const graphEdges: [string, string][] = [
  ['install', 'stack'],
  ['stack', 'access'],
  ['access', 'sandbox'],
  ['sandbox', 'runner'],
  ['runner', 'cmp-01'],
  ['runner', 'cmp-02'],
  ['runner', 'cmp-03'],
  ['runner', 'cmp-04'],
]

const nodeById = (id: string) =>
  graphNodes.find((node) => node.id === id) as IGraphNode

const edgePath = (fromId: string, toId: string) => {
  const from = nodeById(fromId)
  const to = nodeById(toId)
  const x1 = from.x + from.width
  const y1 = from.y + NODE_HEIGHT / 2
  const x2 = to.x
  const y2 = to.y + NODE_HEIGHT / 2
  const midX = (x1 + x2) / 2

  return `M ${x1} ${y1} C ${midX} ${y1}, ${midX} ${y2}, ${x2} ${y2}`
}

export interface IInstallGraph {
  selectedId?: string
  onSelect: (node: IGraphNode) => void
}

export const InstallGraph = ({ selectedId, onSelect }: IInstallGraph) => (
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
        {graphEdges.map(([from, to]) => (
          <path
            key={`${from}-${to}`}
            d={edgePath(from, to)}
            fill="none"
            stroke="currentColor"
            strokeWidth={2}
          />
        ))}
      </svg>

      {graphNodes.map((node) => (
        <button
          key={node.id}
          type="button"
          title={node.label}
          onClick={() => onSelect(node)}
          className={cn(
            'absolute flex flex-col justify-center gap-2 rounded-lg p-3 text-left transition-colors',
            'bg-cool-grey-100 dark:bg-dark-grey-800',
            'hover:bg-cool-grey-200 dark:hover:bg-dark-grey-700',
            selectedId === node.id && 'bg-cool-grey-300 dark:bg-dark-grey-600'
          )}
          style={{
            left: node.x,
            top: node.y,
            width: node.width,
            height: NODE_HEIGHT,
          }}
        >
          <Block
            className="h-[10px]"
            style={{ width: labelWidth(node.label) }}
          />
          <div className="flex items-center gap-2">
            <Block className="h-[12px] w-[12px] rounded-full" />
            <Block className="h-[8px] w-[56px] opacity-50" />
          </div>
        </button>
      ))}
    </div>
  </div>
)
