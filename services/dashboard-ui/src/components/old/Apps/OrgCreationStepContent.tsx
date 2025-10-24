'use client'

import { SpinnerIcon } from '@phosphor-icons/react'
import { Button } from '@/components/old/Button'
import { Text } from '@/components/old/Typography'
import { useAutoOrgCreation } from '@/hooks/use-auto-org-creation'
import { useQuery } from '@/hooks/use-query'
import { TOrg } from '@/types/ctl-api.types'

export const OrgCreationStepContent = ({
  stepComplete,
  onOrgCreated,
  orgId,
}: {
  stepComplete: boolean
  onOrgCreated?: (orgId: string) => void
  orgId: string | undefined
}) => {
  const { isCreating, error, retry, shouldAutoCreate } = useAutoOrgCreation()
  // Load org data
  const {
    data: org,
    isLoading: orgLoading,
    error: orgError,
  } = useQuery<TOrg>({
    path: `/api/orgs/${orgId}`,
    enabled: !!orgId,
  })

  if (orgId) {
    if (orgLoading) {
      return (
        <div className="space-y-3">
          <div className="flex items-center gap-2">
            <SpinnerIcon className="animate-spin" size={16} />
            <Text>
              Your trial organization (${orgId}) was created successfully.
            </Text>
            <Text>Fetching it from the API...</Text>
          </div>
        </div>
      )
    }

    if (orgError) {
      return (
        <div className="space-y-3">
          <div className="flex items-center gap-2">
            <Text>
              Your trial organization (${orgId}) was created successfully, but
              there was an error fetching it from the API.
            </Text>
            <Text>{orgError.error}</Text>
          </div>
        </div>
      )
    }

    return (
      <div className="space-y-6">
        {/* Success Message - Shown when step is complete */}
        {stepComplete && (
          <div className="space-y-3 pb-4 border-b border-gray-200 dark:border-gray-700">
            <div className="flex items-center gap-2">
              <div className="w-2 h-2 bg-green-500 rounded-full" />
              <Text
                variant="semi-14"
                className="text-green-800 dark:text-green-200"
              >
                Your trial organization has been created successfully!
              </Text>
            </div>
            <Text>Name: {org.name}</Text>
            <Text>ID: {org.id}</Text>
            <Text>Status: {org.status}</Text>
            <Text>
              It may take a few minutes to fully provision, but you don&apos;t
              have to wait for it. You can continue to the next step while it
              finishes.
            </Text>
          </div>
        )}
      </div>
    )
  }

  // Early return for non-success states - these don't need the prepend pattern
  // since they're temporary states during the creation process
  if (isCreating) {
    return (
      <div className="space-y-3">
        <div className="flex items-center gap-2">
          <SpinnerIcon className="animate-spin" size={16} />
          <Text className="text-blue-600 dark:text-blue-400">
            Setting up your trial organization...
          </Text>
        </div>
      </div>
    )
  }

  if (error) {
    if (error) {
      return (
        <div className="space-y-3">
          <div className="flex items-center gap-2 text-red-600 dark:text-red-400">
            <Text>Org creation failed</Text>
          </div>
          <Text className="text-sm text-gray-600 dark:text-gray-400 mb-4">
            {error}
          </Text>
          <Button onClick={retry} variant="primary">
            Try Again
          </Button>
        </div>
      )
    }
  }
}
