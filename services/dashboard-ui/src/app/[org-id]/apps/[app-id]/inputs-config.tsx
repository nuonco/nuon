import { EmptyState } from '@/components/common/EmptyState/EmptyState'
import { Expand } from '@/components/common/Expand'
import { Icon } from '@/components/common/Icon'
import {
  PropertyGrid,
  PropertyGridSkeleton,
} from '@/components/common/PropertyGrid'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { getAppConfig } from '@/lib'
import type { TAppConfig } from '@/types'

export async function AppInputs({
  appConfigId,
  appId,
  orgId,
}: {
  appConfigId?: string
  appId: string
  orgId: string
}) {
  if (!appConfigId) {
    return <AppInputsError />
  }

  const { data: config, error } = await getAppConfig({
    appConfigId,
    appId,
    orgId,
    recurse: true,
  })

  return !error && config?.input && config?.input?.input_groups?.length ? (
    <div className="flex flex-col gap-4">
      <Text variant="h3" weight="strong">
        Inputs config
      </Text>

      {normalizeInputGroups(config.input.input_groups, config.input.inputs).map(
        (inputGroup) => (
          <Expand
            isOpen
            id={inputGroup?.id}
            key={inputGroup.id}
            heading={
              <div className="flex flex-col items-start">
                <Text weight="strong">{inputGroup?.display_name}</Text>
                <Text variant="subtext" theme="neutral">
                  {inputGroup?.description}
                </Text>
              </div>
            }
            className="border rounded-md"
            headerClassName="!px-4"
          >
            <div className="p-4 border-t bg-code">
              <PropertyGrid
                columns={[
                  { key: 'name', header: 'Name' },
                  { key: 'description', header: 'Description' },
                  { key: 'default', header: 'Default' },
                  { key: 'required', header: 'Required' },
                  { key: 'sensitive', header: 'Sensitive' },
                  { key: 'source', header: 'Source' },
                ]}
                gridTemplate="minmax(150px, 2fr) minmax(200px, 3.5fr) minmax(120px, 1.5fr) minmax(80px, max-content) minmax(80px, max-content) minmax(80px, max-content)"
                values={inputGroup?.app_inputs?.map((input) => ({
                  name: (
                    <span className="flex flex-col">
                      <Text variant="subtext" weight="strong">
                        {input.display_name}
                      </Text>
                      <Text variant="label" family="mono" theme="neutral">
                        {input.name}
                      </Text>
                    </span>
                  ),
                  description: (
                    <Text variant="subtext">{input?.description}</Text>
                  ),
                  default: (
                    <Text variant="label" family="mono" theme="neutral">
                      {input?.default}
                    </Text>
                  ),
                  required: (
                    <Icon
                      variant={input?.required ? 'CheckIcon' : 'MinusIcon'}
                    />
                  ),
                  sensitive: (
                    <Icon
                      variant={input?.sensitive ? 'CheckIcon' : 'MinusIcon'}
                    />
                  ),
                  source: (
                    <Text
                      variant="label"
                      family="mono"
                      theme={input?.source === 'vendor' ? 'info' : 'brand'}
                    >
                      {input?.source}
                    </Text>
                  ),
                }))}
              />
            </div>
          </Expand>
        )
      )}
    </div>
  ) : (
    <AppInputsError />
  )
}

export const AppInputsError = () => (
  <EmptyState
    variant="diagram"
    emptyTitle="No app inputs configured"
    emptyMessage="Configure app inputs in your application configuration to see them here."
  />
)

export const AppInputsSkeleton = () => (
  <div className="flex flex-col gap-6">
    <div className="flex items-center gap-3">
      <Skeleton height="28px" width="120px" />
    </div>

    {/* Skeleton for input groups */}
    {Array.from({ length: 2 }).map((_, groupIndex) => (
      <div key={groupIndex} className="border rounded-md">
        <div className="px-4 py-3 border-b">
          <div className="flex flex-col gap-2">
            <Skeleton height="20px" width="180px" />
            <Skeleton height="16px" width="240px" />
          </div>
        </div>
        <div className="p-4 bg-code">
          <PropertyGridSkeleton count={3} columns={6} />
        </div>
      </div>
    ))}
  </div>
)

function normalizeInputGroups(
  groups: TAppConfig['input']['input_groups'],
  inputs: TAppConfig['input']['inputs']
) {
  return groups.map((group) => ({
    ...group,
    app_inputs: inputs.filter((input) => input.group_id === group.id),
  }))
}
