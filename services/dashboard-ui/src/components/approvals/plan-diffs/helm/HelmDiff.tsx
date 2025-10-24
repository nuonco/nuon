'use client'

import { useMemo } from 'react'
import { Badge } from '@/components/common/Badge'
import { Card } from '@/components/common/Card'
import { CodeBlock } from '@/components/common/CodeBlock'
import { EmptyState } from '@/components/common/EmptyState'
import { Expand } from '@/components/common/Expand'
import { Text } from '@/components/common/Text'
import { useHelmK8sPlanFilter } from '@/hooks/use-helm-k8s-plan-filter'
import type { THelmPlan, THelmPlanChange } from '@/types'
import { diffLines } from '@/utils/code-utils'
import { parseHelmPlan } from '@/utils/helm-utils'
import {
  HELM_ACTION_BADGE_THEME,
  getHelmActionBgColor,
  getHelmActionBorderColor,
} from '../diff-style-utils'
import { HelmDiffSummary } from './HelmDiffSummary'
import { DiffFilter } from '../DiffFilter'

export const HelmDiff = ({ plan }: { plan: THelmPlan }) => {
  const { changes, summary } = useMemo(() => parseHelmPlan(plan), [plan])

  const {
    selectedActions,
    searchQuery,
    filteredItems: filteredChanges,
    handleInputToggle,
    handleButtonClick,
    handleReset,
    handleSearchChange,
    filterStats,
  } = useHelmK8sPlanFilter<THelmPlanChange>(changes)

  return (
    <Card className="bg-cool-grey-50 dark:bg-dark-grey-900 !p-0 !gap-0">
      <div className="flex flex-col px-4 py-4 sm:px-6 border-b">
        <Text variant="base" weight="strong">
          Helm changes
        </Text>
        <Text variant="subtext" theme="neutral">
          Operation: {plan.op}
        </Text>
      </div>

      <HelmDiffSummary summary={summary} />

      <DiffFilter
        title="changes"
        diffType="helm-k8s"
        selectedActions={selectedActions}
        onInputToggle={handleInputToggle}
        onButtonClick={handleButtonClick}
        onReset={handleReset}
        selectedCount={filterStats.selectedCount}
        totalCount={filterStats.totalCount}
        searchValue={searchQuery}
        onSearchChange={handleSearchChange}
        searchPlaceholder="Search by release, resource, or type"
      />

      {filteredChanges && filteredChanges.length > 0 ? (
        <div className="divide-y">
          {filteredChanges.map((change, idx) => {
            const bgColor = getHelmActionBgColor(change.action)
            const borderColor = getHelmActionBorderColor(change.action)

            return (
              <Expand
                id={`change-${idx}`}
                key={`${change.release}-${idx}`}
                className={`border-l-4 ${borderColor}`}
                headerClassName={`!px-4 sm:!px-6 ${bgColor}`}
                heading={
                  <div className="text-left w-full">
                    <div className="flex items-start justify-between w-full">
                      <div className="flex flex-col">
                        <Text weight="strong">{change.release}</Text>
                        <Text variant="subtext" theme="neutral">
                          {change.resource} ({change.resourceType})
                        </Text>
                        <Text variant="subtext" theme="neutral">
                          Workspace: {change.workspace}
                        </Text>
                      </div>

                      <div className="flex items-center pr-4 self-center">
                        <Badge
                          size="sm"
                          theme={HELM_ACTION_BADGE_THEME[change.action]}
                        >
                          {change.action}
                        </Badge>
                      </div>
                    </div>
                  </div>
                }
              >
                <CodeBlock
                  className="!rounded-none border-t"
                  language="yaml"
                  isDiff
                >
                  {diffLines(change.before, change.after)}
                </CodeBlock>
              </Expand>
            )
          })}
        </div>
      ) : filteredChanges && filteredChanges.length === 0 ? (
        <div className="px-4 py-3 text-center">
          <EmptyState
            emptyMessage="Try clearing the search term or resetting the filter"
            emptyTitle="No changes to show"
            variant="search"
            size="sm"
          />
        </div>
      ) : null}
    </Card>
  )
}
