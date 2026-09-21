import { useMemo, type HTMLAttributes } from 'react'
import type { TPulumiPlan } from '@/types'
import { cn } from '@/utils/classnames'
import {
  PULUMI_DEFAULT_DIFF_OPERATIONS,
  PULUMI_DIFF_OPERATIONS,
  pulumiPlanDiff,
} from '@/lib/diffs/pulumi'
import type { IPlanDiffDiagnostic } from '@/lib/diffs'
import { Card } from '@/components/common/Card'
import { Text } from '@/components/common/Text'
import { usePlanDiffFilter } from './use-plan-diff-filter'
import { CodeBlock } from './CodeBlock'
import { DiffSummary } from './DiffSummary'
import { DiffFilter } from './DiffFilter'
import { DiffSection } from './DiffSection'
import { DiffSections } from './DiffSections'
import { DiffEmptyState } from './DiffEmptyState'

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
    <Text as="p" variant="subtext" family="mono">
      {diagnostic.message}
    </Text>
  </div>
)

export const PulumiDiff = ({
  plan,
  defaultOpen,
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
      <Card className={cn('flex flex-col gap-3', className)} {...props}>
        <Text as="p" variant="body" weight="strong">
          No Pulumi preview data available
        </Text>
      </Card>
    )
  }

  return (
    <Card className={cn('flex flex-col gap-3', className)} {...props}>
      <header className="flex flex-wrap items-start justify-between gap-3 px-1">
        <Text as="h2" variant="h3">
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
            <DiffEmptyState />
          )}
        </DiffSections>
      ) : plan.stdout ? (
        <CodeBlock value={plan.stdout} language="text" filename="pulumi.txt" />
      ) : (
        <DiffEmptyState />
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
