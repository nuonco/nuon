import { cn } from '@/utils/classnames'

const LINE_COLOR = 'border-cool-grey-300 dark:border-cool-grey-700'
const DOT = 'bg-cool-grey-300 dark:bg-cool-grey-600'

interface ITopologyConnectorProps {
  variant?: 'straight' | 'branch'
  count?: number
}

export function TopologyConnector({
  variant = 'straight',
  count = 1,
}: ITopologyConnectorProps) {
  if (variant === 'straight') {
    return (
      <div className="flex flex-col items-center">
        <div className={cn('w-0 h-6 border-l', LINE_COLOR)} />
        <div className={cn('w-1.5 h-1.5 rounded-full', DOT)} />
      </div>
    )
  }

  if (count <= 1) {
    return (
      <div className="flex flex-col items-center">
        <div className={cn('w-0 h-6 border-l', LINE_COLOR)} />
        <div className={cn('w-1.5 h-1.5 rounded-full', DOT)} />
        <div className={cn('w-0 h-3 border-l', LINE_COLOR)} />
      </div>
    )
  }

  return (
    <div className="flex flex-col items-center w-full">
      <div className={cn('w-0 h-6 border-l', LINE_COLOR)} />
      <div className={cn('w-1.5 h-1.5 rounded-full', DOT)} />
      <div className="relative w-full flex justify-center">
        <div
          className={cn('absolute top-0 h-0 border-t', LINE_COLOR)}
          style={{
            width: `min(100%, ${(count - 1) * 226 + 100}px)`,
            left: '50%',
            transform: 'translateX(-50%)',
          }}
        />
        <div className="flex justify-center gap-4" style={{ width: `${count * 226}px` }}>
          {Array.from({ length: count }).map((_, i) => (
            <div key={i} className="flex flex-col items-center" style={{ width: '210px' }}>
              <div className={cn('w-0 h-4 border-l', LINE_COLOR)} />
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
