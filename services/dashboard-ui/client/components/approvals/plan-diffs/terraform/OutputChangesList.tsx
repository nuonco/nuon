import React from 'react'
import { Badge } from '@/components/common/Badge'
import { EmptyState } from '@/components/common/EmptyState'
import { Expand } from '@/components/common/Expand'
import { Text } from '@/components/common/Text'
import { isOutputAfterUnknown } from '@/utils/terraform-utils'
import type { TTerraformOutputChange } from '@/types'
import { TerraformValuesDiff } from './TerraformValuesDiff'
import {
  TERRAFORM_ACTION_BADGE_THEME,
  getTerraformActionBgColor,
  getTerraformActionBorderColor,
} from '../diff-style-utils'

interface OutputChangesListProps {
  changes: TTerraformOutputChange[]
}

export function OutputChangesList({ changes }: OutputChangesListProps) {
  return (
    <div className="flex flex-col divide-y">
      {changes.length ? (
        changes.map((change, idx) => {
          const bgColor = getTerraformActionBgColor(change.action)
          const borderColor = getTerraformActionBorderColor(change.action)

          return (
            <Expand
              key={`${change.output}-${idx}`}
              id={change.output}
              className={`border-l-4 ${borderColor}`}
              headerClassName={`w-full px-4 py-3 gap-3 text-left focus:outline-none ${bgColor}`}
              heading={
                <div className="text-left w-full">
                  <div className="flex items-start justify-between w-full">
                    <div className="flex flex-col max-w-[550px]">
                      <Text
                        className="block !text-nowrap truncate"
                        weight="strong"
                      >
                        {change.output}
                      </Text>
                      <span className="flex items-center gap-6">
                        <Text variant="subtext" theme="neutral">
                          <b>Sensitive before</b>:{' '}
                          {change.beforeSensitive ? 'Yes' : 'No'}
                        </Text>
                        <Text variant="subtext" theme="neutral">
                          <b>Sensitive after</b>:{' '}
                          {change.afterSensitive ? 'Yes' : 'No'}
                        </Text>
                        <Text variant="subtext" theme="neutral">
                          <b>Status</b>:{' '}
                          {isOutputAfterUnknown(change.afterUnknown)
                            ? 'Some values known after apply'
                            : 'Value known at plan time'}
                        </Text>
                      </span>
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
              <TerraformValuesDiff values={change} />
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
