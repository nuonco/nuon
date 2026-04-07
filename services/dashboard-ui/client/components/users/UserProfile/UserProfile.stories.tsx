export default {
  title: 'Users/UserProfile',
}

import { UserProfile } from './UserProfile'

export const Default = () => (
  <UserProfile
    isLoading={false}
    user={{ name: 'Jane Smith', email: 'jane@example.com', picture: '' }}
  />
)

export const Loading = () => (
  <UserProfile isLoading user={null} />
)
