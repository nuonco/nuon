'use client'

import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { useQuery } from '@/hooks/use-query'
import { useOrg } from '@/hooks/use-org'
import { RunnerCard } from './RunnerCard'
import type { TRunner } from '@/types'

interface LoadRunnerCardProps {
  runnerId: string
  installId: string
}

export const LoadRunnerCard = ({ runnerId, installId }: LoadRunnerCardProps) => {
  const { org } = useOrg()
  const orgId = org.id

  const { data: runner, error: queryError, isLoading } = useQuery<TRunner>({
    path: `/api/orgs/${orgId}/runners/${runnerId}`,
    dependencies: [runnerId]
  })

  const error = queryError ? 'Unable to load runner' : null

  const refetch = () => {
    // The useQuery hook will handle refetching when dependencies change
    // For manual refetch, we could expose a refetch function from useQuery
  }

  if (error) {
    return (
      <Text variant="subtext" className="text-red-600">
        {error}
      </Text>
    )
  }

  if (isLoading) {
    return (
      <div className="flex items-center gap-2 p-4">
        <Icon variant="Loading" className="animate-spin" size="16" />
        <Text variant="subtext">Loading {runnerId} runner...</Text>
      </div>
    )
  }

  if (!runner) {
    return <Text variant="subtext">Runner not found</Text>
  }

  return (
    <RunnerCard
      runner={runner}
      href={`/${orgId}/installs/${installId}/runner`}
      isInstallRunner={true}
      onAction={refetch}
    />
  )
}