import { useMemo, type HTMLAttributes } from 'react'
import type { THelmPlan } from '@/types'
import { cn } from '@/utils/classnames'
import { HELM_DIFF_OPERATIONS, helmPlanDiff } from '@/lib/diffs/helm'
import { Card } from '@/components/common/Card'
import { Text } from '@/components/common/Text'
import { usePlanDiffFilter } from './use-plan-diff-filter'
import { DiffSummary } from './DiffSummary'
import { DiffFilter } from './DiffFilter'
import { DiffSection } from './DiffSection'
import { DiffSections } from './DiffSections'
import { DiffEmptyState } from './DiffEmptyState'

export interface IHelmDiff
  extends Omit<HTMLAttributes<HTMLDivElement>, 'children'> {
  plan?: THelmPlan
  defaultOpen?: boolean
}

export const HelmDiff = ({
  plan,
  defaultOpen,
  className,
  ...props
}: IHelmDiff) => {
  const group = useMemo(() => helmPlanDiff(plan), [plan])
  const filter = usePlanDiffFilter(group.sections, HELM_DIFF_OPERATIONS)

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
        <DiffSummary summary={group.summary} operations={HELM_DIFF_OPERATIONS} />
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
