import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Panel, type IPanel } from '@/components/surfaces/Panel'
import { CheckboxInput } from '@/components/common/form/CheckboxInput'
import { Skeleton } from '@/components/common/Skeleton'
import type { TOrg } from '@/types'

interface IAdminOrgFeaturesPanel extends Omit<IPanel, 'onSubmit'> {
  org: TOrg
  orgId: string
  featuresList: string[]
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
  return (
    <Panel
      heading={
        <div className="flex items-center gap-3">
          <Icon variant="Sliders" size="24" />
          <Text weight="strong" variant="h2">Organization features</Text>
        </div>
      }
      size={size}
      {...props}
    >
      <div className="flex flex-col gap-6">
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
            <div className="flex justify-end pt-4 border-t">
              <Skeleton className="h-10 w-32" />
            </div>
          </div>
        ) : featuresList.length > 0 ? (
          <form id="features-form" onSubmit={onSubmit}>
            <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
              {featuresList.map((feature) => (
                <CheckboxInput
                  key={feature}
                  name={feature}
                  defaultChecked={org?.features?.[feature] || false}
                  labelProps={{
                    labelText: feature,
                  }}
                />
              ))}
            </div>
            <div className="flex justify-end gap-3 mt-6 pt-4 border-t">
              <Button
                type="submit"
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
            </div>
          </form>
        ) : (
          <div className="flex flex-col items-center justify-center py-12 text-center">
            <Icon variant="Warning" size="48" className="text-gray-400 mb-4" />
            <Text variant="base" weight="strong" className="mb-2">No features available</Text>
            <Text variant="subtext">No feature flags are configured for this organization.</Text>
          </div>
        )}
      </div>
    </Panel>
  )
}
