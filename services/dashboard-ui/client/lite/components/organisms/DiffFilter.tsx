import type { HTMLAttributes } from 'react'
import { cn } from '@/utils/classnames'
import type { TDiffOperation } from '../../lib/diffs'
import { Text } from '../atoms/Text'
import { FilterDropdown } from '../molecules/FilterMenu'
import { SearchInput } from '../molecules/SearchInput'

const LABELS: Record<TDiffOperation, string> = {
  create: 'Create',
  update: 'Update',
  replace: 'Replace',
  delete: 'Delete',
  read: 'Read',
  'no-op': 'Unchanged',
}

const RAIL_CLASSES: Record<TDiffOperation, string> = {
  create: 'bg-diff-add',
  update: 'bg-diff-change',
  replace: 'bg-accent',
  delete: 'bg-diff-remove',
  read: 'bg-secondary',
  'no-op': 'bg-tertiary',
}

export interface IDiffFilter
  extends Omit<HTMLAttributes<HTMLDivElement>, 'onChange'> {
  title: string
  operations: readonly TDiffOperation[]
  selectedOperations: Set<TDiffOperation>
  selectedCount: number
  totalCount: number
  searchValue: string
  searchPlaceholder: string
  onSearchChange: (value: string) => void
  onOperationToggle: (operation: TDiffOperation) => void
  onOperationOnly: (operation: TDiffOperation) => void
  onReset: () => void
}

export const DiffFilter = ({
  title,
  operations,
  selectedOperations,
  selectedCount,
  totalCount,
  searchValue,
  searchPlaceholder,
  onSearchChange,
  onOperationToggle,
  onOperationOnly,
  onReset,
  className,
  ...props
}: IDiffFilter) => (
  <div
    className={cn(
      'flex min-w-0 flex-1 flex-wrap items-center gap-2',
      className
    )}
    {...props}
  >
    <SearchInput
      size="sm"
      value={searchValue}
      placeholder={searchPlaceholder}
      aria-label={searchPlaceholder}
      onValueChange={onSearchChange}
      className="min-w-56 flex-1"
    />

    <Text variant="caption" color="tertiary" className="whitespace-nowrap">
      {selectedCount} of {totalCount} {title}
    </Text>

    <FilterDropdown
      label={`Filter ${title}`}
      options={operations.map((operation) => ({
        value: operation,
        label: LABELS[operation],
        rail: RAIL_CLASSES[operation],
      }))}
      selected={selectedOperations}
      onToggle={onOperationToggle}
      onIsolate={onOperationOnly}
      onReset={onReset}
    />
  </div>
)
