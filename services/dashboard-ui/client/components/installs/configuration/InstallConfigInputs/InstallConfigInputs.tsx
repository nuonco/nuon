import type { ReactNode } from 'react'
import { Card } from '@/components/common/Card'
import { EmptyState } from '@/components/common/EmptyState'
import { PropertyGrid } from '@/components/common/PropertyGrid'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { InputValue } from '@/components/installs/management/InputValue'
import type { TConfigurationInputGroup } from '@/components/installs/configuration/shared/use-configuration-inputs'
import { getInputDisplayName } from '@/utils/install-utils'

export interface IInstallConfigInputs {
  action?: ReactNode
  groups: TConfigurationInputGroup[]
  isLoading?: boolean
  values?: Record<string, string>
}

export const InstallConfigInputs = ({
  action,
  groups,
  isLoading,
  values = {},
}: IInstallConfigInputs) => {
  if (isLoading) return <Skeleton height="180px" width="100%" />

  if (!groups.length) {
    return (
      <EmptyState
        variant="table"
        emptyTitle="No inputs configured"
        emptyMessage="This app config does not define any install inputs."
      />
    )
  }

  return (
    <div className="flex flex-col gap-4">
      {action ? <div className="flex justify-end">{action}</div> : null}
      {groups.map((group) => (
        <Card key={group.id ?? group.name} className="!p-4 !gap-4">
          <div className="flex flex-col gap-0.5">
            <Text variant="body" weight="strong">
              {group.display_name ?? group.name}
            </Text>
            {group.description ? (
              <Text variant="subtext" theme="neutral">
                {group.description}
              </Text>
            ) : null}
          </div>
          <PropertyGrid
            align="start"
            columns={[
              { key: 'input', header: 'Input' },
              { key: 'value', header: 'Current value' },
              { key: 'defaultValue', header: 'Default' },
            ]}
            gridTemplate="minmax(150px, 1fr) minmax(150px, 2fr) minmax(120px, 1fr)"
            values={(group.app_inputs ?? []).map((input) => ({
              input: (
                <span className="flex flex-col">
                  <Text variant="subtext" weight="strong">
                    {input.display_name ?? input.name}
                  </Text>
                  <Text variant="label" family="mono" theme="neutral">
                    {input.name ? getInputDisplayName(input.name) : null}
                  </Text>
                </span>
              ),
              value: (
                <InputValue
                  name={input.name}
                  value={input.name ? values[input.name] : undefined}
                />
              ),
              defaultValue: (
                <Text variant="label" family="mono" theme="neutral">
                  {input.default ?? '—'}
                </Text>
              ),
            }))}
          />
        </Card>
      ))}
    </div>
  )
}
