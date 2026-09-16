import { useMemo, useState } from 'react'
import { Button } from '@/components/common/Button'
import { EmptyState } from '@/components/common/EmptyState'
import { Icon } from '@/components/common/Icon'
import { SearchInput } from '@/components/common/SearchInput'
import { Text } from '@/components/common/Text'
import { Panel, type IPanel } from '@/components/surfaces/Panel'
import { CheckboxInput } from '@/components/common/form/CheckboxInput'
import { Skeleton } from '@/components/common/Skeleton'
import type { TOrgFeatureInfo } from '@/lib'
import type { TOrg } from '@/types'
import { cn } from '@/utils/classnames'

interface IAdminOrgFeaturesPanel extends Omit<IPanel, 'onSubmit'> {
  org: TOrg
  orgId: string
  featuresList: TOrgFeatureInfo[]
  isLoading: boolean
  isSubmitting: boolean
  error?: string
  onSubmit: (e: React.FormEvent<HTMLFormElement>) => void
}

export const AdminOrgFeaturesPanel = ({
  org,
  orgId,
  featuresList,
  isLoading,
  isSubmitting,
  error,
  onSubmit,
  size = 'half',
  ...props
}: IAdminOrgFeaturesPanel) => {
  const [query, setQuery] = useState('')
  const features = useMemo(
    () => [...featuresList].reverse(),
    [featuresList]
  )
  const match = query.trim().toLowerCase()
  const visibleCount = match
    ? features.filter((feature) => feature.name.toLowerCase().includes(match))
        .length
    : features.length

  return (
    <Panel
      heading={
        <div className="flex items-center gap-3">
          <Icon variant="SlidersIcon" size="24" />
          <Text weight="strong" variant="h2">Organization features</Text>
        </div>
      }
      size={size}
      footer={
        isLoading || featuresList.length > 0 ? (
          <Button
            type="submit"
            form="features-form"
            disabled={isSubmitting || isLoading}
            variant="primary"
          >
            {isSubmitting ? (
              <>
                <Icon variant="Loading" className="animate-spin" />
                Updating...
              </>
            ) : (
              'Update features'
            )}
          </Button>
        ) : null
      }
      {...props}
    >
      <div className="@container flex flex-col gap-6">
        <Text variant="body" className="text-gray-600 dark:text-gray-300">
          Configure feature flags for organization: <span className="font-mono">{orgId}</span>
        </Text>

        {error && (
          <div className="p-4 rounded-lg bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800">
            <Text variant="subtext" className="text-red-700 dark:text-red-300">
              {error}
            </Text>
          </div>
        )}

        {isLoading ? (
          <div className="space-y-6">
            <div className="grid grid-cols-2 md:grid-cols-3 gap-3">
              {Array.from({ length: 15 }).map((_, index) => (
                <div key={index} className="flex items-center gap-3">
                  <Skeleton className="w-4 h-4 rounded-sm" />
                  <Skeleton className="h-5 w-24 md:w-32" />
                </div>
              ))}
            </div>
          </div>
        ) : featuresList.length > 0 ? (
          <form id="features-form" onSubmit={onSubmit} className="flex flex-col gap-4">
            <SearchInput
              aria-label="Search feature flags"
              labelClassName="w-full max-w-full"
              className="w-full md:min-w-0"
              placeholder="Search by flag name"
              value={query}
              onChange={setQuery}
              onClear={() => setQuery('')}
            />
            {visibleCount === 0 ? (
              <EmptyState
                variant="table"
                emptyTitle="No flags match this search"
                emptyMessage="Clear the search to see all feature flags."
              />
            ) : null}
            <div
              className={cn(
                'grid grid-cols-1 @md:grid-cols-2 @3xl:grid-cols-3 @5xl:grid-cols-4 gap-4',
                visibleCount === 0 && 'hidden'
              )}
            >
              {features.map((feature) => (
                <div
                  key={feature.name}
                  className={
                    match && !feature.name.toLowerCase().includes(match)
                      ? 'hidden'
                      : undefined
                  }
                >
                <CheckboxInput
                  name={feature.name}
                  defaultChecked={feature.forced || org?.features?.[feature.name] || false}
                  disabled={feature.forced}
                  labelProps={{
                    className: cn('!items-start', feature.forced && 'cursor-not-allowed'),
                    labelText: (
                      <div className="flex flex-col gap-0.5">
                        <span className="flex items-center gap-2">
                          {feature.name}
                          {feature.forced && (
                            <span
                              className="rounded bg-amber-100 px-1.5 py-0.5 text-[10px] font-mono text-amber-700 dark:bg-amber-900/30 dark:text-amber-300"
                              title="Forced on for every org by this deployment's forced_enabled_features config"
                            >
                              forced
                            </span>
                          )}
                        </span>
                        {feature.description && (
                          <Text variant="subtext" className="text-gray-500 dark:text-gray-400">
                            {feature.description}
                          </Text>
                        )}
                      </div>
                    ),
                  }}
                />
                </div>
              ))}
            </div>
          </form>
        ) : (
          <div className="flex flex-col items-center justify-center py-12 text-center">
            <Icon variant="WarningIcon" size="48" className="text-gray-400 mb-4" />
            <Text variant="base" weight="strong" className="mb-2">No features available</Text>
            <Text variant="subtext">No feature flags are configured for this organization.</Text>
          </div>
        )}
      </div>
    </Panel>
  )
}
