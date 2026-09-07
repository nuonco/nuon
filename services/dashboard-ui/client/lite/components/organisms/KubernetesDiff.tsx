import { useMemo, type HTMLAttributes } from 'react'
import type { TKubernetesPlan } from '@/types'
import { cn } from '@/utils/classnames'
import { usePlanDiffFilter } from '../../hooks/use-plan-diff-filter'
import {
  KUBERNETES_DIFF_OPERATIONS,
  kubernetesPlanDiff,
} from '../../lib/diffs/kubernetes'
import { Card } from '../atoms/Card'
import { Text } from '../atoms/Text'
import { DiffSummary } from '../molecules/DiffSummary'
import { DiffFilter } from './DiffFilter'
import { DiffSection } from './DiffSection'
import { DiffSections } from './DiffSections'

export interface IKubernetesDiff
  extends Omit<HTMLAttributes<HTMLDivElement>, 'children'> {
  plan?: TKubernetesPlan
  defaultOpen?: boolean
}

export const KubernetesDiff = ({
  plan,
  defaultOpen = true,
  className,
  ...props
}: IKubernetesDiff) => {
  const group = useMemo(() => kubernetesPlanDiff(plan), [plan])
  const filter = usePlanDiffFilter(group.sections, KUBERNETES_DIFF_OPERATIONS)

  return (
    <Card
      as="section"
      padding="sm"
      className={cn('flex flex-col gap-3', className)}
      {...props}
    >
      <header className="flex flex-wrap items-start justify-between gap-3 px-1">
        <span className="flex min-w-0 flex-col">
          <Text as="h2" variant="heading">
            {group.title}
          </Text>
          {group.description ? (
            <Text variant="caption" color="tertiary">
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
          <div className="rounded-lg bg-surface-02 px-4 py-8 text-center">
            <Text as="p" variant="body" weight="medium">
              No changes to show
            </Text>
            <Text as="p" variant="caption" color="tertiary">
              Clear the search or reset the filters.
            </Text>
          </div>
        )}
      </DiffSections>
    </Card>
  )
}
