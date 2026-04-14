import { Banner } from '@/components/common/Banner'
import { Card } from '@/components/common/Card'
import { Text } from '@/components/common/Text'
import { usePulumiPlanFilter } from '@/hooks/use-pulumi-plan-filter'
import { DiffFilter } from '../DiffFilter'
import { PulumiSummary } from './PulumiSummary'
import { PulumiResourceChangesList } from './PulumiResourceChangesList'

interface IPropertyDiff {
  kind: string
  inputDiff: boolean
}

interface IResourceChange {
  urn: string
  type: string
  name: string
  action: string
  diffs?: string[]
  detailed_diff?: Record<string, IPropertyDiff>
  old_inputs?: Record<string, unknown>
  new_inputs?: Record<string, unknown>
  provider?: string
}

interface IPulumiPreviewResult {
  stdout: string
  stderr: string
  change_summary: Record<string, number>
  resource_changes?: IResourceChange[]
  diagnostics?: string[]
}

export function PulumiDiff({
  plan,
}: {
  plan: IPulumiPreviewResult | undefined
}) {
  const resourceFilter = usePulumiPlanFilter(plan?.resource_changes || [])

  if (!plan) {
    return <Banner theme="neutral">No Pulumi preview data available</Banner>
  }

  const hasResourceChanges =
    plan.resource_changes && plan.resource_changes.length > 0

  return (
    <div className="flex flex-col gap-6">
      <Card className="bg-cool-grey-50 dark:bg-dark-grey-900 !p-0 !gap-0">
        <div className="px-4 sm:px-6 py-4 border-b">
          <Text variant="base" weight="strong">
            Pulumi preview
          </Text>
        </div>

        {plan.change_summary && (
          <PulumiSummary changeSummary={plan.change_summary} />
        )}

        {hasResourceChanges ? (
          <>
            <DiffFilter
              title="resources"
              diffType="pulumi"
              selectedActions={resourceFilter.selectedActions}
              onInputToggle={resourceFilter.handleInputToggle}
              onButtonClick={resourceFilter.handleButtonClick}
              onReset={resourceFilter.handleReset}
              selectedCount={resourceFilter.filterStats.selectedCount}
              totalCount={resourceFilter.filterStats.totalCount}
              searchValue={resourceFilter.searchQuery}
              onSearchChange={resourceFilter.handleSearchChange}
              searchPlaceholder="Search by type, name, or URN"
            />

            <PulumiResourceChangesList
              changes={resourceFilter.filteredItems}
            />
          </>
        ) : (
          plan.stdout && (
            <div className="px-4 sm:px-6 py-4">
              <pre className="text-xs font-mono whitespace-pre-wrap overflow-x-auto text-grey-700 dark:text-grey-300">
                {plan.stdout}
              </pre>
            </div>
          )
        )}

        {plan.diagnostics && plan.diagnostics.length > 0 && (
          <div className="px-4 sm:px-6 py-4 border-t flex flex-col gap-3">
            <Text variant="subtext" weight="strong">
              Diagnostics
            </Text>
            {plan.diagnostics.map((d, i) => {
              const theme = d.startsWith('error') ? 'error' as const : 'warn' as const
              return (
                <Banner key={i} theme={theme} className="!rounded-md">
                  <Text variant="subtext" family="mono" className="whitespace-pre-wrap">
                    {d}
                  </Text>
                </Banner>
              )
            })}
          </div>
        )}
      </Card>
    </div>
  )
}
