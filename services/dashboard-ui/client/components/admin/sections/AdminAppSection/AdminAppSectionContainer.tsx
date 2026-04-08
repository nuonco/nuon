import { useAuth } from '@/hooks/use-auth'
import { AdminAppSection } from './AdminAppSection'

interface IAdminAppSectionContainer {
  orgId: string
  appId: string
}

export const AdminAppSectionContainer = ({ orgId, appId }: IAdminAppSectionContainer) => {
  const { user } = useAuth()
  const adminEmail = user?.email ?? ''

  return <AdminAppSection orgId={orgId} appId={appId} adminEmail={adminEmail} />
}
