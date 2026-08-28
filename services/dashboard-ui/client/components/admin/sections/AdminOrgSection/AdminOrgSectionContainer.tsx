import { useEffect, useState } from 'react'
import type { TRunner } from '@/types'
import { useOrg } from '@/hooks/use-org'
import { useAuth } from '@/hooks/use-auth'
import { useConfig } from '@/hooks/use-config'
import { adminGetOrgRunner } from '@/lib'
import { AdminOrgSection } from './AdminOrgSection'

interface IAdminOrgSectionContainer {
  orgId: string
}

export const AdminOrgSectionContainer = ({
  orgId,
}: IAdminOrgSectionContainer) => {
  const { org } = useOrg()
  const { user } = useAuth()
  const config = useConfig()
  const adminEmail = user?.email ?? ''
  const [runner, setRunner] = useState<TRunner>()

  useEffect(() => {
    if (orgId) {
      adminGetOrgRunner({ orgId }).then(setRunner)
    }
  }, [orgId])

  return (
    <AdminOrgSection
      orgId={orgId}
      org={org}
      adminEmail={adminEmail}
      adminDashboardUrl={config.adminDashboardUrl}
      runner={runner}
    />
  )
}
