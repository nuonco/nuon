'use client'

import { useParams } from 'next/navigation'
import { useUser } from '@auth0/nextjs-auth0'
import { AdminOrgSection } from './sections/AdminOrgSection'
import { AdminAppSection } from './sections/AdminAppSection'
import { AdminInstallSection } from './sections/AdminInstallSection'

export const AdminControls = () => {
  const { user } = useUser()
  const params = useParams()
  
  // Only show to Nuon employees
  if (!user?.email?.endsWith('@nuon.co')) {
    return null
  }

  const orgId = params?.['org-id'] as string
  const appId = params?.['app-id'] as string
  const installId = params?.['install-id'] as string

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
