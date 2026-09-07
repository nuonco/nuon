import { useMemo, type HTMLAttributes } from 'react'
import type { TPulumiPlan } from '@/types'
import { cn } from '@/utils/classnames'
import { usePlanDiffFilter } from '../../hooks/use-plan-diff-filter'
import {
  PULUMI_DEFAULT_DIFF_OPERATIONS,
  PULUMI_DIFF_OPERATIONS,
  pulumiPlanDiff,
} from '../../lib/diffs/pulumi'
import type { IPlanDiffDiagnostic } from '../../lib/diffs'
import { Card } from '../atoms/Card'
import { Text } from '../atoms/Text'
import { CodeBlock } from '../molecules/CodeBlock'
import { DiffSummary } from '../molecules/DiffSummary'
import { DiffFilter } from './DiffFilter'
import { DiffSection } from './DiffSection'
import { DiffSections } from './DiffSections'

export interface IPulumiDiff
  extends Omit<HTMLAttributes<HTMLDivElement>, 'children'> {
  plan?: TPulumiPlan
  defaultOpen?: boolean
}

const Diagnostic = ({ diagnostic }: { diagnostic: IPlanDiffDiagnostic }) => (
  <div
    className={cn(
      'rounded-r-md border-l-4 px-3 py-2',
      diagnostic.severity === 'error'
        ? 'border-l-diff-remove bg-diff-remove-section'
        : diagnostic.severity === 'warning'
          ? 'border-l-diff-change bg-diff-change-section'
          : 'border-l-diff-neutral bg-diff-neutral-section'
    )}
  >
    <Text as="p" variant="caption" family="mono">
      {diagnostic.message}
    </Text>
  </div>
)

export const PulumiDiff = ({
  plan,
  defaultOpen = true,
  className,
  ...props
}: IPulumiDiff) => {
  const group = useMemo(() => pulumiPlanDiff(plan), [plan])
  const filter = usePlanDiffFilter(
    group.sections,
    PULUMI_DIFF_OPERATIONS,
    PULUMI_DEFAULT_DIFF_OPERATIONS
  )

  if (!plan) {
    return (
      <Card
        as="section"
        padding="sm"
        className={cn('flex flex-col gap-3', className)}
        {...props}
      >
        <Text as="p" variant="body" weight="medium">
          No Pulumi preview data available
        </Text>
      </Card>
    )
  }

  return (
    <Card
      as="section"
      padding="sm"
      className={cn('flex flex-col gap-3', className)}
      {...props}
    >
      <header className="flex flex-wrap items-start justify-between gap-3 px-1">
        <Text as="h2" variant="heading">
          {group.title}
        </Text>
        <DiffSummary
          summary={group.summary}
          operations={PULUMI_DIFF_OPERATIONS}
        />
      </header>

      {group.sections.length ? (
        <DiffSections
          defaultOpen={defaultOpen}
          toolbar={
            <DiffFilter
              title="resources"
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
                note={section.note}
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
      ) : plan.stdout ? (
        <CodeBlock value={plan.stdout} language="text" filename="pulumi.txt" />
      ) : (
        <div className="rounded-lg bg-surface-02 px-4 py-8 text-center">
          <Text as="p" variant="body" weight="medium">
            No changes to show
          </Text>
        </div>
      )}

      {group.diagnostics?.length ? (
        <section className="flex flex-col gap-2 px-1">
          <Text as="h3" variant="label">
            Diagnostics
          </Text>
          {group.diagnostics.map((diagnostic) => (
            <Diagnostic key={diagnostic.id} diagnostic={diagnostic} />
          ))}
        </section>
      ) : null}
    </Card>
  )
}
