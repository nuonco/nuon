import { useState, useEffect } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { type IPanel } from '@/components/surfaces/Panel'
import { useToast } from '@/hooks/use-toast'
import { useAuth } from '@/hooks/use-auth'
import { adminGetOrgFeaturesList, adminUpdateOrgFeatures } from '@/lib'
import type { TOrg } from '@/types'
import { AdminOrgFeaturesPanel } from './AdminOrgFeaturesPanel'

interface AdminOrgFeaturesPanelContainerProps extends IPanel {
  org: TOrg
  orgId: string
}

export const AdminOrgFeaturesPanelContainer = ({
  org,
  orgId,
  onSubmit: _onSubmit,
  ...props
}: AdminOrgFeaturesPanelContainerProps) => {
  const { addToast } = useToast()
  const queryClient = useQueryClient()
  const { user } = useAuth()
  const adminEmail = user?.email ?? ''
  const [featuresList, setFeaturesList] = useState<string[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string>()

  const { mutate: submit, isPending: isSubmitting } = useMutation({
    mutationFn: (formData: FormData) => {
      const features: Record<string, boolean> = {}
      featuresList.forEach((feature) => {
        features[feature] = formData.get(feature) === 'on'
      })
      return adminUpdateOrgFeatures({ orgId, features, adminEmail })
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['org'] })
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
    adminGetOrgFeaturesList()
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
  }, [])

  const handleSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    const formData = new FormData(e.currentTarget)
    submit(formData)
  }

  return (
    <AdminOrgFeaturesPanel
      org={org}
      orgId={orgId}
      featuresList={featuresList}
      isLoading={isLoading}
      isSubmitting={isSubmitting}
      error={error}
      onSubmit={handleSubmit}
      {...props}
    />
  )
}
