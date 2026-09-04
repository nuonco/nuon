import { useMemo, type ReactNode } from 'react'
import { cn } from '@/utils/classnames'
import { changeCounts, type TDiffOperation } from '../../lib/diffs'
import { Text } from '../atoms/Text'
import { Diff, type IDiff, type TDiffView } from '../molecules/Diff'
import { Disclosure, type IDisclosure } from '../molecules/Disclosure'

export type { TDiffOperation } from '../../lib/diffs'

const RAIL_CLASSES: Record<TDiffOperation, string> = {
  create: 'border-l-diff-add',
  update: 'border-l-diff-change',
  replace: 'border-l-divider-accent',
  delete: 'border-l-diff-remove',
  read: 'border-l-diff-neutral',
  'no-op': 'border-l-diff-neutral',
}

const TINT_CLASSES: Record<TDiffOperation, string> = {
  create: 'bg-diff-add-section',
  update: 'bg-diff-change-section',
  replace: 'bg-surface-accent',
  delete: 'bg-diff-remove-section',
  read: 'bg-diff-neutral-section',
  'no-op': 'bg-diff-neutral-section',
}

const HOVER_CLASSES: Record<TDiffOperation, string> = {
  create: 'hover:bg-diff-add-row!',
  update: 'hover:bg-diff-change-row!',
  replace: '',
  delete: 'hover:bg-diff-remove-row!',
  read: 'hover:bg-diff-neutral-row!',
  'no-op': 'hover:bg-diff-neutral-row!',
}

export interface IDiffSection
  extends Omit<IDisclosure, 'children' | 'status'>,
    Omit<IDiff, 'className' | 'view'> {
  operation: TDiffOperation
  view?: TDiffView
  note?: ReactNode
  error?: ReactNode
}

export const DiffSection = ({
  operation,
  before,
  after,
  language,
  filename,
  view: viewProp,
  defaultWrap,
  lineNumbers,
  search,
  maxHeight,
  note,
  error,
  className,
  headerClassName,
  contentClassName,
  ...props
}: IDiffSection) => {
  const counts = useMemo(() => changeCounts(before, after), [after, before])

  return (
    <Disclosure
      status={
        <span
          className="flex items-center gap-1.5"
          aria-label={`${counts.added} added, ${counts.removed} removed`}
        >
          {counts.added ? (
            <Text variant="caption" family="mono" className="text-diff-add">
              +{counts.added}
            </Text>
          ) : null}
          {counts.removed ? (
            <Text variant="caption" family="mono" className="text-diff-remove">
              -{counts.removed}
            </Text>
          ) : null}
        </span>
      }
      className={cn(
        'overflow-hidden rounded-r-md border-l-4',
        RAIL_CLASSES[operation],
        TINT_CLASSES[operation],
        className
      )}
      headerClassName={cn(
        'rounded-none',
        HOVER_CLASSES[operation],
        headerClassName
      )}
      contentClassName={cn('bg-surface-01 p-2', contentClassName)}
      {...props}
    >
      {note || error ? (
        <Text as="div" variant="caption" color="tertiary" className="px-1 py-2">
          {note ?? error}
        </Text>
      ) : null}
      {before || after ? (
        <Diff
          before={before}
          after={after}
          language={language}
          filename={filename}
          view={viewProp}
          defaultWrap={defaultWrap}
          lineNumbers={lineNumbers}
          search={search}
          maxHeight={maxHeight}
        />
      ) : null}
    </Disclosure>
  )
}
