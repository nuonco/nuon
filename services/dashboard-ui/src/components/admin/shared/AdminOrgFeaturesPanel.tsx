'use client'

import { useState, useEffect } from 'react'
import { usePathname } from 'next/navigation'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Panel, type IPanel } from '@/components/surfaces/Panel'
import { CheckboxInput } from '@/components/common/form/CheckboxInput'
import { Skeleton } from '@/components/common/Skeleton'
import { useServerAction } from '@/hooks/use-server-action'
import { useServerActionToast } from '@/hooks/use-server-action-toast'
import { updateOrgFeature } from '@/actions/admin/update-org-feature'
import type { TOrg } from '@/types'

interface AdminOrgFeaturesPanelProps extends IPanel {
  org: TOrg
  orgId: string
}

export const AdminOrgFeaturesPanel = ({ 
  org, 
  orgId,
  size = 'half',
  ...props
}: AdminOrgFeaturesPanelProps) => {
  const pathname = usePathname()
  const [featuresList, setFeaturesList] = useState<string[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string>()

  const { execute, isLoading: isSubmitting, data, error: submitError } = useServerAction({
    action: async (formData: FormData) => {
      const result = await updateOrgFeature(orgId, formData, featuresList, pathname)
      return { data: result, error: null, headers: {}, status: 200 }
    }
  })

  useServerActionToast({
    data,
    error: submitError,
    successContent: <Text>Organization features updated successfully</Text>,
    successHeading: 'Features Updated',
    errorContent: <Text>Failed to update organization features. Please try again.</Text>,
    errorHeading: 'Update Failed'
  })

  useEffect(() => {
    setIsLoading(true)
    setError(undefined)
    
    fetch(`/api/orgs/${orgId}/features`)
      .then((res) => res.json())
      .then((features) => {
        setIsLoading(false)
        if (Array.isArray(features)) {
          setFeaturesList(features)
        } else {
          setError('Invalid features data received')
        }
      })
      .catch((err) => {
        setIsLoading(false)
        setError('Unable to fetch org features list')
      })
  }, [orgId])

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    const formData = new FormData(e.currentTarget)
    await execute(formData)
  }


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
          <form id="features-form" onSubmit={handleSubmit}>
            <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
              {featuresList.map((feature) => (
                <CheckboxInput
                  key={feature}
                  name={feature}
                  defaultChecked={org?.features?.[feature] || false}
                  labelProps={{
                    labelText: feature
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
