import { useMemo, type HTMLAttributes } from 'react'
import type { TAppConfigDiffSection } from '@/types'
import { cn } from '@/utils/classnames'
import { usePlanDiffFilter } from '../../hooks/use-plan-diff-filter'
import {
  APP_CONFIG_DIFF_OPERATIONS,
  appConfigPlanDiff,
  type IAppConfigDiffSummary,
} from '../../lib/diffs/app-config'
import { Card } from '../atoms/Card'
import { Text } from '../atoms/Text'
import { DiffSummary } from '../molecules/DiffSummary'
import { DiffFilter } from './DiffFilter'
import { DiffSection } from './DiffSection'
import { DiffSections } from './DiffSections'

export interface IAppConfigDiff
  extends Omit<HTMLAttributes<HTMLDivElement>, 'children'> {
  sections: TAppConfigDiffSection[]
  summary: IAppConfigDiffSummary | null
  isLoading?: boolean
  defaultSectionsOpen?: boolean
}

export const AppConfigDiff = ({
  sections,
  summary,
  isLoading = false,
  defaultSectionsOpen = true,
  className,
  ...props
}: IAppConfigDiff) => {
  const group = useMemo(
    () => appConfigPlanDiff(sections, summary),
    [sections, summary]
  )
  const filter = usePlanDiffFilter(group.sections, APP_CONFIG_DIFF_OPERATIONS)

  if (isLoading) {
    return (
      <Card
        as="section"
        padding="sm"
        className={cn('flex flex-col gap-3', className)}
        {...props}
      >
        <Text as="h2" variant="heading" loading loadingWidth={18}>
          App config changes
        </Text>
        <Text as="p" variant="body" loading loadingWidth={28}>
          Loading configuration changes
        </Text>
      </Card>
    )
  }

  if (!group.sections.length) {
    return (
      <Card
        as="section"
        padding="sm"
        className={cn('flex flex-col gap-3', className)}
        {...props}
      >
        <Text as="h2" variant="heading">
          App config changes
        </Text>
        <div className="rounded-lg bg-surface-02 px-4 py-8 text-center">
          <Text as="p" variant="body" weight="medium">
            No config changes
          </Text>
          <Text as="p" variant="caption" color="tertiary">
            This config matches the previous version.
          </Text>
        </div>
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
          operations={APP_CONFIG_DIFF_OPERATIONS}
        />
      </header>

      <DiffSections
        defaultOpen={defaultSectionsOpen}
        toolbar={
          <DiffFilter
            title="config changes"
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
          [
            ...new Set(filter.filteredSections.map(({ group }) => group)),
          ].flatMap((sectionGroup) => [
            <Text
              key={`group-${sectionGroup}`}
              as="h3"
              variant="body"
              weight="semibold"
              className="px-1 pt-4 pb-1.5 first:pt-0"
            >
              {sectionGroup}
            </Text>,
            ...filter.filteredSections
              .filter(({ group }) => group === sectionGroup)
              .map((section) => (
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
              )),
          ])
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
