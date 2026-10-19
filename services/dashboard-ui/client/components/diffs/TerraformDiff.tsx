import { useMemo, useState, type HTMLAttributes } from 'react'
import type { TTerraformPlan } from '@/types'
import { cn } from '@/utils/classnames'
import {
  TERRAFORM_DEFAULT_DIFF_OPERATIONS,
  TERRAFORM_DIFF_OPERATIONS,
  terraformPlanDiff,
} from '@/lib/diffs/terraform'
import { terraformActionSelected } from '@/lib/diffs/noop-visibility'
import type { IPlanDiffGroup, IPlanDiffSection } from '@/lib/diffs'
import { Card } from '@/components/common/Card'
import { Text } from '@/components/common/Text'
import { usePlanDiffFilter } from './use-plan-diff-filter'
import { DiffSummary } from './DiffSummary'
import { DiffFilter } from './DiffFilter'
import { DiffSection } from './DiffSection'
import { DiffSections } from './DiffSections'
import { DiffEmptyState } from './DiffEmptyState'
import { ShowNoopsToggle } from './ShowNoopsToggle'

const FILTER_OPERATIONS = TERRAFORM_DIFF_OPERATIONS.filter(
  (operation) => operation !== 'no-op'
)

const sectionMatches = (section: IPlanDiffSection, query: string) =>
  !query ||
  section.searchable.some((value) => value.toLowerCase().includes(query))

export interface ITerraformDiff
  extends Omit<HTMLAttributes<HTMLDivElement>, 'children'> {
  plan?: TTerraformPlan
  defaultOpen?: boolean
}

const GROUP_TITLES: Record<string, string> = {
  'terraform-drift': 'drift',
  'terraform-outputs': 'outputs',
}

const TerraformDiffGroup = ({
  group,
  defaultOpen,
  showNoops,
}: {
  group: IPlanDiffGroup
  defaultOpen?: boolean
  showNoops: boolean
}) => {
  const filter = usePlanDiffFilter(
    group.sections,
    FILTER_OPERATIONS,
    TERRAFORM_DEFAULT_DIFF_OPERATIONS
  )
  const query = filter.searchQuery.trim().toLowerCase()
  const visibleSections = group.sections.filter(
    (section) =>
      terraformActionSelected(
        section.operation,
        filter.selectedOperations,
        showNoops
      ) && sectionMatches(section, query)
  )
  const totalCount = group.sections.filter(
    (section) => showNoops || section.operation !== 'no-op'
  ).length

  if (!totalCount) return null

  return (
    <Card className="flex flex-col gap-3">
      <header className="flex flex-wrap items-start justify-between gap-3 px-1">
        <Text as="h2" variant="h3">
          {group.title}
        </Text>
        <DiffSummary
          summary={group.summary}
          operations={TERRAFORM_DIFF_OPERATIONS}
        />
      </header>

      <DiffSections
        defaultOpen={defaultOpen}
        toolbar={
          <DiffFilter
            title={GROUP_TITLES[group.id] ?? 'resources'}
            operations={filter.operations}
            selectedOperations={filter.selectedOperations}
            selectedCount={visibleSections.length}
            totalCount={totalCount}
            searchValue={filter.searchQuery}
            searchPlaceholder={group.searchPlaceholder}
            onSearchChange={filter.setSearchQuery}
            onOperationToggle={filter.toggleOperation}
            onOperationOnly={filter.onlyOperation}
            onReset={filter.reset}
          />
        }
      >
        {visibleSections.length ? (
          visibleSections.map((section) => (
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
    </Card>
  )
}

export const TerraformDiff = ({
  plan,
  defaultOpen,
  className,
  ...props
}: ITerraformDiff) => {
  const groups = useMemo(() => terraformPlanDiff(plan), [plan])
  const visible = [groups.drift, groups.resources, groups.outputs]
  const [showNoops, setShowNoops] = useState(false)
  const hasNoops = visible.some((group) =>
    group.sections.some((section) => section.operation === 'no-op')
  )

  if (!plan) {
    return (
      <Card className={cn('flex flex-col gap-3', className)} {...props}>
        <Text as="p" variant="body" weight="strong">
          No Terraform plan data available
        </Text>
      </Card>
    )
  }

  if (!visible.some((group) => group.sections.length)) {
    return (
      <Card className={cn('flex flex-col gap-3', className)} {...props}>
        <Text as="h2" variant="h3">
          Terraform changes
        </Text>
        <DiffEmptyState />
      </Card>
    )
  }

  return (
    <div className={cn('flex flex-col gap-3', className)} {...props}>
      {hasNoops ? (
        <div className="flex justify-end">
          <ShowNoopsToggle checked={showNoops} onChange={setShowNoops} />
        </div>
      ) : null}
      {visible.map((group) => (
        <TerraformDiffGroup
          key={group.id}
          group={group}
          defaultOpen={defaultOpen}
          showNoops={showNoops}
        />
      ))}
    </div>
  )
}
