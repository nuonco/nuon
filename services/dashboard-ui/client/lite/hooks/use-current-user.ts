import { useQuery } from '@tanstack/react-query'
import { getMe } from '@/lib'

export const useCurrentUser = () => {
  const query = useQuery({
    queryKey: ['auth', 'me'],
    queryFn: getMe,
    staleTime: Infinity,
    retry: false,
  })
  const identity = query.data?.identities?.[0]

  return {
    ...query,
    user: {
      name: identity?.name,
      email: query.data?.email,
      picture: identity?.picture,
    },
  }
}
