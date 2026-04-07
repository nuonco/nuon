import { useEffect, useState } from 'react'
import type { TRunner } from '@/types'
import { useAuth } from '@/hooks/use-auth'
import { adminGetInstallRunner } from '@/lib'
import { AdminInstallSection } from './AdminInstallSection'

interface IAdminInstallSectionContainer {
  orgId: string
  installId: string
}

export const AdminInstallSectionContainer = ({ orgId, installId }: IAdminInstallSectionContainer) => {
  const { user } = useAuth()
  const adminEmail = user?.email ?? ''
  const [runner, setRunner] = useState<TRunner>()
  const [runnerLoading, setRunnerLoading] = useState(true)

  useEffect(() => {
    if (installId) {
      adminGetInstallRunner({ installId }).then((r) => {
        setRunner(r)
        setRunnerLoading(false)
      })
    }
  }, [installId])

  return (
    <AdminInstallSection
      orgId={orgId}
      installId={installId}
      adminEmail={adminEmail}
      runner={runner}
      runnerLoading={runnerLoading}
    />
  )
}
