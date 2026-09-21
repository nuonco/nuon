import type { HTMLAttributes } from 'react'
import { cn } from '@/utils/classnames'
import type { TDiffOperation } from '@/lib/diffs'
import { SearchInput } from '@/components/common/SearchInput'
import { Text } from '@/components/common/Text'
import { FilterDropdown } from './FilterMenu'

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
  replace: 'bg-primary-500',
  delete: 'bg-diff-remove',
  read: 'bg-diff-neutral',
  'no-op': 'bg-diff-neutral',
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
    className={cn('flex min-w-0 flex-1 flex-wrap items-center gap-2', className)}
    {...props}
  >
    <SearchInput
      value={searchValue}
      placeholder={searchPlaceholder}
      aria-label={searchPlaceholder}
      onChange={onSearchChange}
      labelClassName="min-w-56 flex-1"
      className="!h-8 md:min-w-0 w-full"
    />

    <Text variant="subtext" theme="neutral" className="whitespace-nowrap">
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
