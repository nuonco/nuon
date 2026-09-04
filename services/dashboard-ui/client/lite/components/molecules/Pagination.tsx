import { Button } from '../atoms/Button'
import { Text } from '../atoms/Text'

export interface IPagination {
  offset: number
  pageSize: number
  hasNext: boolean
  onOffsetChange: (offset: number) => void
  loading?: boolean
  label?: string
}

export const Pagination = ({
  offset,
  pageSize,
  hasNext,
  onOffsetChange,
  loading = false,
  label = 'Pagination',
}: IPagination) => {
  const safePageSize = Math.max(1, pageSize)
  const safeOffset = Math.max(0, offset)
  const page = Math.floor(safeOffset / safePageSize) + 1

  return (
    <nav
      aria-label={label}
      aria-busy={loading || undefined}
      className="flex items-center justify-center gap-3"
    >
      <Button
        size="sm"
        variant="ghost"
        disabled={loading || safeOffset === 0}
        onClick={() => onOffsetChange(Math.max(0, safeOffset - safePageSize))}
      >
        Previous
      </Button>
      <Text variant="caption" color="secondary" aria-live="polite">
        Page {page}
      </Text>
      <Button
        size="sm"
        variant="ghost"
        disabled={loading || !hasNext}
        onClick={() => onOffsetChange(safeOffset + safePageSize)}
      >
        Next
      </Button>
    </nav>
  )
}
