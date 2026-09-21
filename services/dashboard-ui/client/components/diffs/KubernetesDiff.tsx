import { useMemo, type HTMLAttributes } from 'react'
import type { TKubernetesPlan } from '@/types'
import { cn } from '@/utils/classnames'
import {
  KUBERNETES_DIFF_OPERATIONS,
  kubernetesPlanDiff,
} from '@/lib/diffs/kubernetes'
import { Card } from '@/components/common/Card'
import { Text } from '@/components/common/Text'
import { usePlanDiffFilter } from './use-plan-diff-filter'
import { DiffSummary } from './DiffSummary'
import { DiffFilter } from './DiffFilter'
import { DiffSection } from './DiffSection'
import { DiffSections } from './DiffSections'
import { DiffEmptyState } from './DiffEmptyState'

export interface IKubernetesDiff
  extends Omit<HTMLAttributes<HTMLDivElement>, 'children'> {
  plan?: TKubernetesPlan
  defaultOpen?: boolean
}

export const KubernetesDiff = ({
  plan,
  defaultOpen,
  className,
  ...props
}: IKubernetesDiff) => {
  const group = useMemo(() => kubernetesPlanDiff(plan), [plan])
  const filter = usePlanDiffFilter(group.sections, KUBERNETES_DIFF_OPERATIONS)

  return (
    <Card className={cn('flex flex-col gap-3', className)} {...props}>
      <header className="flex flex-wrap items-start justify-between gap-3 px-1">
        <span className="flex min-w-0 flex-col">
          <Text as="h2" variant="h3">
            {group.title}
          </Text>
          {group.description ? (
            <Text variant="subtext" theme="neutral">
              {group.description}
            </Text>
          ) : null}
        </span>
        <DiffSummary
          summary={group.summary}
          operations={KUBERNETES_DIFF_OPERATIONS}
        />
      </header>

      <DiffSections
        defaultOpen={defaultOpen}
        toolbar={
          <DiffFilter
            title="changes"
            operations={filter.operations}
            selectedOperations={filter.selectedOperations}
            selectedCount={filter.selectedCount}
            totalCount={filter.totalCount}
            searchValue={filter.searchQuery}
            searchPlaceholder={group.searchPlaceholder}
            onSearchChange={filter.setSearchQuery}
            onOperationToggle={filter.toggleOperation}
            onOperationOnly={filter.onlyOperation}
            onReset={filter.reset}
          />
        }
      >
        {filter.filteredSections.length ? (
          filter.filteredSections.map((section) => (
            <DiffSection
              key={section.id}
              title={section.title}
              description={section.description}
              operation={section.operation}
              before={section.before}
              after={section.after}
              language={section.language}
              filename={section.filename}
              error={section.error}
            />
          ))
        ) : (
          <DiffEmptyState />
        )}
      </DiffSections>
    </Card>
  )
}
