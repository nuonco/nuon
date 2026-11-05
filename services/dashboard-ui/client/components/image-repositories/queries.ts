import { useQuery } from '@tanstack/react-query'
import { getImageRepository } from '@/lib'

export const useImageRepository = (orgId?: string, repositoryId?: string) =>
  useQuery({
    queryKey: ['image-repositories', orgId, repositoryId],
    queryFn: () =>
      getImageRepository({ orgId: orgId!, repositoryId: repositoryId! }),
    enabled: !!orgId && !!repositoryId,
  })
