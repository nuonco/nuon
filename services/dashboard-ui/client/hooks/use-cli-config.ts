import { useQuery } from '@tanstack/react-query'
import { getCLIConfig } from '@/lib'

export const useCLIConfig = () =>
  useQuery({
    queryKey: ['cli-config'],
    queryFn: () => getCLIConfig(),
    staleTime: Infinity,
  })
