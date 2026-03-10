import { useState, useEffect } from 'react'
import { useMutation } from '@tanstack/react-query'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { Panel, type IPanel } from '@/components/surfaces/Panel'
import { CheckboxInput } from '@/components/common/form/CheckboxInput'
import { Skeleton } from '@/components/common/Skeleton'
import { useToast } from '@/hooks/use-toast'
import { useAuth } from '@/hooks/use-auth'
import { useConfig } from '@/hooks/use-config'
import { adminGetOrgFeaturesList, adminUpdateOrgFeatures } from '@/lib'
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
  const { addToast } = useToast()
  const { user } = useAuth()
  const config = useConfig()
  const adminEmail = user?.email ?? ''
  const adminApiUrl = config.adminApiUrl ?? ''
  const [featuresList, setFeaturesList] = useState<string[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string>()

  const { mutate: submit, isPending: isSubmitting } = useMutation({
    mutationFn: (formData: FormData) => {
      const features: Record<string, boolean> = {}
      featuresList.forEach((feature) => {
        features[feature] = formData.get(feature) === 'on'
      })
      return adminUpdateOrgFeatures({ orgId, features, adminApiUrl, adminEmail })
    },
    onSuccess: () => {
      addToast(
        <Toast heading="Features Updated" theme="success">
          <Text>Organization features updated successfully</Text>
        </Toast>
      )
    },
    onError: () => {
      addToast(
        <Toast heading="Update Failed" theme="error">
          <Text>Failed to update organization features. Please try again.</Text>
        </Toast>
      )
    },
  })

  useEffect(() => {
    setIsLoading(true)
    setError(undefined)

    adminGetOrgFeaturesList({ adminApiUrl })
      .then((features) => {
        setIsLoading(false)
        if (Array.isArray(features)) {
          setFeaturesList(features)
        } else {
          setError('Invalid features data received')
        }
      })
      .catch(() => {
        setIsLoading(false)
        setError('Unable to fetch org features list')
      })
  }, [adminApiUrl])

  const handleSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    const formData = new FormData(e.currentTarget)
    submit(formData)
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
