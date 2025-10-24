'use client'

import { useMemo } from 'react'
import { Card } from '@/components/common/Card'
import { Text } from '@/components/common/Text'
import type {
  TTerraformPlan,
  TTerraformOutputChange,
  TTerraformResourceChange,
} from '@/types'
import { parseTerraformPlan } from '@/utils/terraform-utils'
import { useTerraformResourceFilter } from '@/hooks/use-terraform-plan-resource-filter'
import { useTerraformOutputFilter } from '@/hooks/use-terraform-plan-output-filter'
import { DiffFilter } from '../DiffFilter'
import { TerraformSummary } from './TerraformSummary'
import { ResourceChangesList } from './ResourceChangesList'
import { OutputChangesList } from './OutputChangesList'

const EMPTY_PARSED_PLAN = {
  summary: {
    create: 0,
    'create-before-destroy': 0,
    'destroy-before-create': 0,
    delete: 0,
    replace: 0,
    read: 0,
    update: 0,
    noop: 0,
  },
  changes: [],
}

export function TerraformDiff({ plan }: { plan: TTerraformPlan | undefined }) {
  const { resources, outputs } = useMemo(() => {
    if (!plan) {
      return {
        resources: EMPTY_PARSED_PLAN,
        outputs: EMPTY_PARSED_PLAN,
      }
    }
    return parseTerraformPlan(plan)
  }, [plan])

  const resourceFilter = useTerraformResourceFilter<TTerraformResourceChange>(
    resources.changes
  )
  const outputFilter = useTerraformOutputFilter<TTerraformOutputChange>(
    outputs.changes
  )

  // Show loading/empty state if plan is undefined
  if (!plan) {
    return (
      <div className="flex flex-col gap-6">
        <Card className="bg-cool-grey-50 dark:bg-dark-grey-900 !p-0 !gap-0">
          <div className="px-4 sm:px-6 py-4 text-center">
            <Text variant="base" theme="neutral">
              No Terraform plan data available
            </Text>
          </div>
        </Card>
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-6">
      <Card className="bg-cool-grey-50 dark:bg-dark-grey-900 !p-0 !gap-0">
        <div className="px-4 sm:px-6 py-4 border-b">
          <Text variant="base" weight="strong">
            Resource changes
          </Text>
        </div>

        <TerraformSummary summary={resources.summary} />

        <DiffFilter
          title="resources"
          selectedActions={resourceFilter.selectedActions}
          onInputToggle={resourceFilter.handleInputToggle}
          onButtonClick={resourceFilter.handleButtonClick}
          onReset={resourceFilter.handleReset}
          selectedCount={resourceFilter.filterStats.selectedCount}
          totalCount={resourceFilter.filterStats.totalCount}
          searchValue={resourceFilter.searchQuery}
          onSearchChange={resourceFilter.handleSearchChange}
          searchPlaceholder="Search by address, resource, or name"
        />

        <ResourceChangesList changes={resourceFilter.filteredItems} />
      </Card>

      <Card className="bg-cool-grey-50 dark:bg-dark-grey-900 !p-0 !gap-0">
        <div className="px-4 sm:px-6 py-4 border-b">
          <Text variant="base" weight="strong">
            Output changes
          </Text>
        </div>

        <DiffFilter
          title="outputs"
          selectedActions={outputFilter.selectedActions}
          onInputToggle={outputFilter.handleInputToggle}
          onButtonClick={outputFilter.handleButtonClick}
          onReset={outputFilter.handleReset}
          selectedCount={outputFilter.filterStats.selectedCount}
          totalCount={outputFilter.filterStats.totalCount}
          searchValue={outputFilter.searchQuery}
          onSearchChange={outputFilter.handleSearchChange}
          searchPlaceholder="Search outputs by name"
        />

        <TerraformSummary summary={outputs.summary} />
        <OutputChangesList changes={outputFilter.filteredItems} />
      </Card>
    </div>
  )
}
