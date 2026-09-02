import type { HTMLAttributes } from 'react'
import { cn } from '@/utils/classnames'
import type { TDiffOperation } from '../../lib/diffs'
import { Icon } from '../atoms/Icon'
import { Input } from '../atoms/Input'
import { Text } from '../atoms/Text'
import { FilterDropdown } from '../molecules/FilterMenu'

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
      'flex flex-wrap items-center gap-3 rounded-lg bg-surface-02 p-2',
      className
    )}
    {...props}
  >
    <label className="relative min-w-56 flex-1">
      <span className="sr-only">{searchPlaceholder}</span>
      <Icon
        variant="MagnifyingGlassIcon"
        size={14}
        className="pointer-events-none absolute top-1/2 left-2.5 -translate-y-1/2 text-tertiary"
      />
      <Input
        type="search"
        size="sm"
        value={searchValue}
        placeholder={searchPlaceholder}
        onChange={(event) => onSearchChange(event.target.value)}
        className="pl-8"
      />
    </label>

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
