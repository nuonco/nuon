import { useAuth } from '@/hooks/use-auth'
import { UserProfile } from './UserProfile'

export const UserProfileContainer = () => {
  const { user, isLoading } = useAuth()
  return <UserProfile isLoading={isLoading} user={user} />
}
