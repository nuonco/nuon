import { useParams } from 'react-router'
import { useAuth } from '@/hooks/use-auth'
import { AdminOrgSection } from './sections/AdminOrgSection'
import { AdminAppSection } from './sections/AdminAppSection'
import { AdminInstallSection } from './sections/AdminInstallSection'

export const AdminControls = () => {
  const { user } = useAuth()
  const params = useParams()
  
  // Only show to Nuon employees
  if (!user?.email?.endsWith('@nuon.co')) {
    return null
  }

  const orgId = params?.orgId as string
  const appId = params?.appId as string
  const installId = params?.installId as string

  return (
    <div className="flex flex-col gap-8 h-full overflow-y-auto">
      {/* Always show org section */}
      <AdminOrgSection orgId={orgId} />
      
      {/* Conditionally show app section */}
      {appId && <AdminAppSection orgId={orgId} appId={appId} />}
      
      {/* Conditionally show install section */}
      {installId && <AdminInstallSection orgId={orgId} installId={installId} />}
    </div>
  )
}
