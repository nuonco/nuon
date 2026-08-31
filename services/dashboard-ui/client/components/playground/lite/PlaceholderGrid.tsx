import { cn } from '@/utils/classnames'
import { Block } from './Block'

export interface IPlaceholderGrid {
  rows: number
  columns?: 1 | 2 | 3
  height?: string
}

export const PlaceholderGrid = ({
  rows,
  columns = 1,
  height = 'h-[4rem]',
}: IPlaceholderGrid) => (
  <div
    className={cn('grid gap-4', {
      'grid-cols-1': columns === 1,
      'grid-cols-2': columns === 2,
      'grid-cols-3': columns === 3,
    })}
  >
    {Array.from({ length: rows }, (_, i) => (
      <Block key={i} className={cn('w-full', height)} />
    ))}
  </div>
)
