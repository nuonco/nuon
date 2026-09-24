import { isNuonSession } from '@/utils/session-utils'
import { useCurrentUser } from './use-current-user'

export const useNuonStaff = () => {
  const { user, isLoading } = useCurrentUser()

  return {
    staff: isNuonSession({ sub: '', email: user.email ?? '' }),
    loading: isLoading,
  }
}
