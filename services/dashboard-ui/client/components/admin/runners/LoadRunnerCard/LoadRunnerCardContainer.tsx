import { useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { getRunner } from '@/lib'
import type { TRunner } from '@/types'
import { LoadRunnerCard } from './LoadRunnerCard'

interface LoadRunnerCardContainerProps {
  runnerId: string
  installId: string
}

export const LoadRunnerCardContainer = ({ runnerId, installId }: LoadRunnerCardContainerProps) => {
  const { org } = useOrg()
  const orgId = org.id

  const { data: runner, error: queryError, isLoading, refetch } = useQuery<TRunner>({
    queryKey: ['runner', orgId, runnerId],
    queryFn: () => getRunner({ orgId, runnerId }),
    enabled: !!orgId && !!runnerId,
  })

  return (
    <LoadRunnerCard
      runner={runner}
      error={queryError ? 'Unable to load runner' : null}
      isLoading={isLoading}
      href={`/${orgId}/installs/${installId}/runner`}
      onAction={() => refetch()}
    />
  )
}
