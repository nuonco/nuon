import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { getCLIConfig } from '@/lib'

export const useCLIConfig = () =>
  useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['cli-config'],
    queryFn: () => getCLIConfig(),
    staleTime: Infinity,
  })
