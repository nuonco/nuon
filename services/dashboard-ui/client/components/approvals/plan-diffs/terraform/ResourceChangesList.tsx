import React from 'react'
import { Badge } from '@/components/common/Badge'
import { EmptyState } from '@/components/common/EmptyState'
import { Expand } from '@/components/common/Expand'
import { Text } from '@/components/common/Text'
import type { TTerraformResourceChange } from '@/types'
import { TerraformValuesDiff } from './TerraformValuesDiff'
import {
  TERRAFORM_ACTION_BADGE_THEME,
  getTerraformActionBgColor,
  getTerraformActionBorderColor,
} from '../diff-style-utils'

interface ResourceChangesListProps {
  changes: TTerraformResourceChange[]
}

export function ResourceChangesList({ changes }: ResourceChangesListProps) {
  return (
    <div className="flex flex-col divide-y">
      {changes.length ? (
        changes.map((change, idx) => {
          const bgColor = getTerraformActionBgColor(change.action)
          const borderColor = getTerraformActionBorderColor(change.action)

          return (
            <Expand
              key={`${change.address}-${idx}`}
              id={change.address}
              className={`border-l-4 ${borderColor}`}
              headerClassName={`w-full px-4 py-3 gap-3 text-left focus:outline-none ${bgColor}`}
              heading={
                <div className="text-left w-full">
                  <div className="flex items-start justify-between w-full">
                    <div className="flex flex-col max-w-[500px]">
                      <Text
                        className="block !text-nowrap truncate"
                        weight="strong"
                      >
                        {change.address}
                      </Text>
                      <Text variant="subtext" theme="neutral">
                        {change.resource} ({change.name})
                      </Text>
                      {change.module && (
                        <Text variant="subtext" theme="neutral">
                          Module: {change.module}
                        </Text>
                      )}
                    </div>

                    <div className="flex items-center pr-4 self-center">
                      <Badge
                        theme={TERRAFORM_ACTION_BADGE_THEME[change.action]}
                        size="sm"
                      >
                        {change.action}
                      </Badge>
                    </div>
                  </div>
                </div>
              }
            >
              {change.action === 'read' ? (
                <div className="px-4 sm:px-6 py-2">
                  <Text>
                    Terraform will refresh this resource from the provider.
                  </Text>
                </div>
              ) : (
                <TerraformValuesDiff values={change} />
              )}
            </Expand>
          )
        })
      ) : (
        <div className="px-4 py-3 text-center">
          <EmptyState
            emptyMessage="Try clearing the search term or resetting the filter"
            emptyTitle="No changes to show"
            variant="search"
            size="sm"
          />
        </div>
      )}
    </div>
  )
}
