import { useMemo, type HTMLAttributes } from 'react'
import type { TTerraformPlan } from '@/types'
import { cn } from '@/utils/classnames'
import {
  TERRAFORM_DEFAULT_DIFF_OPERATIONS,
  TERRAFORM_DIFF_OPERATIONS,
  terraformPlanDiff,
} from '@/lib/diffs/terraform'
import type { IPlanDiffGroup } from '@/lib/diffs'
import { Card } from '@/components/common/Card'
import { Text } from '@/components/common/Text'
import { usePlanDiffFilter } from './use-plan-diff-filter'
import { DiffSummary } from './DiffSummary'
import { DiffFilter } from './DiffFilter'
import { DiffSection } from './DiffSection'
import { DiffSections } from './DiffSections'
import { DiffEmptyState } from './DiffEmptyState'

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
}: {
  group: IPlanDiffGroup
  defaultOpen?: boolean
}) => {
  const filter = usePlanDiffFilter(
    group.sections,
    TERRAFORM_DIFF_OPERATIONS,
    TERRAFORM_DEFAULT_DIFF_OPERATIONS
  )

  if (!group.sections.length) return null

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
      {visible.map((group) => (
        <TerraformDiffGroup
          key={group.id}
          group={group}
          defaultOpen={defaultOpen}
        />
      ))}
    </div>
  )
}
