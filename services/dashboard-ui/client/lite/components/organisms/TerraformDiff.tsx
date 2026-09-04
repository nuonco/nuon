import { useMemo, type HTMLAttributes } from 'react'
import type { TTerraformPlan } from '@/types'
import { cn } from '@/utils/classnames'
import { usePlanDiffFilter } from '../../hooks/use-plan-diff-filter'
import {
  TERRAFORM_DEFAULT_DIFF_OPERATIONS,
  TERRAFORM_DIFF_OPERATIONS,
  terraformPlanDiff,
} from '../../lib/diffs/terraform'
import type { IPlanDiffGroup } from '../../lib/diffs'
import { Card } from '../atoms/Card'
import { Text } from '../atoms/Text'
import { DiffSummary } from '../molecules/DiffSummary'
import { DiffFilter } from './DiffFilter'
import { DiffSection } from './DiffSection'
import { DiffSections } from './DiffSections'

export interface ITerraformDiff
  extends Omit<HTMLAttributes<HTMLDivElement>, 'children'> {
  plan?: TTerraformPlan
  defaultOpen?: boolean
}

const TerraformDiffGroup = ({
  group,
  defaultOpen,
}: {
  group: IPlanDiffGroup
  defaultOpen: boolean
}) => {
  const filter = usePlanDiffFilter(
    group.sections,
    TERRAFORM_DIFF_OPERATIONS,
    TERRAFORM_DEFAULT_DIFF_OPERATIONS
  )

  if (!group.sections.length) return null

  return (
    <Card as="section" padding="sm" className="flex flex-col gap-3">
      <header className="flex flex-wrap items-start justify-between gap-3 px-1">
        <Text as="h2" variant="heading">
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
            title={
              group.id === 'terraform-drift'
                ? 'drift'
                : group.id === 'terraform-outputs'
                  ? 'outputs'
                  : 'resources'
            }
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
    </Card>
  )
}

export const TerraformDiff = ({
  plan,
  defaultOpen = true,
  className,
  ...props
}: ITerraformDiff) => {
  const groups = useMemo(() => terraformPlanDiff(plan), [plan])
  const visible = [groups.drift, groups.resources, groups.outputs]

  if (!plan) {
    return (
      <Card
        as="section"
        padding="sm"
        className={cn('flex flex-col gap-3', className)}
        {...props}
      >
        <Text as="p" variant="body" weight="medium">
          No Terraform plan data available
        </Text>
      </Card>
    )
  }

  if (!visible.some((group) => group.sections.length)) {
    return (
      <Card
        as="section"
        padding="sm"
        className={cn('flex flex-col gap-3', className)}
        {...props}
      >
        <Text as="h2" variant="heading">
          Terraform changes
        </Text>
        <div className="rounded-lg bg-surface-02 px-4 py-8 text-center">
          <Text as="p" variant="body" weight="medium">
            No changes to show
          </Text>
        </div>
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
