import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { useSearchParams } from 'react-router'
import { useOrg } from '@/hooks/use-org'
import { getImageRepositories } from '@/lib'
import { ImageRepositories } from './ImageRepositories'

export const ImageRepositoriesContainer = () => {
  const { org } = useOrg()
  const [params] = useSearchParams()
  const offset = Number(params.get('offset') || 0)
  const q = params.get('q') || undefined
  const platform = params.get('platform') || undefined
  const query = useQuery({
    queryKey: ['image-repositories', org?.id, offset, q, platform],
    queryFn: () =>
      getImageRepositories({ orgId: org.id, offset, limit: 20, q, platform }),
    enabled: !!org?.id,
    placeholderData: keepPreviousData,
  })
  return (
    <ImageRepositories
      orgId={org?.id}
      repositories={query.data?.data ?? []}
      isLoading={query.isLoading}
      error={query.error}
      pagination={{
        hasNext: query.data?.pagination.hasNext ?? false,
        offset,
        limit: 20,
      }}
    />
  )
}
